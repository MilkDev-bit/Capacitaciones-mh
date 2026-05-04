package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	var dsn string

	// Railway inyecta DATABASE_URL automáticamente al vincular el plugin de PostgreSQL.
	// lib/pq acepta la URL directamente, que ya incluye los parámetros SSL de Railway.
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
		log.Fatalf("Error abriendo base de datos: %v", err)
	}
	// Retry ping hasta 10 intentos (Railway puede tardar unos segundos en levantar el plugin de PostgreSQL)
	for i := 1; i <= 10; i++ {
		if err = DB.Ping(); err == nil {
			log.Println("Conexión a PostgreSQL exitosa")
			return
		}
		log.Printf("DB no disponible, intento %d/10: %v", i, err)
		time.Sleep(3 * time.Second)
	}
	log.Fatalf("No se pudo conectar a la base de datos tras 10 intentos: %v", err)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
