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
	List(ctx context.Context, userID int, limit int, offset int) ([]Habit, error)
	ByID(ctx context.Context, userID int, habitID int) (Habit, error)
	Update(ctx context.Context, userID int, habitID int, input UpdateInput) (Habit, error)
	Delete(ctx context.Context, userID int, habitID int) error
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

func (s *Service) List(ctx context.Context, userID int, limit int, offset int) ([]Habit, error) {
	if userID <= 0 {
		return nil, &ValidationError{Field: "userID"}
	}

	return s.repo.List(ctx, userID, limit, offset)
}

func (s *Service) ByID(ctx context.Context, userID int, habitID int) (Habit, error) {
	if userID <= 0 {
		return Habit{}, &ValidationError{Field: "userID"}
	}
	if habitID <= 0 {
		return Habit{}, &ValidationError{Field: "ID"}
	}

	habit, err := s.repo.ByID(ctx, userID, habitID)
	if err != nil {
		return Habit{}, fmt.Errorf("get habit by id: %w", err)
	}

	return habit, nil
}

func (s *Service) Update(ctx context.Context, userID int, habitID int, input UpdateInput) (Habit, error) {
	if userID <= 0 {
		return Habit{}, &ValidationError{Field: "userID"}
	}
	if habitID <= 0 {
		return Habit{}, &ValidationError{Field: "ID"}
	}
	if input.Title != nil && strings.TrimSpace(*input.Title) == "" {
		return Habit{}, &ValidationError{Field: "title"}
	}

	if input.Title == nil &&
		input.ScheduledAt == nil &&
		input.IsComplete == nil &&
		input.Comment == nil {
		return Habit{}, &ValidationError{Field: "update"}
	}

	update := UpdateHabit{
		Title:       input.Title,
		ScheduledAt: input.ScheduledAt,
		IsComplete:  input.IsComplete,
		Comment:     input.Comment,
	}
	habit, err := s.repo.Update(ctx, userID, habitID, update)
	if err != nil {
		return Habit{}, fmt.Errorf("update habit %d %w", habitID, err)
	}

	return habit, nil
}

func (s *Service) Delete(ctx context.Context, userID int, habitID int) error {
	if userID <= 0 {
		return &ValidationError{Field: "userID"}
	}
	if habitID <= 0 {
		return &ValidationError{Field: "habitID"}
	}

	return s.repo.Delete(ctx, userID, habitID)
}
