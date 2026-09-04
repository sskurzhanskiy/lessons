package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionManager struct {
	db *pgxpool.Pool
}

func NewTransactionManager(db *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{db: db}
}

func (t *TransactionManager) WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err = fn(tx); err != nil {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			err = errors.Join(err, fmt.Errorf("rollback transaction: %w", rollbackErr))
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

const retryDelay = 3000 * time.Millisecond

func (t *TransactionManager) WithRetry(ctx context.Context, maxAttempts int, fn func(tx pgx.Tx) error) error {
	return retry(ctx, maxAttempts, func() error {
		return t.WithTransaction(ctx, fn)
	})
}

var ErrAttemptsZero = errors.New("retry attemptions must be greater than zero")

func retry(ctx context.Context, maxAttempts int, fn func() error) error {
	if maxAttempts <= 0 {
		return ErrAttemptsZero
	}

	for attemp := 0; attemp < maxAttempts; attemp++ {
		err := fn()
		if err == nil {
			return nil
		}

		if !isRetryable(err) {
			return err
		}

		if attemp == maxAttempts-1 {
			return err
		}

		if err := waitRetry(ctx); err != nil {
			return err
		}
	}

	return nil
}

func waitRetry(ctx context.Context) error {
	timer := time.NewTimer(retryDelay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func isRetryable(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "40P01" || pgErr.Code == "40001"
	}

	return false
}
