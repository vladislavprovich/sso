package app

import (
	"fmt"
	defaultLog "log"
	"log/slog"
	"net"
	"strconv"

	"github.com/vladislavprovich/sso/internal/rabbitmq/publisher"

	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/context"

	_ "github.com/lib/pq" // Used for Postgres DB. PG driver.
	grpcapp "github.com/vladislavprovich/sso/internal/app/grpc"
	"github.com/vladislavprovich/sso/internal/config"
	"github.com/vladislavprovich/sso/internal/services/auth"
	pg "github.com/vladislavprovich/sso/internal/storage/postgres"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	ctx context.Context,
	log *slog.Logger,
	cfg *config.Config,
	trace trace.TracerProvider,
	publisher publisher.UserPublisher,
) *App {
	hostAndPort := net.JoinHostPort(cfg.Postgres.Host, strconv.Itoa(cfg.Postgres.Port))

	postgresDataBaseURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		hostAndPort,
		cfg.Postgres.DBName,
		cfg.Postgres.SSLMode,
	)

	storage, err := pg.New(postgresDataBaseURL, cfg)
	if err != nil {
		log.ErrorContext(ctx, "Error creating postgres storage", "error", err)
		defaultLog.Fatalf("Error creating postgres storage: %v", err)
	}

	authService := auth.New(log, storage, storage, storage, cfg.TokenTTL, trace, publisher)

	grpcApp := grpcapp.New(log, authService, cfg.GRPC.Port, trace.Tracer(cfg.Tracing.NameSpase))

	return &App{
		GRPCSrv: grpcApp,
	}
}
