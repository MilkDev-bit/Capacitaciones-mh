// examenes-service: gestiona exámenes, preguntas, envío y calificaciones.
package main

import (
	"log/slog"
	"net"
	"os"

	examenespb "Prueba-Go/gen/examenes"
	"Prueba-Go/services/examenes/internal/handler"
	"Prueba-Go/services/examenes/internal/repository"
	"Prueba-Go/services/examenes/internal/service"

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

	repo := repository.NewExamenesRepository(db)
	svc := service.NewExamenesService(repo)
	h := handler.NewExamenesHandler(svc)

	lis, _ := net.Listen("tcp", ":"+getEnvOr("GRPC_PORT", "50055"))
	srv := grpc.NewServer()
	examenespb.RegisterExamenesServiceServer(srv, h)
	reflection.Register(srv)

	slog.Info("examenes-service iniciado", "port", getEnvOr("GRPC_PORT", "50055"))
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
