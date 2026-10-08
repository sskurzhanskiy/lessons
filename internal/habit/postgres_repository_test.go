package habit

import (
	"context"
	"errors"
	"lessonHttp/internal/database"
	"lessonHttp/internal/user"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Ptr[T any](v T) *T {
	return &v
}

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

func createUser(t *testing.T, ctx context.Context, db database.DBTX, name string, email string) (int, error) {
	t.Helper()

	repo := user.NewPostgresRepository(db)
	params := user.CreateUserParams{
		Name:         name,
		Age:          27,
		Email:        email,
		PasswordHash: "12345",
	}
	user, err := repo.Create(ctx, params)
	if err != nil {
		t.Fatalf("create user %v", err)
	}

	return user.ID, nil
}

func TestPostgresRepositoryUserConflict(t *testing.T) {
	repo := setupPostgresRepository(t)
	ctx := context.Background()
	userID, _ := createUser(t, ctx, repo.db, "Alice", "alice@email.com")
	createParam := CreateHabit{
		UserID:      userID,
		Title:       "test 2",
		ScheduledAt: time.Date(2026, time.October, 5, 15, 0, 0, 0, time.UTC),
		Comment:     "comment 2",
	}
	habit, err := repo.Create(ctx, createParam)
	if err != nil {
		t.Fatalf("create habit error %v", err)
	}

	fUserID, _ := createUser(t, ctx, repo.db, "Bob", "bob@email.com")
	_, err = repo.ByID(ctx, fUserID, habit.ID)
	if err == nil {
		t.Errorf("got habit error %v", err)
	}

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got error %v;want %v", err, ErrNotFound)
	}
}

func TestPostgresRepositoryCreate(t *testing.T) {
	repo := setupPostgresRepository(t)
	ctx := context.Background()
	userRepo := user.NewPostgresRepository(repo.db)
	params := user.CreateUserParams{
		Name:         "Alice",
		Age:          27,
		Email:        "alice@email.com",
		PasswordHash: "12345",
	}
	user, err := userRepo.Create(ctx, params)
	if err != nil {
		t.Fatalf("create user %v", err)
	}

	param := CreateHabit{
		UserID:      user.ID,
		Title:       "test 2",
		ScheduledAt: time.Date(2026, time.October, 5, 15, 0, 0, 0, time.Local),
		Comment:     "comment 2",
	}

	habit, err := repo.Create(ctx, param)
	if err != nil {
		t.Fatalf("got %v;unexpeced error", err)
	}

	if habit.ID == 0 {
		t.Error("got id 0, want not 0")
	}
	if habit.UserID != param.UserID {
		t.Errorf("got user id %d, want %d", habit.UserID, param.UserID)
	}
	if habit.Title != param.Title {
		t.Errorf("got title %q, want %q", habit.Title, param.Title)
	}
	if habit.ScheduledAt != param.ScheduledAt {
		t.Errorf("got schedule at %q, want %q", habit.ScheduledAt, param.ScheduledAt)
	}
	if habit.IsComplete != param.IsComplete {
		t.Errorf("got is comlete %t; want %t", habit.IsComplete, param.IsComplete)
	}
	if habit.Comment != param.Comment {
		t.Errorf("got comment %q, want %q", habit.Comment, param.Comment)
	}
}

func TestPostgresRepositorySuccessUpdate(t *testing.T) {
	repo := setupPostgresRepository(t)
	ctx := context.Background()
	userRepo := user.NewPostgresRepository(repo.db)
	params := user.CreateUserParams{
		Name:         "Alice",
		Age:          27,
		Email:        "alice@email.com",
		PasswordHash: "12345",
	}
	user, err := userRepo.Create(ctx, params)
	if err != nil {
		t.Fatalf("create user %v", err)
	}

	param := CreateHabit{
		UserID:      user.ID,
		Title:       "test 2",
		ScheduledAt: time.Date(2026, time.October, 5, 15, 0, 0, 0, time.Local),
		Comment:     "comment 2",
	}

	habit, err := repo.Create(ctx, param)
	if err != nil {
		t.Fatalf("create habit %v", err)
	}

	update := UpdateHabit{
		Title:       Ptr("update title"),
		ScheduledAt: Ptr(time.Date(2025, time.September, 3, 9, 0, 0, 0, time.Local)),
		IsComplete:  Ptr(true),
		Comment:     Ptr("update comment"),
	}

	updateHabit, err := repo.Update(ctx, user.ID, habit.ID, update)
	if err != nil {
		t.Fatalf("update habit %v", err)
	}

	if updateHabit.ID != habit.ID {
		t.Errorf("got habit id %d;want %d", updateHabit.ID, habit.ID)
	}
	if updateHabit.UserID != habit.UserID {
		t.Errorf("got habit userID %d;want %d", updateHabit.UserID, habit.UserID)
	}
	if updateHabit.Title != *update.Title {
		t.Errorf("got habit id %q;want %v", updateHabit.Title, update.Title)
	}
	if updateHabit.ScheduledAt != *update.ScheduledAt {
		t.Errorf("got scheduled %q;want %v", updateHabit.ScheduledAt, update.ScheduledAt)
	}
	if updateHabit.IsComplete != *update.IsComplete {
		t.Errorf("got is completed %t;want %v", updateHabit.IsComplete, update.IsComplete)
	}
	if updateHabit.Comment != *update.Comment {
		t.Errorf("got comment %q;want %v", updateHabit.Comment, update.Comment)
	}
}

func TestPostgresRepositoryNotFoundUpdate(t *testing.T) {
	repo := setupPostgresRepository(t)
	ctx := context.Background()
	userRepo := user.NewPostgresRepository(repo.db)
	params := user.CreateUserParams{
		Name:         "Alice",
		Age:          27,
		Email:        "alice@email.com",
		PasswordHash: "12345",
	}
	user, err := userRepo.Create(ctx, params)
	if err != nil {
		t.Fatalf("create user %v", err)
	}

	param := CreateHabit{
		UserID:      user.ID,
		Title:       "test 2",
		ScheduledAt: time.Date(2026, time.October, 5, 15, 0, 0, 0, time.UTC),
		Comment:     "comment 2",
	}

	habit, err := repo.Create(ctx, param)
	if err != nil {
		t.Fatalf("create user %v", err)

	}

	update := UpdateHabit{
		Title:       Ptr("update title"),
		ScheduledAt: Ptr(time.Date(2025, time.September, 3, 9, 0, 0, 0, time.UTC)),
		IsComplete:  Ptr(true),
		Comment:     Ptr("update comment"),
	}

	_, err = repo.Update(ctx, 5, habit.ID, update)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got error %v;want %v", err, ErrNotFound)
	}
}

func createHabit(ctx context.Context, repo Repository, userID int, title string) (Habit, error) {
	param := CreateHabit{
		UserID:      userID,
		Title:       title,
		ScheduledAt: time.Date(2026, time.October, 5, 15, 0, 0, 0, time.UTC),
		Comment:     "comment 2",
	}
	habit, err := repo.Create(ctx, param)

	return habit, err
}

func TestIntergrationHabit(t *testing.T) {
	ctx := context.Background()
	repo := setupPostgresRepository(t)
	aliceID, _ := createUser(t, ctx, repo.db, "Alice", "alice@email.com")
	bobID, _ := createUser(t, ctx, repo.db, "Bob", "bob@email.com")

	aliceHabit, err := createHabit(ctx, repo, aliceID, "alice habit")
	if err != nil {
		t.Fatalf("create alice habit %v", err)
	}

	_, err = repo.ByID(ctx, aliceID, aliceHabit.ID)
	if err != nil {
		t.Errorf("Alice try to get habit. Error %v;want not error", err)
	}
	_, err = repo.ByID(ctx, bobID, aliceHabit.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Bob try to get Alice's habit. Expected %v;got %v", ErrNotFound, err)
	}

	update := UpdateHabit{
		ScheduledAt: Ptr(time.Date(2026, 10, 8, 16, 41, 0, 0, time.Local)),
		IsComplete:  Ptr(true),
	}

	_, err = repo.Update(ctx, bobID, aliceHabit.ID, update)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Bob try to update Alice's habit. Expected %v; got %v", ErrNotFound, err)
	}

	err = repo.Delete(ctx, bobID, aliceHabit.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Bob try to delete Alice's habit. Expected %v; got %v", ErrNotFound, err)
	}

	_, err = repo.ByID(ctx, aliceID, aliceHabit.ID)
	if err != nil {
		t.Errorf("Alice try to get habit. Error %v;want not error", err)
	}

	habit, err := repo.Update(ctx, aliceID, aliceHabit.ID, update)
	if err != nil {
		t.Fatalf("Alice try to update habit. Error %v;want not error", err)
	}
	if habit.ID != aliceHabit.ID {
		t.Errorf("habit id %d;want %d", habit.ID, aliceHabit.ID)
	}
	if habit.Title != aliceHabit.Title {
		t.Errorf("title %q;want %q", habit.Title, aliceHabit.Title)
	}
	if !habit.ScheduledAt.Equal(*update.ScheduledAt) {
		t.Errorf("schedule %q;want %q", habit.ScheduledAt, *update.ScheduledAt)
	}

	if habit.IsComplete != *update.IsComplete {
		t.Errorf("is complete %t;want %t", habit.IsComplete, *update.IsComplete)
	}
	if habit.Comment != aliceHabit.Comment {
		t.Errorf("comment %q;want %q", habit.Comment, aliceHabit.Comment)
	}

	err = repo.Delete(ctx, aliceID, aliceHabit.ID)
	if err != nil {
		t.Errorf("Alice try to delete habit. Error %v;want not error", err)
	}
	_, err = repo.ByID(ctx, aliceID, aliceHabit.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v;want %v", err, ErrNotFound)
	}
}
