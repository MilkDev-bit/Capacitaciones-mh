// lecciones-service: gestiona lecciones, progreso y preguntas intermedias.
package main

import (
	"log/slog"
	"net"
	"os"

	leccionespb "Prueba-Go/gen/lecciones"
	"Prueba-Go/services/lecciones/internal/handler"
	"Prueba-Go/services/lecciones/internal/repository"
	"Prueba-Go/services/lecciones/internal/service"

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

	repo := repository.NewLeccionesRepository(db)
	svc := service.NewLeccionesService(repo)
	h := handler.NewLeccionesHandler(svc)

	lis, _ := net.Listen("tcp", ":"+getEnvOr("GRPC_PORT", "50054"))
	srv := grpc.NewServer()
	leccionespb.RegisterLeccionesServiceServer(srv, h)
	reflection.Register(srv)

	slog.Info("lecciones-service iniciado", "port", getEnvOr("GRPC_PORT", "50054"))
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
