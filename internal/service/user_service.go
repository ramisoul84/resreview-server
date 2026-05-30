package service

import (
	"context"

	"github.com/ramisoul84/resreview-server/internal/domain"
	"github.com/ramisoul84/resreview-server/internal/repository"
	"github.com/ramisoul84/resreview-server/pkg/logger"
)

type UserService interface {
	ListUsers(ctx context.Context) ([]domain.UserResponse, error)
	UpdateRole(ctx context.Context, userID string, isAdmin bool) error
	UpdateProfile(ctx context.Context, userID, name, color string) error
}

type userService struct {
	repo repository.UserRepository
	log  logger.Logger
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo, log: logger.Get()}
}

func (s *userService) ListUsers(ctx context.Context) ([]domain.UserResponse, error) {
	log := s.log.WithFields(map[string]any{
		"layer":      "user_service",
		"method":     "ListUsers",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Msg("listing all users")

	users, err := s.repo.List(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to list users")
		return nil, err
	}

	return users, nil
}

func (s *userService) UpdateRole(ctx context.Context, userID string, isAdmin bool) error {
	log := s.log.WithFields(map[string]any{
		"layer":      "user_service",
		"method":     "UpdateRole",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("user_id", userID).Bool("is_admin", isAdmin).Msg("updating user role")

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	return s.repo.UpdateIsAdmin(ctx, userID, isAdmin)
}

func (s *userService) UpdateProfile(ctx context.Context, userID, name, color string) error {
	log := s.log.WithFields(map[string]any{
		"layer":      "user_service",
		"method":     "UpdateProfile",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("user_id", userID).Str("color", color).Msg("updating user profile")

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	return s.repo.UpdateProfile(ctx, userID, name, color)
}
