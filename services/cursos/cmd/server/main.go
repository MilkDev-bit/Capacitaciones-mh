// cursos-service: gestiona capacitaciones, inscripciones y asignaciones.
package main

import (
	"log/slog"
	"net"
	"os"

	cursospb "Prueba-Go/gen/cursos"
	"Prueba-Go/services/cursos/internal/handler"
	"Prueba-Go/services/cursos/internal/repository"
	"Prueba-Go/services/cursos/internal/service"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	db, err := sqlx.Connect("postgres", requireEnv("DATABASE_URL"))
	if err != nil {
		slog.Error("DB", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// DI
	repo := repository.NewCursosRepository(db)
	svc := service.NewCursosService(repo)
	h := handler.NewCursosHandler(svc)

	lis, _ := net.Listen("tcp", ":"+getEnvOr("GRPC_PORT", "50053"))
	srv := grpc.NewServer()

	cursospb.RegisterCursosServiceServer(srv, h)
	reflection.Register(srv)

	slog.Info("cursos-service iniciado", "port", getEnvOr("GRPC_PORT", "50053"))
	if err := srv.Serve(lis); err != nil {
		slog.Error("Serve", "error", err)
		os.Exit(1)
	}
}

func requireEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		slog.Error("variable requerida", "key", k)
		os.Exit(1)
	}
	return v
}

func getEnvOr(k, fb string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fb
}
