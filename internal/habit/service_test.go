package habit

import (
	"context"
	"errors"
	"testing"
	"time"
)

type FakeHabitRepository struct {
	IsCreateCall bool
	Habit        Habit
	CreateErr    error
}

func (r *FakeHabitRepository) Create(ctx context.Context, param CreateHabit) (Habit, error) {
	r.IsCreateCall = true

	return r.Habit, r.CreateErr
}

func (r *FakeHabitRepository) ByID(ctx context.Context, userID int, habitID int) (Habit, error) {
	return Habit{}, nil
}

func (r *FakeHabitRepository) List(ctx context.Context, userID int, limit int, offset int) ([]Habit, error) {
	return []Habit{}, nil
}

func (r *FakeHabitRepository) Update(ctx context.Context, userID int, habitID int, update UpdateHabit) (Habit, error) {
	return Habit{}, nil
}

func (r *FakeHabitRepository) Delete(ctx context.Context, userID int, habitID int) error {
	return nil
}

var ErrRepo = errors.New("some error from repo")

func TestServiceCreate(t *testing.T) {
	scheduleAt := time.Date(2026, time.October, 5, 15, 0, 0, 0, time.UTC)
	var tests = []struct {
		name        string
		userID      int
		input       CreateInput
		isCallRepo  bool
		habit       Habit
		wantRepoErr error
		wantErr     error
	}{
		{
			name:    "user id invalid",
			wantErr: &ValidationError{Field: "userID"},
		},
		{
			name:    "input param title invalid",
			userID:  5,
			input:   CreateInput{},
			wantErr: &ValidationError{Field: "title"},
		},
		{
			name:   "repo error",
			userID: 5,
			input: CreateInput{
				Title:       "habit 2",
				ScheduledAt: time.Now(),
				Comment:     "comment",
			},
			isCallRepo:  true,
			wantRepoErr: ErrRepo,
			wantErr:     ErrRepo,
		},
		{
			name:   "correct",
			userID: 5,
			input: CreateInput{
				Title:       "habit 3",
				ScheduledAt: scheduleAt,
				Comment:     "comment",
			},
			isCallRepo: true,
			habit: Habit{
				ID:          1,
				UserID:      5,
				Title:       "habit 3",
				ScheduledAt: scheduleAt,
				Comment:     "comment",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := FakeHabitRepository{
				Habit:     tt.habit,
				CreateErr: tt.wantRepoErr,
			}
			service := NewService(&repo)
			habit, err := service.Create(context.Background(), tt.userID, tt.input)

			if repo.IsCreateCall != tt.isCallRepo {
				t.Fatalf("service create is call %t; want %t", repo.IsCreateCall, tt.isCallRepo)
			}

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected %v", tt.wantErr)
				}
				var validErr *ValidationError
				if errors.As(err, &validErr) {
					if validErr.Error() != tt.wantErr.Error() {
						t.Errorf("service create error %q; want %q", validErr.Error(), tt.wantErr.Error())
					}
				} else if !errors.Is(err, tt.wantErr) {
					t.Errorf("service create error %v; want %v", err, tt.wantErr)
				}
				return
			}

			if habit.ID == 0 {
				t.Errorf("got id 0; want %d", tt.habit.ID)
			}
			if habit.UserID != tt.habit.UserID {
				t.Errorf("got user id %d; want %d", habit.UserID, tt.habit.UserID)
			}
			if habit.Title != tt.habit.Title {
				t.Errorf("got title %q; want %q", habit.Title, tt.habit.Title)
			}
			if habit.ScheduledAt != tt.habit.ScheduledAt {
				t.Errorf("got scheduledAt %q; want %q", habit.ScheduledAt, tt.habit.ScheduledAt)
			}
			if habit.Comment != tt.habit.Comment {
				t.Errorf("got user id %q; want %q", habit.Comment, tt.habit.Comment)
			}
		})
	}
}
