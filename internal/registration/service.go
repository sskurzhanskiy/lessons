package registration

import (
	"context"
	"fmt"

	"lessonHttp/internal/profile"
	"lessonHttp/internal/unitofwork"
	"lessonHttp/internal/user"
)

type ValidationError struct {
	Field string
}

func (v *ValidationError) Error() string {
	return "validation error for field: " + v.Field
}

type Service struct {
	tm unitofwork.Manager
}

func NewRegistrationService(tm unitofwork.Manager) *Service {
	return &Service{
		tm: tm,
	}
}

func (s *Service) CreateUserWithProfile(ctx context.Context, name string, age int, bio string) error {
	if name == "" {
		return &ValidationError{Field: "name"}
	}
	if age <= 0 {
		return &ValidationError{Field: "age"}
	}
	if bio == "" {
		return &ValidationError{Field: "bio"}
	}

	return s.tm.WithTransaction(ctx, func(uow *unitofwork.UnitOfWork) error {
		createdUser, err := uow.Users().Create(ctx, user.User{
			Name: name,
			Age:  age,
		})
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		err = uow.Profiles().Create(ctx, profile.Profile{
			UserID: createdUser.ID,
			Bio:    bio,
		})
		if err != nil {
			return fmt.Errorf("create profile: %w", err)
		}

		return nil
	})
}
