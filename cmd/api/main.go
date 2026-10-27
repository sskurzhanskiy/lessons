package main

import (
	"context"
	"errors"
	"fmt"
	"lessonHttp/config"
	_ "lessonHttp/config"
	"lessonHttp/internal/habit"
	"lessonHttp/internal/health"
	"lessonHttp/internal/middleware"
	"lessonHttp/internal/token"
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
	if err := run(); err != nil {
		log.Printf("application error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	conf, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config %w", err)
	}

	db, err := pgxpool.New(ctx, conf.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create postgres pool: %w", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	tokenManager := token.NewManager([]byte(conf.JWTSecret), conf.JWTTTL)
	userRepo := user.NewPostgresRepository(db)
	userService := user.NewService(userRepo, tokenManager)
	usersHandler := user.NewHandler(userService)

	habitRepo := habit.NewPostgresRepository(db)
	habitService := habit.NewService(habitRepo)
	habitHanlder := habit.NewHandler(habitService)

	mux := http.NewServeMux()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handlers := middleware.RequestID(middleware.Logging(logger, middleware.Recovery(logger, mux)))

	authorization := middleware.Auth(tokenManager)

	healthHandler := health.NewHandler()

	registerRoutes(mux, usersHandler, habitHanlder, healthHandler, authorization)

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

	err = gracefullShutdown(ctx, server, serverErr)
	if err != nil {
		return fmt.Errorf("gracefull shutdown %w", err)
	}

	return nil
}

func registerRoutes(mux *http.ServeMux, usersHandler *user.Handler, habitHandler *habit.Handler, healthHandler *health.Handler, authorization func(http.Handler) http.Handler) {
	mux.Handle("GET /users", authorization(http.HandlerFunc(usersHandler.ListHandler)))
	mux.Handle("GET /users/{id}", authorization(http.HandlerFunc(usersHandler.UserByIDHandler)))

	mux.Handle("GET /habits/{id}", authorization(http.HandlerFunc(habitHandler.ByIDHandler)))
	mux.Handle("GET /habits", authorization(http.HandlerFunc(habitHandler.ListHandler)))
	mux.Handle("POST /habits", authorization(http.HandlerFunc(habitHandler.CreateHandler)))
	mux.Handle("PATCH /habits/{id}", authorization(http.HandlerFunc(habitHandler.UpdateHandler)))
	mux.Handle("DELETE /habits/{id}", authorization(http.HandlerFunc(habitHandler.DeleteHandler)))

	mux.HandleFunc("POST /auth/register", usersHandler.RegisterHandler)
	mux.HandleFunc("POST /auth/login", usersHandler.LoginHandler)
	mux.HandleFunc("GET /healthz", healthHandler.HealthzHandler)
}

func gracefullShutdown(ctx context.Context, server *http.Server, serverErr <-chan error) error {
	sigtermCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil

	case <-sigtermCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		err := server.Shutdown(shutdownCtx)
		cancel()

		if err != nil {
			if closeErr := server.Close(); closeErr != nil {
				return errors.Join(err, closeErr)
			}
			return err
		}

		err = <-serverErr
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	}
}
