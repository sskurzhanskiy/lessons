package user

import (
	"context"
	"errors"
	"fmt"
	httpx "lessonHttp/internal/httpx"
	"net/http"
)

type ValidationError struct {
	Field string
}

func (e *ValidationError) Error() string {
	return "validation failed for field: " + e.Field
}

func writeError(w http.ResponseWriter, err error) {
	var validationErr *ValidationError
	switch {
	case errors.As(err, &validationErr):
		httpx.WriteJSON(w,
			http.StatusBadRequest,
			httpx.ErrorResponse{Error: validationErr.Error()},
		)
	case errors.Is(err, ErrNotFound):
		httpx.WriteJSON(w,
			http.StatusNotFound,
			httpx.ErrorResponse{Error: ErrNotFound.Error()},
		)
	default:
		httpx.WriteJSON(w,
			http.StatusInternalServerError,
			httpx.ErrorResponse{Error: "internal server error"},
		)
	}
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

	user := User{
		Name: name,
		Age:  age,
	}

	rUser, err := s.repo.Create(ctx, user)
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
