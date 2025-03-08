package main

import (
	"context"
	"io"
	defaultLog "log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/vladislavprovich/sso/internal/rabbitmq"
	"github.com/vladislavprovich/sso/internal/rabbitmq/publisher"

	"github.com/vladislavprovich/sso/internal/lib/logger/handlers/slogpretty"

	"github.com/vladislavprovich/sso/internal/lib/telemetry"

	"github.com/vladislavprovich/sso/internal/app"
	"github.com/vladislavprovich/sso/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	ctx := context.Background()
	log := setupLogger(cfg)

	// Init logs directory.
	err := telemetry.EnsureLogDir(cfg.Logging.LogDir)
	if err != nil {
		log.Error("failed to ensure log dir",
			slog.String("dir", cfg.Logging.LogDir),
			slog.String("error ", err.Error()))
		os.Exit(1)
	}

	log.Info("starting application",
		slog.String("env", cfg.Env),
		slog.Int("grpc_port", cfg.GRPC.Port),
	)

	_, err = telemetry.InitMetrics(ctx, log, cfg)
	if err != nil {
		defaultLog.Fatalf("failed to init metrics: %v", err)
	}

	tracerProvider, err := telemetry.InitTracing(ctx, cfg.Otel.Endpoint, log)
	if err != nil {
		defaultLog.Fatalf("failed to init tracing: %v", err)
	}
	defer func() {
		if err = tracerProvider.Shutdown(context.Background()); err != nil {
			log.Error("failed to shutdown tracer", slog.String("error", err.Error()))
		}
	}()

	publisher, publisherClose, err := initPublisher(ctx, cfg, log)
	if err != nil {
		defaultLog.Fatalf("failed to init publisher: %v", err)
	}
	defer publisherClose()

	application := app.New(ctx, log, cfg, tracerProvider, publisher)

	go application.GRPCSrv.MustRun()

	// Graceful shutdown.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sign := <-stop
	log.Info("stopping application", slog.String("signal", sign.String()))
	application.GRPCSrv.Stop()

	log.Info("application stopped")
}

func setupLogger(cfg *config.Config) *slog.Logger {
	var log *slog.Logger
	logFilePath := filepath.Join(cfg.Logging.LogDir, "app.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Error("failed to open log file",
			slog.String("logFilePath", logFilePath),
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Use io.MultiWriter to write logs to both stdout and file.
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	switch cfg.Env {
	case envLocal:
		prettyHandler := slogpretty.PrettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{Level: slog.LevelDebug},
		}.NewPrettyHandler(multiWriter)
		log = slog.New(prettyHandler)
	case envDev, envProd:
		log = slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{Level: slog.LevelDebug}))
	default:
		log = slog.New(slog.NewTextHandler(multiWriter, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	return log
}

func initPublisher(ctx context.Context,
	cfg *config.Config,
	log *slog.Logger) (
	publisher.InterfacePublisher,
	func(),
	error) {
	rabbitConn, err := rabbitmq.NewRabbitMQ(ctx, cfg)
	if err != nil {
		log.ErrorContext(ctx, "Failed to create RabbitMQ connection", "error", err)
		return nil, nil, err
	}

	pub, err := publisher.NewPublisher(rabbitConn, cfg, log)
	if err != nil {
		log.ErrorContext(ctx, "Failed to create RabbitMQ publisher", "error", err)
		return nil, nil, err
	}

	stopFn := func() {
		if err = pub.PublisherClose(); err != nil {
			log.ErrorContext(ctx, "Error closing publisher", slog.Any("error", err))
		}
		defer func() {
			if err = rabbitConn.Close(); err != nil {
				log.ErrorContext(ctx, "Error closing connection", slog.Any("error", err))
			}
		}()
	}
	return pub, stopFn, nil
}
