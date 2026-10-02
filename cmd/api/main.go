package main

import (
	"context"
	"errors"
	"fmt"
	"lessonHttp/config"
	_ "lessonHttp/config"
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
	tokenManager := token.NewManager([]byte(conf.JWTSecret), conf.JWTTTL)
	service := user.NewService(repo, tokenManager)
	handler := user.NewHandler(service)

	mux := http.NewServeMux()

	authorization := middleware.Auth(tokenManager)
	mux.Handle("GET /users", authorization(http.HandlerFunc(handler.ListHandler)))
	mux.Handle("GET /users/{id}", authorization(http.HandlerFunc(handler.UserByIDHandler)))

	mux.HandleFunc("POST /auth/register", handler.RegisterHandler)
	mux.HandleFunc("POST /auth/login", handler.LoginHandler)

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

	printAbout(conf)
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

func printAbout(conf config.Config) {
	fmt.Printf("Server %q start...\n", conf.HTTPAddr)
	fmt.Printf("Config mode: %q", conf.NameConfig)
	fmt.Println("")
	fmt.Println("GET /users\t\t\t - list all users")
	fmt.Println("GET /user/{id} \t\t\t - get user by id(int)")
	fmt.Println("POST /auth/register \t\t\t - create user ")
	fmt.Println("\t\t\tname - string")
	fmt.Println("\t\t\tage - int")
	fmt.Println("\t\t\temail - string")
	fmt.Println("\t\t\tpassword - string")
	fmt.Println("POST /auth/login \t- authontication user")
	fmt.Println("\t\t\tname - string")
	fmt.Println("\t\t\tpassword - string")
}
