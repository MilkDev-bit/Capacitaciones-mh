package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		fmt.Println("No DATABASE_URL found")
		return
	}
	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		fmt.Printf("Unable to connect to database: %v\n", err)
		return
	}
	defer pool.Close()

	var exists bool
	err = pool.QueryRow(context.Background(), "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'videocall_tickets')").Scan(&exists)
	if err != nil {
		fmt.Printf("Query failed: %v\n", err)
		return
	}
	fmt.Printf("Table videocall_tickets exists: %v\n", exists)
}
