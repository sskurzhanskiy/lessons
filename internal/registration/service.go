package registration

import (
	"context"
	"fmt"

	"lessonHttp/internal/profile"
	"lessonHttp/internal/unitofwork"
	"lessonHttp/internal/user"
)

type Service struct {
	tm *unitofwork.TransactionManager
}

func NewRegistrationService(tm *unitofwork.TransactionManager) *Service {
	return &Service{
		tm: tm,
	}
}

func (s *Service) CreateUserWithProfile(ctx context.Context, name string, age int, bio string) error {
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
