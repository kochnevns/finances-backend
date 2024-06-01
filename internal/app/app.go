package app

import (
	"log/slog"

	grpcapp "github.com/kochnevns/finances-backend/internal/app/grpc"
	httpapp "github.com/kochnevns/finances-backend/internal/app/http"
	"github.com/kochnevns/finances-backend/internal/services/finances"
	"github.com/kochnevns/finances-backend/internal/storage/sqlite"
)

type App struct {
	GRPCServer *grpcapp.App
	HTTPServer *httpapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	httpPort int,
	storagePath string,
) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	financesService := finances.New(log, storage, storage, storage, storage, storage)

	grpcApp := grpcapp.New(log, financesService, grpcPort)
	httpApp := httpapp.New(httpPort, grpcPort, log)

	return &App{
		GRPCServer: grpcApp,
		HTTPServer: httpApp,
	}
}
