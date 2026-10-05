package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type Config struct {
	NameConfig            string
	DatabaseURL           string
	HTTPAddr              string
	HTTPReadHeaderTimeout time.Duration
	HTTPReadTimeout       time.Duration
	HTTPWriteTimeout      time.Duration
	HTTPIdleTimeout       time.Duration
	JWTSecret             string
	JWTTTL                time.Duration
}

func Load() (Config, error) {
	name := os.Getenv("HABBIT_CONFIG_NAME")
	if name == "" {
		return Config{}, errors.New("HABBIT_CONFIG_NAME is required")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8080"
	}

	httpReadHeaderTimeout, err := durationFromEnv("HTTP_READ_HEADER_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}

	httpReadTimeout, err := durationFromEnv("HTTP_READ_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	httpWriteTimeout, err := durationFromEnv("HTTP_WRITE_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, err
	}

	httpIdleTimeout, err := durationFromEnv("HTTP_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return Config{}, err
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}
	if len(jwtSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must be at least 32 bytes")
	}

	jwtTTL, err := durationFromEnv("JWT_TTL", 60*time.Minute)
	if err != nil {
		return Config{}, err
	}

	return Config{
		NameConfig:            name,
		DatabaseURL:           databaseURL,
		HTTPAddr:              httpAddr,
		HTTPReadHeaderTimeout: httpReadHeaderTimeout,
		HTTPReadTimeout:       httpReadTimeout,
		HTTPWriteTimeout:      httpWriteTimeout,
		HTTPIdleTimeout:       httpIdleTimeout,
		JWTSecret:             jwtSecret,
		JWTTTL:                jwtTTL,
	}, nil
}

func durationFromEnv(key string, defaultValue time.Duration) (time.Duration, error) {
	strValue := os.Getenv(key)
	if strValue == "" {
		return defaultValue, nil
	}

	result, err := time.ParseDuration(strValue)
	if err != nil {
		return 0, fmt.Errorf(
			"parse %q: %w",
			key,
			err,
		)
	}

	if result <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}

	return result, nil
}
