package registration

import (
	"context"
	"errors"
	"fmt"
	"lessonHttp/internal/database"
	"lessonHttp/internal/profile"
	"lessonHttp/internal/unitofwork"
	"lessonHttp/internal/user"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FailingProfileRepository struct {
}

func NewFailingProfileRepository() *FailingProfileRepository {
	return &FailingProfileRepository{}
}

func (r *FailingProfileRepository) Create(ctx context.Context, profile profile.Profile) error {
	return fmt.Errorf("create profile error")
}

type TransactionManager struct {
	txManager *database.TransactionManager
}

func NewTransactionManager(txManager *database.TransactionManager) *TransactionManager {
	return &TransactionManager{txManager: txManager}
}

func (t *TransactionManager) WithTransaction(ctx context.Context, fn func(uow *unitofwork.UnitOfWork) error) error {
	return t.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		uow := unitofwork.New(user.NewPostgresRepository(tx), NewFailingProfileRepository())
		return fn(uow)
	})
}

type FakeTransactionManager struct {
	Calls int
}

func NewFakeTransactionManager() *FakeTransactionManager {
	return &FakeTransactionManager{}
}

func (t *FakeTransactionManager) WithTransaction(ctx context.Context, fn func(uow *unitofwork.UnitOfWork) error) error {
	t.Calls++
	return nil
}

type RegistrationModel struct {
	Name string
	Age  int
	Bio  string
}

func setupDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseUrl := os.Getenv("TEST_DATABASE_URL")

	if databaseUrl == "" {
		t.Skip("database url is empty")
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, databaseUrl)
	if err != nil {
		t.Fatalf("create db: %v", err)
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		t.Fatalf("ping db: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TestCreateUserWithProfileRollback(t *testing.T) {
	db := setupDB(t)

	txManager := database.NewTransactionManager(db)
	tm := NewTransactionManager(txManager)
	service := NewRegistrationService(tm)

	ctx := context.Background()
	name := fmt.Sprintf("rollback-test-name-%d", time.Now().UnixNano())
	err := service.CreateUserWithProfile(ctx, name, 23, "bio")

	if err == nil {
		t.Fatalf("expected create user with profile to fail")
	}

	var count int
	err = db.QueryRow(ctx,
		"SELECT COUNT(*) FROM users WHERE name=$1",
		name,
	).Scan(&count)
	if err != nil {
		t.Fatalf("check user existence: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected create user to be rollback; got count: %d", count)
	}
}

func TestCreateUserWithProfile(t *testing.T) {
	db := setupDB(t)

	txManager := database.NewTransactionManager(db)
	tm := unitofwork.NewTransactionManager(txManager)
	service := NewRegistrationService(tm)

	ctx := context.Background()
	name := fmt.Sprintf("rollback-test-name-%d", time.Now().UnixNano())
	err := service.CreateUserWithProfile(ctx, name, 12, "bio")

	if err != nil {
		t.Fatal("unexpected error")
	}

	var userId int
	err = db.QueryRow(ctx,
		"SELECT id FROM users WHERE name=$1",
		name,
	).Scan(&userId)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}

	var bio string
	err = db.QueryRow(ctx,
		"SELECT bio FROM profiles WHERE user_id=$1",
		userId,
	).Scan(&bio)
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}
	if bio != "bio" {
		t.Errorf("unexpected bio: got %q want: bio", bio)
	}
}

func TestCreateUserWithProfileValidationErrors(t *testing.T) {
	var tests = []struct {
		name     string
		userName string
		age      int
		bio      string
		wantErr  *ValidationError
	}{
		{
			name:     "name is empty",
			userName: "",
			age:      21,
			bio:      "bio",
			wantErr: &ValidationError{
				Field: "name",
			},
		},
		{
			name:     "age incorrect",
			userName: "Alice",
			age:      0,
			bio:      "bio",
			wantErr: &ValidationError{
				Field: "age",
			},
		},
		{
			name:     "bio is empty",
			userName: "Bob",
			age:      32,
			bio:      "",
			wantErr: &ValidationError{
				Field: "bio",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewFakeTransactionManager()
			service := NewRegistrationService(tm)

			ctx := context.Background()
			err := service.CreateUserWithProfile(ctx, tt.userName, tt.age, tt.bio)
			if err == nil {
				t.Fatal("expected error")
			}

			var validationErr *ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("expected ValidationError, got %T: %v", err, err)
			}

			if validationErr.Field != tt.wantErr.Field {
				t.Errorf("validation error for field = %q; want field = %q", validationErr.Field, tt.wantErr.Field)
			}

			if tm.Calls != 0 {
				t.Error("transaction call")
			}
		})
	}
}

func TestCreateUserWithProfileTransactionStart(t *testing.T) {
	tm := NewFakeTransactionManager()
	service := NewRegistrationService(tm)

	ctx := context.Background()
	err := service.CreateUserWithProfile(ctx, "Alice", 23, "bio")
	if err != nil {
		t.Fatal("expected error")
	}

	if tm.Calls != 1 {
		t.Error("transaction must call")
	}
}
