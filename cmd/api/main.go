package main

import (
	"context"
	"lessonHttp/internal/user"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("create postgres pool: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("ping postgres: %v", err)
	}

	repo := user.NewPostgresRepository(db)
	service := user.NewService(repo)
	handler := user.NewHandler(service)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /users/{id}", handler.UserByIDHandler)
	mux.HandleFunc("POST /users", handler.CreateUserHandler)

	log.Println("server started on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
