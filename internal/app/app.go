package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/vladislavprovich/sso/internal/app/grpc"
	"github.com/vladislavprovich/sso/internal/services/auth"
	"github.com/vladislavprovich/sso/internal/storage/sqlite"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	port int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, port)

	return &App{
		GRPCSrv: grpcApp,
	}
}
