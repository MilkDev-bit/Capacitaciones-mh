package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	var dsn string

	if url := os.Getenv("DATABASE_URL"); url != "" {
		dsn = url
	} else {
		host := getEnv("DB_HOST", "localhost")
		port := getEnv("DB_PORT", "5432")
		user := getEnv("DB_USER", "postgres")
		password := getEnv("DB_PASSWORD", "")
		dbname := getEnv("DB_NAME", "capacitaciones")
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
	}

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		slog.Error("Error abriendo base de datos", "error", err)
		os.Exit(1)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	for i := 1; i <= 10; i++ {
		if err = DB.Ping(); err == nil {
			slog.Info("Conexión a PostgreSQL exitosa")
			return
		}
		slog.Warn("DB no disponible", "intento", i, "error", err)
		time.Sleep(3 * time.Second)
	}
	slog.Error("No se pudo conectar a la base de datos tras 10 intentos", "error", err)
	os.Exit(1)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
