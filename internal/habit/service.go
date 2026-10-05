package habit

import (
	"context"
	"fmt"
	"strings"
)

type ValidationError struct {
	Field string
}

func (e *ValidationError) Error() string {
	return "validation failed for field: " + e.Field
}

type ServiceInterface interface {
	Create(ctx context.Context, userID int, input CreateInput) (Habit, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID int, input CreateInput) (Habit, error) {
	if userID <= 0 {
		return Habit{}, &ValidationError{Field: "userID"}
	}
	if strings.TrimSpace(input.Title) == "" {
		return Habit{}, &ValidationError{Field: "title"}
	}

	var createParam = CreateHabit{
		UserID:      userID,
		Title:       input.Title,
		ScheduledAt: input.ScheduledAt,
		Comment:     input.Comment,
	}
	habit, err := s.repo.Create(ctx, createParam)
	if err != nil {
		return Habit{}, fmt.Errorf("create habit: %w", err)
	}

	return habit, nil
}
