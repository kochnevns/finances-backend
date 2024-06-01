package httpapp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	gw "github.com/kochnevns/finances-protos/finances" // import proto files for gateway to work
)

type App struct {
	port     int
	grpcPort int // gRPC server port
	log      *slog.Logger
} // App

func New(port int, grpcPort int, log *slog.Logger) *App {
	return &App{port: port, grpcPort: grpcPort, log: log}
}

func (a *App) MustRunHTTP() {
	if err := a.RunHTTP(); err != nil {
		panic(err)
	}
}

func (a *App) RunHTTP() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	grpcServerEnpoint := fmt.Sprintf(":%d", a.grpcPort)

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()

	h := cors.New(cors.Options{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"ACCEPT", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}).Handler(mux)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := gw.RegisterFinancesHandlerFromEndpoint(ctx, mux, grpcServerEnpoint, opts)
	if err != nil {
		return err
	}

	a.log.Info("Starting HTTP server", slog.String("port", strconv.Itoa(a.port))) // log

	if err := http.ListenAndServe(fmt.Sprintf(":%d", a.port), h); err != nil {
		return err
	}

	return nil
}
