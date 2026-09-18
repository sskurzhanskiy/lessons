package main

import (
	"context"
	"errors"
	"lessonHttp/config"
	_ "lessonHttp/config"
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

	conf, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := pgxpool.New(ctx, conf.DatabaseURL)
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

	mux.HandleFunc("GET /users", handler.ListHandler)
	mux.HandleFunc("GET /users/{id}", handler.UserByIDHandler)
	mux.HandleFunc("POST /users", handler.CreateUserHandler)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handlers := middleware.RequestID(middleware.Logging(logger, middleware.Recovery(logger, mux)))

	server := &http.Server{
		Addr:              conf.HTTPAddr,
		Handler:           handlers,
		ReadHeaderTimeout: conf.HTTPReadHeaderTimeout,
		ReadTimeout:       conf.HTTPReadTimeout,
		WriteTimeout:      conf.HTTPWriteTimeout,
		IdleTimeout:       conf.HTTPIdleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	gracefullShutdown(ctx, server, serverErr)
}

func gracefullShutdown(ctx context.Context, server *http.Server, serverErr <-chan error) {
	sigtermCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server stopped: %v\n", err)
			return
		}

	case <-sigtermCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		err := server.Shutdown(shutdownCtx)
		cancel()

		if err != nil {
			log.Printf("server shutdown: %v\n", err)
			if err = server.Close(); err != nil {
				log.Printf("server close: %v", err)
			}
			return
		}

		err = <-serverErr
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server stopped: %v", err)
		}
	}
}
