package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ramisoul84/resreview-server/internal/domain"
	"github.com/ramisoul84/resreview-server/internal/repository"
	"github.com/ramisoul84/resreview-server/pkg/logger"
)

type SessionService interface {
	CreateSession(ctx context.Context, userID, name string) (*domain.SessionResponse, error)
	GetSession(ctx context.Context, sessionID, userID string) (*domain.SessionResponse, error)
	ListSessions(ctx context.Context, userID string) ([]domain.SessionResponse, error)
	UpdateSession(ctx context.Context, sessionID, userID, name string) error
	DeleteSession(ctx context.Context, sessionID, userID string) error
}

type sessionService struct {
	repo repository.SessionRepository
	log  logger.Logger
}

func NewSessionService(repo repository.SessionRepository) SessionService {
	return &sessionService{repo: repo, log: logger.Get()}
}

func (s *sessionService) CreateSession(ctx context.Context, userID, name string) (*domain.SessionResponse, error) {
	log := s.log.WithFields(map[string]any{
		"layer":      "session_service",
		"method":     "CreateSession",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("user_id", userID).Msg("creating session")

	session := &domain.Session{
		ID:        uuid.New().String(),
		Name:      name,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, session); err != nil {
		log.Error().Err(err).Msg("failed to create session")
		return nil, err
	}

	return &domain.SessionResponse{
		ID:        session.ID,
		Name:      session.Name,
		UserID:    session.UserID,
		CreatedAt: session.CreatedAt,
	}, nil
}

func (s *sessionService) GetSession(ctx context.Context, sessionID, userID string) (*domain.SessionResponse, error) {
	log := s.log.WithFields(map[string]any{
		"layer":      "session_service",
		"method":     "GetSession",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("session_id", sessionID).Msg("getting session")

	session, err := s.repo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.UserID != userID {
		return nil, domain.ErrSessionNotFound
	}

	return &domain.SessionResponse{
		ID:        session.ID,
		Name:      session.Name,
		UserID:    session.UserID,
		CreatedAt: session.CreatedAt,
	}, nil
}

func (s *sessionService) ListSessions(ctx context.Context, userID string) ([]domain.SessionResponse, error) {
	log := s.log.WithFields(map[string]any{
		"layer":      "session_service",
		"method":     "ListSessions",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("user_id", userID).Msg("listing sessions")

	sessions, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list sessions")
		return nil, err
	}

	return sessions, nil
}

func (s *sessionService) UpdateSession(ctx context.Context, sessionID, userID, name string) error {
	log := s.log.WithFields(map[string]any{
		"layer":      "session_service",
		"method":     "UpdateSession",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("session_id", sessionID).Msg("updating session")

	return s.repo.Update(ctx, sessionID, userID, name)
}

func (s *sessionService) DeleteSession(ctx context.Context, sessionID, userID string) error {
	log := s.log.WithFields(map[string]any{
		"layer":      "session_service",
		"method":     "DeleteSession",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("session_id", sessionID).Msg("deleting session")

	return s.repo.Delete(ctx, sessionID, userID)
}
