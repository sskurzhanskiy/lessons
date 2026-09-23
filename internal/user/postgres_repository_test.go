package user

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setupPostgresRepository(t *testing.T) *PostgresRepository {
	t.Helper()

	databaseUrl := os.Getenv("TEST_DATABASE_URL")
	if databaseUrl == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, databaseUrl)
	if err != nil {
		t.Fatalf("create pgxpool: %v", err)
	}

	// _, err = db.Exec(ctx, `TRUNCATE TABLE users RESTART IDENTITY`) // сбрасывает счетчик ID
	// if err != nil {
	// 	t.Fatalf("truncate users: %v", err)
	// }

	tx, err := db.Begin(ctx)
	if err != nil {
		db.Close()
		t.Fatalf("start transaction: %v", err)
	}
	t.Cleanup(func() {
		tx.Rollback(ctx)
		db.Close()
	})

	return NewPostgresRepository(tx)
}

func TestPostgresRepositoryCreateAndByID(t *testing.T) {
	repo := setupPostgresRepository(t)

	ctx := context.Background()
	created, err := repo.Create(ctx, CreateUserParams{
		Name: "Alice",
		Age:  23,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if created.ID == 0 {
		t.Errorf("created user has id zero")
	}

	got, err := repo.ByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get user: %v id %d", err, created.ID)
	}

	if created != got {
		t.Errorf("got: %+v; want %+v", got, created)
	}
}

func TestPostgresRepositoryErrNotFound(t *testing.T) {
	repo := setupPostgresRepository(t)

	ctx := context.Background()
	_, err := repo.ByID(ctx, 999)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("error %v; want %v", err, ErrNotFound)
	}
}
