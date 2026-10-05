package habit

import (
	"context"
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

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("start transaction: %v", err)
	}
	t.Cleanup(func() {
		tx.Rollback(ctx)
		db.Close()
	})

	return NewPostgresRepository(tx)
}

func createUser(t *testing.T, ctx context.Context) (int, error) {
	repo := setupPostgresRepository(t)
	var userID int
	err := repo.db.QueryRow(ctx,
		`INTO INSERT users (name, age) VALUES ("Alice", 27) RETURNING id`,
	).Scan(&userID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	return userID, nil
}

/*
func TestPostgresRepositoryCreate(t *testing.T) {
	t.Helper()
	var tests = []struct {
		name      string
		param     CreateHabit
		wantHabit Habit
		wantErr   error
	}{
		{
			name: "Incorrect params",
			param: CreateHabit{
				Title: "test 1",
			},
			wantErr: fmt.Errorf("some error"),
		},
		{
			name: "Correct params",
			param: CreateHabit{
				UserID:      5,
				Title:       "test 2",
				ScheduledAt: time.Date(2026, time.October, 5, 15, 0, 0, 0, time.UTC),
				Comment:     "comment 2",
			},
			wantHabit: Habit{
				ID:          1,
				UserID:      5,
				Title:       "test 2",
				ScheduledAt: time.Date(2026, time.October, 5, 15, 0, 0, 0, time.UTC), //time.Now(),
				Comment:     "comment 2",
			},
		},
	}
	repo := setupPostgresRepository(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			habit, err := repo.Create(context.Background(), tt.param)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error")
				}
			} else {
				if err != nil {
					t.Fatalf("got %v;unexpeced error", err)
				}
			}

			if habit.ID != tt.wantHabit.ID {
				t.Errorf("got id %d, want %d", habit.ID, tt.wantHabit.ID)
			}
			if habit.UserID != tt.wantHabit.UserID {
				t.Errorf("got user id %d, want %d", habit.UserID, tt.wantHabit.UserID)
			}
			if habit.Title != tt.wantHabit.Title {
				t.Errorf("got title %q, want %q", habit.Title, tt.wantHabit.Title)
			}
			if habit.ScheduledAt != tt.wantHabit.ScheduledAt {
				t.Errorf("got schedule at %q, want %q", habit.ScheduledAt, tt.wantHabit.ScheduledAt)
			}
			if habit.IsComplete != tt.wantHabit.IsComplete {
				t.Errorf("got is comlete %t; want %t", habit.IsComplete, tt.wantHabit.IsComplete)
			}
			if habit.Comment != tt.wantHabit.Comment {
				t.Errorf("got comment %q, want %q", habit.Comment, tt.wantHabit.Comment)
			}
		})
	}

}
*/
