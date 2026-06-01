package main

import (
	"log/slog"
	"net"
	"os"

	usuariospb "Prueba-Go/gen/usuarios"
	"Prueba-Go/services/usuarios/internal/handler"
	"Prueba-Go/services/usuarios/internal/repository"
	"Prueba-Go/services/usuarios/internal/service"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	dbURL := requireEnv("DATABASE_URL")
	grpcPort := getEnvOr("GRPC_PORT", "50052")

	db, err := sqlx.Connect("pgx", dbURL)
	if err != nil {
		slog.Error("DB connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// DI: Repository → Service → Handler
	repo := repository.NewUsuarioRepository(db)
	svc := service.NewUsuariosService(repo)
	h := handler.NewUsuariosHandler(svc)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		slog.Error("net.Listen", "error", err)
		os.Exit(1)
	}

	srv := grpc.NewServer()
	usuariospb.RegisterUsuariosServiceServer(srv, h)
	reflection.Register(srv)

	slog.Info("usuarios-service iniciado", "port", grpcPort)
	if err := srv.Serve(lis); err != nil {
		slog.Error("gRPC Serve", "error", err)
		os.Exit(1)
	}
}

func requireEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		slog.Error("variable de entorno requerida", "key", k)
		os.Exit(1)
	}
	return v
}

func getEnvOr(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
