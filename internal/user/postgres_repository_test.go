package user

import (
	"context"
	"errors"
	"os"
	"testing"

	"lessonHttp/internal/database"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setupPostgresRepository(t *testing.T) *PostgresRepository {
	t.Helper()

	ctx := context.Background()
	db := setupDB(ctx, t)

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("start transaction: %v", err)
	}
	t.Cleanup(func() {
		tx.Rollback(ctx)
	})

	initialDB(ctx, tx, t)

	return NewPostgresRepository(tx)
}

func setupDB(ctx context.Context, t *testing.T) *pgxpool.Pool {
	databaseUrl := os.Getenv("TEST_DATABASE_URL")
	if databaseUrl == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	db, err := pgxpool.New(ctx, databaseUrl)
	if err != nil {
		t.Fatalf("create pgxpool: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("ping db %v", err)
	}

	return db
}

func initialDB(ctx context.Context, db database.DBTX, t *testing.T) {
	_, err := db.Exec(ctx, `DELETE FROM users`)
	if err != nil {
		t.Fatalf("clean users: %v", err)
	}
	_, err = db.Exec(ctx, `
	INSERT INTO users (name, age) VALUES('Alice', 27), ('Bob', 15), ('Charlie', 25), ('David', 37), ('Eve', 11)
	`)
	if err != nil {
		t.Fatalf("create users: %v", err)
	}
}

func TestPostgresRepositoryCreateAndByID(t *testing.T) {
	repo := setupPostgresRepository(t)

	ctx := context.Background()
	created, err := repo.Create(ctx, User{
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

func TestList(t *testing.T) {
	t.Helper()

	repo := setupPostgresRepository(t)
	ctx := context.Background()

	var tests = []struct {
		name      string
		limit     int
		offset    int
		wantUsers []User
	}{
		{
			name:  "limit 2 offset 0",
			limit: 2,
			wantUsers: []User{
				{
					Name: "Alice",
				},
				{
					Name: "Bob",
				},
			},
		},
		{
			name:   "limit 2 offset 2",
			limit:  2,
			offset: 2,
			wantUsers: []User{
				{
					Name: "Charlie",
				},
				{
					Name: "David",
				},
			},
		},
		{
			name:   "limit 2 oggset 4",
			limit:  2,
			offset: 4,
			wantUsers: []User{
				{
					Name: "Eve",
				},
			},
		},
		{
			name:      "out of bounds",
			limit:     2,
			offset:    10,
			wantUsers: []User{},
		},
		{
			name:      "empty result",
			limit:     10,
			offset:    1000,
			wantUsers: []User{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, err := repo.List(ctx, tt.limit, tt.offset)
			if err != nil {
				t.Fatalf("list users %v", err)
			}

			if users == nil {
				t.Error("users must not be nil")
			}

			if len(users) != len(tt.wantUsers) {
				t.Errorf("count users %d; want = %d", len(users), len(tt.wantUsers))
			}

			for i := 0; i < len(users); i++ {
				if users[i].Name != tt.wantUsers[i].Name {
					t.Errorf("order wrong")
					break
				}
			}
		})
	}
}
