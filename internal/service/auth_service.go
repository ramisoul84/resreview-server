package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ramisoul84/resreview-server/internal/domain"
	"github.com/ramisoul84/resreview-server/internal/repository"
	"github.com/ramisoul84/resreview-server/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, req *domain.RegisterRequest) error
	Login(ctx context.Context, req *domain.LoginRequest, fingerprint string) (*domain.LoginResponse, error)
	RefreshSession(ctx context.Context, refreshTokenStr, fingerprint string) (*domain.LoginResponse, error)
}

type authService struct {
	repo repository.UserRepository
	svc  TokenService
	log  logger.Logger
}

func NewAuthService(repo repository.UserRepository, svc TokenService) AuthService {
	return &authService{repo: repo, svc: svc, log: logger.Get()}
}

func (s *authService) Register(ctx context.Context, req *domain.RegisterRequest) error {
	log := s.log.WithFields(map[string]any{
		"layer":      "auth_service",
		"method":     "Register",
		"request_id": ctx.Value("request_id"),
	})
	log.Debug().Msg("Registering New User")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("failed to hash password")
		return err
	}

	user := domain.User{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		PasswordHash: string(hashedPassword),
		Color:        "#db072b",
		IsAdmin:      false,
		CreatedAt:    time.Now(),
		LastLoginAt:  nil,
	}

	return s.repo.CreateUser(ctx, &user)
}

func (s *authService) Login(ctx context.Context, req *domain.LoginRequest, fingerprint string) (*domain.LoginResponse, error) {
	log := s.log.WithFields(map[string]any{
		"layer":      "auth_service",
		"method":     "Login",
		"request_id": ctx.Value("request_id"),
	})
	log.Debug().Msg("Logining User")

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Debug().Err(err).Msg("failed to fetch user")
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		log.Debug().Err(err).Msg("invalid credentials")
		return nil, domain.ErrInvalidCredentials
	}

	accessToken, err := s.svc.IssueAccessToken(ctx, user.ID, user.IsAdmin)
	if err != nil {
		log.Error().Err(err).Msg("failed to issue access token")
		return nil, err
	}

	refreshToken, err := s.svc.IssueRefreshToken(ctx, user.ID, fingerprint)
	if err != nil {
		log.Error().Err(err).Msg("failed to issue refresh token")
		return nil, err
	}

	if err := s.repo.UpdateLastLoginAt(ctx, user.ID); err != nil {
		log.Error().Err(err).Msg("failed to update last login at")
		return nil, err
	}

	log.Debug().Msg("user logged in successfully")

	resp := domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		Color:        user.Color,
		IsAdmin:      user.IsAdmin,
	}

	return &resp, nil
}

func (s *authService) RefreshSession(ctx context.Context, refreshTokenStr, fingerprint string) (*domain.LoginResponse, error) {
	log := s.log.WithFields(map[string]any{
		"layer":      "auth_service",
		"method":     "RefreshSession",
		"request_id": ctx.Value("request_id"),
	})
	log.Debug().Msg("refreshing session")

	claims, err := s.svc.ValidateRefreshToken(ctx, refreshTokenStr, fingerprint)
	if err != nil {
		return nil, err
	}

	userID := claims.Subject

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("user not found during refresh")
		return nil, err
	}

	if err := s.svc.RevokeRefreshToken(ctx, userID, claims.ID); err != nil {
		log.Error().Err(err).Msg("failed to revoke old refresh token")
		return nil, err
	}

	accessToken, err := s.svc.IssueAccessToken(ctx, user.ID, user.IsAdmin)
	if err != nil {
		log.Error().Err(err).Msg("failed to issue new access token")
		return nil, err
	}

	newRefreshToken, err := s.svc.IssueRefreshToken(ctx, user.ID, fingerprint)
	if err != nil {
		log.Error().Err(err).Msg("failed to issue new refresh token")
		return nil, err
	}

	log.Debug().Msg("session refreshed successfully")

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		Color:        user.Color,
		IsAdmin:      user.IsAdmin,
	}, nil
}
