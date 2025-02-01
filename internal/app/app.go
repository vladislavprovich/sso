package app

import (
	"fmt"
	_ "github.com/lib/pq"
	grpcapp "github.com/vladislavprovich/sso/internal/app/grpc"
	"github.com/vladislavprovich/sso/internal/config"
	"github.com/vladislavprovich/sso/internal/services/auth"
	pg "github.com/vladislavprovich/sso/internal/storage/postgres"
	"log/slog"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	cfg *config.Config,
) *App {
	postgresDB := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host,
		cfg.Database.Port, cfg.Database.DBName, cfg.Database.SSLMode)

	storage, err := pg.New(postgresDB)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, cfg.TokenTTL)

	grpcApp := grpcapp.New(log, authService, cfg.GRPC.Port)

	return &App{
		GRPCSrv: grpcApp,
	}
}
