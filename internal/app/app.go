package app

import (
	"fmt"
	"log/slog"
	"net"
	"strconv"

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
	log *slog.Logger,
	cfg *config.Config,
) *App {
	hostAndPort := net.JoinHostPort(cfg.Database.Host, strconv.Itoa(cfg.Database.Port))

	postgresDB := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		hostAndPort,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	storage, err := pg.New(postgresDB, *cfg)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, cfg.TokenTTL)

	grpcApp := grpcapp.New(log, authService, cfg.GRPC.Port)

	return &App{
		GRPCSrv: grpcApp,
	}
}
