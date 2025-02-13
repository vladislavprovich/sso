package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/vladislavprovich/sso/internal/lib/logger/handlers/slogpretty"

	"github.com/vladislavprovich/sso/internal/lib/telemetry"

	"github.com/vladislavprovich/sso/internal/app"
	"github.com/vladislavprovich/sso/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
	logDir   = "logs"
)

func main() {
	cfg := config.MustLoad()
	ctx := context.Background()
	log := setupLogger(cfg)

	// Init logs dir.
	err := telemetry.EnsureLogDir(logDir)
	if err != nil {
		log.Error("failed to ensure log dir",
			slog.String("dir", logDir),
			slog.String("error ", err.Error()))
		os.Exit(1)
	}

	log.Info("starting application",
		slog.String("env", cfg.Env),
		slog.Int("grpc_port", cfg.GRPC.Port),
	)

	_, err = telemetry.InitMetrics(ctx, log, cfg)
	if err != nil {
		log.Error("failed to initialize metrics", slog.String("error", err.Error()))
		os.Exit(1)
	}

	tracerProvider, err := telemetry.InitTracing(ctx, cfg.Otel.Endpoint, log)
	if err != nil {
		log.Error("failed to initialize tracing", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err = tracerProvider.Shutdown(context.Background()); err != nil {
			log.Error("failed to shutdown tracer", slog.String("error", err.Error()))
		}
	}()

	application := app.New(log, cfg, tracerProvider)

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
	logFilePath := filepath.Join(logDir, "app.log")
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
