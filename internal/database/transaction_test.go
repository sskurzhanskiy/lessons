package database

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestRetryNotPgError(t *testing.T) {
	var callCount int
	err := retry(context.Background(), 3, func() error {
		callCount++
		return nil
	})

	if err != nil {
		t.Error("expected error")
	}

	if callCount != 1 {
		t.Errorf("attemption is %d; want = 1", callCount)
	}
}

func TestRetryMaxAttemptZero(t *testing.T) {
	var callCount int
	err := retry(context.Background(), 0, func() error {
		callCount++
		return nil
	})

	if !errors.Is(err, ErrAttemptsZero) {
		t.Errorf("expected error ErrAttemptsZero, got %T", err)
	}

	if callCount != 0 {
		t.Errorf("attemption is %d; want = 0", callCount)
	}
}

func TestRetryNoPgError(t *testing.T) {
	var callCount int
	err := retry(context.Background(), 3, func() error {
		callCount++
		return fmt.Errorf("custom error")
	})

	if err.Error() != fmt.Errorf("custom error").Error() {
		t.Errorf("expected error %v, got %v", fmt.Errorf("custom error"), err)
	}

	if callCount != 1 {
		t.Errorf("attemption is %d; want = 1", callCount)
	}
}

func TestRetryPgError(t *testing.T) {
	var callCount int
	maxAttempts := 3
	err := retry(context.Background(), maxAttempts, func() error {
		callCount++
		return &pgconn.PgError{
			Code: "40001",
		}
	})

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code != "40001" {
			t.Errorf("error pgError.Code = %q; want pgError.Code = 40001", pgErr.Code)
		}
	}

	if callCount != maxAttempts {
		t.Errorf("attemption is %d; want = 1", maxAttempts)
	}
}

func TestRetrySuccess(t *testing.T) {
	var callCount int
	err := retry(context.Background(), 3, func() error {
		callCount++
		if callCount == 2 {
			return nil
		}

		return &pgconn.PgError{
			Code: "40P01",
		}
	})

	if err != nil {
		t.Fatal("unexpected error")
	}

	if callCount != 2 {
		t.Errorf("attemption is %d; want = 2", callCount)
	}
}

func TestRetryCtxCancel(t *testing.T) {
	var callCount int
	ctx, cancel := context.WithCancel(context.Background())
	err := retry(ctx, 5, func() error {
		callCount++
		if callCount == 4 {
			cancel()
		}

		return &pgconn.PgError{
			Code: "40P01",
		}
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("error %T; want %T", err, context.Canceled)
	}
}
