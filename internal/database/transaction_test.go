package database

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRetry(t *testing.T) {
	var someErr = errors.New("something went wrong")
	var retryableErr = &pgconn.PgError{Code: "40P01"}
	var serializationErr = &pgconn.PgError{Code: "40001"}

	maxAppempts := 3
	tests := []struct {
		name          string
		appemt        int
		wantCallCount int
		wantErr       error
		isCancel      bool
		isSuccess     bool
	}{
		{
			name:          "Attemption is zero",
			appemt:        0,
			wantCallCount: 0,
			wantErr:       ErrAttemptsZero,
		},
		{
			name:          "non-retryable",
			appemt:        maxAppempts,
			wantCallCount: 1,
			wantErr:       someErr,
		},
		{
			name:          "retryable always",
			appemt:        maxAppempts,
			wantCallCount: maxAppempts,
			wantErr:       retryableErr,
		},
		{
			name:          "retryable always serialization",
			appemt:        maxAppempts,
			wantCallCount: maxAppempts,
			wantErr:       serializationErr,
		},
		{
			name:          "context cancelled",
			appemt:        maxAppempts,
			wantCallCount: 1,
			wantErr:       serializationErr,
			isCancel:      true,
		},
		{
			name:          "success on second",
			appemt:        maxAppempts,
			wantCallCount: 2,
			wantErr:       nil,
			isSuccess:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var callCount int
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := retry(ctx, tt.appemt, func() error {
				callCount++
				if tt.isCancel {
					if callCount == tt.wantCallCount {
						cancel()
						return serializationErr
					}
				}
				if tt.isSuccess {
					if callCount != tt.wantCallCount {
						return serializationErr
					}
				}

				return tt.wantErr
			})

			if tt.isCancel {
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("error %v; want %v", err, context.Canceled)
				}
			} else {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error %v; want %v", err, tt.wantErr)
				}
			}

			if callCount != tt.wantCallCount {
				t.Errorf("call count is %d; want %d", callCount, tt.wantCallCount)
			}
		})
	}
}

func TestWriteSkewWithTransaction(t *testing.T) {
	ctx := context.Background()
	db := setupDB(ctx, t)
	cleanDB(ctx, db, t)

	tm := NewTransactionManager(db)
	if tm == nil {
		t.Fatal("Transaction manager don't create")
	}

	doctors := []string{"Alice", "Bob"}

	ready := make(chan int, 2)
	release := make(chan struct{})
	result := make(chan error, 2)

	for _, name := range doctors {
		go func(doctor string) {
			result <- tm.WithTransaction(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
				count, err := getCountAvailableDoctor(ctx, tx)
				if err != nil {
					return err
				}

				ready <- count
				<-release

				if count > 1 {
					_, err := tx.Exec(ctx, `UPDATE doctors SET on_call=false WHERE name=$1`, doctor)
					if err != nil {
						return err
					}
				}

				return nil
			})
		}(name)
	}
	countTx1 := <-ready
	countTx2 := <-ready
	close(release)

	if countTx1 != countTx2 {
		t.Fatal("count from transactions not equel; want count1 == count2")
	}
	if countTx1 != 2 {
		t.Fatalf("count from transaction = %d; want = 2", countTx1)
	}

	err1 := <-result
	err2 := <-result

	var txErr error
	switch {
	case err1 == nil && err2 != nil:
		txErr = err2
	case err1 != nil && err2 == nil:
		txErr = err1
	default:
		t.Fatalf("expected exactly one transaction to fail; err1=%v err2=%v", err1, err2)
	}

	var pgxErr *pgconn.PgError
	if !errors.As(txErr, &pgxErr) {
		t.Fatalf("expected pgError; got %v", txErr)
	}

	if pgxErr.Code != "40001" {
		t.Fatalf("got Code %q; want error Code = 40001", pgxErr.Code)
	}

	var count int
	count, err := getCountAvailableDoctor(ctx, db)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if count != 1 {
		t.Errorf("count of available docktors is %d; want = 1", count)
	}
}

var ErrLastDoctor = errors.New("cannot disable last on-call doctor")

func TestSerializableWriteSkew(t *testing.T) {
	ctx := context.Background()
	db := setupDB(ctx, t)
	cleanDB(ctx, db, t)

	tm := NewTransactionManager(db)
	if tm == nil {
		t.Fatal("Transaction manager don't create")
	}

	doctors := []string{"Alice", "Bob"}
	ready := make(chan int, 2)
	release := make(chan struct{})
	result := make(chan error, 2)

	for _, name := range doctors {
		go func(doctor string) {
			var attempt int
			result <- tm.WithRetry(ctx, 3, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
				attempt++
				count, err := getCountAvailableDoctor(ctx, tx)
				if err != nil {
					return err
				}

				if attempt == 1 {
					ready <- count
					<-release
				}

				if count <= 1 {
					return ErrLastDoctor
				}

				_, err = tx.Exec(ctx, `UPDATE doctors SET on_call=false WHERE name=$1`, doctor)
				if err != nil {
					return err
				}

				return nil
			})
		}(name)
	}
	countTx1 := <-ready
	countTx2 := <-ready
	close(release)

	if countTx1 != countTx2 {
		t.Fatal("count from transactions not equel; want count1 == count2")
	}
	if countTx1 != 2 {
		t.Fatalf("count from transaction = %d; want = 2", countTx1)
	}

	err1 := <-result
	err2 := <-result

	var txErr error
	switch {
	case err1 == nil && err2 != nil:
		txErr = err2
	case err1 != nil && err2 == nil:
		txErr = err1
	default:
		t.Fatalf("expected exactly one transaction to fail; err1=%v err2=%v", err1, err2)
	}

	if !errors.Is(txErr, ErrLastDoctor) {
		t.Fatalf("expected ErrLastDoctor; got %v", txErr)
	}

	count, err := getCountAvailableDoctor(ctx, db)

	if err != nil {
		t.Fatalf("get count %v", err)
	}
	if count != 1 {
		t.Errorf("count of available docktors is %d; want = 1", count)
	}
}

func getCountAvailableDoctor(ctx context.Context, db DBTX) (int, error) {
	var count int
	err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM doctors WHERE on_call = $1`,
		true).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func setupDB(ctx context.Context, t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("database url is empty")
	}

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create db %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	if err := db.Ping(ctx); err != nil {
		db.Close()
		t.Fatalf("ping db: %v", err)
	}

	return db
}

func cleanDB(ctx context.Context, db *pgxpool.Pool, t *testing.T) {
	t.Helper()

	_, err := db.Exec(ctx, `DELETE FROM doctors`)
	if err != nil {
		t.Fatalf("clean doctors: %v", err)
	}
	_, err = db.Exec(ctx, `
	INSERT INTO doctors (name, on_call) VALUES('Alice', true), ('Bob', true)
	`)
	if err != nil {
		t.Fatalf("create doctors: %v", err)
	}
}
