package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/vladislavprovich/sso/internal/lib/telemetry"

	"github.com/vladislavprovich/sso/internal/app"
	"github.com/vladislavprovich/sso/internal/config"
	"github.com/vladislavprovich/sso/internal/lib/logger/handlers/slogpretty"
)

const (
	envLocal  = "local"
	envDev    = "dev"
	envProd   = "prod"
	nameSpace = "integration-services"
)

func main() {
	cfg := config.MustLoad()
	ctx := context.Background()
	log := setupLogger(cfg.Env)
	logLoki := telemetry.InitLogger()

	logLoki.Info("starting loki")
	log.Info("starting application",
		slog.String("env", cfg.Env),
		slog.Int("grpc_port", cfg.GRPC.Port),
	)

	_, err := telemetry.InitMetrics(ctx, log, cfg)
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

	application := app.New(log, cfg, tracerProvider.Tracer(nameSpace))

	go application.GRPCSrv.MustRun()

	// Graceful shutdown.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sign := <-stop
	log.Info("stopping application", slog.String("signal", sign.String()))
	application.GRPCSrv.Stop()

	log.Info("application stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		prettyHandler := slogpretty.PrettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{Level: slog.LevelDebug},
		}.NewPrettyHandler(os.Stdout)
		log = slog.New(prettyHandler)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	}

	return log
}
