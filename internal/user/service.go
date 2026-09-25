package user

import (
	"context"
	"errors"
	"fmt"
	"lessonHttp/internal/password"
	pass "lessonHttp/internal/password"
	"strings"
)

type ServiceInterface interface {
	Create(ctx context.Context, name string, age int) (User, error)
	ByID(ctx context.Context, id int) (User, error)
	List(ctx context.Context, limit int, offset int) ([]User, error)
	Register(ctx context.Context, input RegisterInput) (User, error)
	Login(ctx context.Context, email string, password string) (int, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, name string, age int) (User, error) {
	if name == "" {
		return User{}, &ValidationError{Field: "name"}
	}
	if age <= 0 {
		return User{}, &ValidationError{Field: "age"}
	}

	params := CreateUserParams{
		Name: name,
		Age:  age,
	}

	rUser, err := s.repo.Create(ctx, params)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}

	return rUser, nil
}

func (s *Service) ByID(ctx context.Context, id int) (User, error) {
	if id <= 0 {
		return User{}, &ValidationError{Field: "id"}
	}

	user, err := s.repo.ByID(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("get user %d: %w", id, err)
	}

	return user, nil
}

func (s *Service) List(ctx context.Context, limit int, offset int) ([]User, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (User, error) {
	if input.Name == "" {
		return User{}, &ValidationError{Field: "name"}
	}
	if input.Email == "" {
		return User{}, &ValidationError{Field: "email"}
	}
	if input.Password == "" {
		return User{}, &ValidationError{Field: "password"}
	}
	if input.Age <= 0 {
		return User{}, &ValidationError{Field: "age"}
	}

	passwordHash, err := password.Hash(input.Password)
	if err != nil {
		return User{}, err
	}

	return s.repo.Create(ctx, CreateUserParams{
		Name:         input.Name,
		Age:          input.Age,
		Email:        input.Email,
		PasswordHash: passwordHash,
	})
}

func (s *Service) Login(ctx context.Context, email string, password string) (int, error) {
	if strings.TrimSpace(email) == "" {
		return 0, ErrInvalidCredentials
	}
	if strings.TrimSpace(password) == "" {
		return 0, ErrInvalidCredentials
	}

	output, err := s.repo.ByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return 0, ErrInvalidCredentials
		}
		return 0, fmt.Errorf("get user %q: %w", email, err)
	}

	ok, err := pass.Verify(password, output.PasswordHash)
	if err != nil {
		return 0, err
	}

	if !ok {
		return 0, ErrInvalidCredentials
	}

	return output.UserID, nil
}
