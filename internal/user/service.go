package user

import (
	"context"
	"fmt"
)

type ServiceInterface interface {
	Create(ctx context.Context, name string, age int) (User, error)
	ByID(ctx context.Context, id int) (User, error)
	List(ctx context.Context, limit int, offset int) ([]User, error)
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

func (s *Service) List(ctx context.Context, limit int, offset int) ([]User, error) {
	return s.repo.List(ctx, limit, offset)
}
