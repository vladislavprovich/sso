package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/vladislavprovich/sso/internal/app/grpc"
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
	// TODO: init storage

	// TODO: init auth service (auth)

	grpcApp := grpcapp.New(log, port)

	return &App{
		GRPCSrv: grpcApp,
	}
}
