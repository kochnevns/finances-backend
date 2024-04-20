package app

import (
	grpcapp "github.com/kochnevns/finances-backend/internal/app/grpc"
	"github.com/kochnevns/finances-backend/internal/services/finances"
	"github.com/kochnevns/finances-backend/internal/storage/sqlite"
	"log/slog"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	storagePath string,
) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	financesService := finances.New(log, storage, storage, storage)

	grpcApp := grpcapp.New(log, financesService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
