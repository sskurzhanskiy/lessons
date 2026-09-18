package user

import (
	"context"
	"errors"
	"testing"
)

type FakeUserRepository struct {
	CreateUser User
	CreateErr  error
	ByIDUser   User
	ByIDErr    error

	CreateArg User
	ByIDArg   int

	CreateCalled bool
	ByIDCalled   bool
}

func (r *FakeUserRepository) Create(ctx context.Context, user User) (User, error) {
	r.CreateCalled = true
	r.CreateArg = user
	return r.CreateUser, r.CreateErr
}

func (r *FakeUserRepository) ByID(ctx context.Context, id int) (User, error) {
	r.ByIDCalled = true
	r.ByIDArg = id
	return r.ByIDUser, r.ByIDErr
}

func (r *FakeUserRepository) List(ctx context.Context, limit int, offset int) ([]User, error) {
	return []User{}, nil
}

func TestServiceCreate(t *testing.T) {
	ctx := context.Background()
	repo := &FakeUserRepository{
		CreateUser: User{
			ID:   1,
			Name: "Alice",
			Age:  22,
		},
	}
	service := NewService(repo)

	user, err := service.Create(ctx, "Alice", 22)

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if !repo.CreateCalled {
		t.Fatal("repository Create was not called")
	}

	if repo.CreateArg.Name != "Alice" {
		t.Errorf("argument name %q; want Alice", repo.CreateArg.Name)
	}
	if repo.CreateArg.Age != 22 {
		t.Errorf("argument age %d; want 22", repo.CreateArg.Age)
	}

	if user.ID != 1 {
		t.Errorf("user id = %d; want id = 1", user.ID)
	}
	if user.Name != "Alice" {
		t.Errorf("user name = %q; want name = Alice", user.Name)
	}
	if user.Age != 22 {
		t.Errorf("user age = %d; want age = 22", user.Age)
	}
}

func TestServiceCreateInvalidName(t *testing.T) {
	ctx := context.Background()
	repo := &FakeUserRepository{}
	service := NewService(repo)

	_, err := service.Create(ctx, "", 22)
	if repo.CreateCalled {
		t.Errorf("repository Create was called for invalid input")
	}

	var validationErr *ValidationError
	if err == nil {
		t.Fatal("expected error")
	}

	if errors.As(err, &validationErr) {
		if validationErr.Field != "name" {
			t.Errorf("validation error field = %q; want field = name", validationErr.Field)
		}
	} else {
		t.Fatalf("error %T; want *ValidationError", err)
	}
}

func TestServiceCreateInvalidAge(t *testing.T) {
	ctx := context.Background()
	repo := &FakeUserRepository{}
	service := NewService(repo)

	_, err := service.Create(ctx, "Alice", 0)
	if repo.CreateCalled {
		t.Errorf("repository Create was called for invalid input")
	}

	var validationErr *ValidationError
	if err == nil {
		t.Fatal("expected error")
	}

	if errors.As(err, &validationErr) {
		if validationErr.Field != "age" {
			t.Errorf("validation error field = %q; want field = age", validationErr.Field)
		}
	} else {
		t.Fatalf("error %T; want *ValidationError", err)
	}
}

func TestServiceCreatePropagationError(t *testing.T) {
	var errRepo = errors.New("somewhere done error")
	ctx := context.Background()
	repo := &FakeUserRepository{
		CreateErr: errRepo,
	}
	service := NewService(repo)
	_, err := service.Create(ctx, "Alice", 31)
	if err == nil {
		t.Fatal("expected error")
	}

	if !repo.CreateCalled {
		t.Fatal("repository Create was not called")
	}

	if !errors.Is(err, errRepo) {
		t.Errorf("unexpected error %v; want %v", err, errRepo)
	}
}

func TestServiceByIDInvalidID(t *testing.T) {
	ctx := context.Background()
	repo := &FakeUserRepository{}
	service := NewService(repo)

	_, err := service.ByID(ctx, -1)
	if repo.ByIDCalled {
		t.Error("repository ByID was called for invalid input")
	}

	var validationErr *ValidationError
	if err == nil {
		t.Fatal("expected error")
	}

	if errors.As(err, &validationErr) {
		if validationErr.Field != "id" {
			t.Errorf("validation error field = %q; want = id", validationErr.Field)
		}
	} else {
		t.Fatalf("error %T; want *ValidationError", err)
	}
}
func TestServiceByIDNotFound(t *testing.T) {
	ctx := context.Background()
	repo := &FakeUserRepository{
		ByIDErr: ErrNotFound,
	}
	service := NewService(repo)

	_, err := service.ByID(ctx, 999)

	if err == nil {
		t.Fatal("expected error")
	}
	if !repo.ByIDCalled {
		t.Fatal("repository ByID was not called")
	}

	if repo.ByIDArg != 999 {
		t.Errorf("argument ByID = %d; want 999", repo.ByIDArg)
	}

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("error is %v, want %v", err, ErrNotFound)
	}
}

func TestServiceByIDHappyPath(t *testing.T) {
	ctx := context.Background()
	repo := &FakeUserRepository{
		ByIDUser: User{
			ID:   1,
			Name: "Bob",
			Age:  35,
		},
	}
	service := NewService(repo)
	user, err := service.ByID(ctx, 1)

	if err != nil {
		t.Fatal("unexpected error")
	}
	if !repo.ByIDCalled {
		t.Fatal("repository ByID was not called")
	}
	if repo.ByIDArg != 1 {
		t.Errorf("argument id %d; want = 1", repo.ByIDArg)
	}
	if user.ID != 1 {
		t.Errorf("user id = %d; want id = 1", user.ID)
	}
	if user.Name != "Bob" {
		t.Errorf("user name = %q; want name = Bob", user.Name)
	}
	if user.Age != 35 {
		t.Errorf("user age = %d; want age = 35", user.Age)
	}
}

func TestServiceByIDPropagationError(t *testing.T) {
	errRepo := errors.New("repository error")
	ctx := context.Background()
	repo := &FakeUserRepository{
		ByIDErr: errRepo,
	}

	service := NewService(repo)

	_, err := service.ByID(ctx, 10)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, errRepo) {
		t.Errorf("error = %v; want wrapped %v", err, errRepo)
	}
}
