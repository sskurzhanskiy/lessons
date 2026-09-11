package main

import (
	"context"
	"errors"
	"lessonHttp/internal/middleware"
	"lessonHttp/internal/user"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handlers := middleware.RequestID(middleware.Logging(logger, middleware.Recovery(logger, mux)))
	server := &http.Server{
		Addr:              ":8080",
		Handler:           handlers,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	sigtermCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("server stopped: %v", err)
			} else {
				log.Println("correct stop server")
			}

		}
	}()

	<-sigtermCtx.Done()
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
}
