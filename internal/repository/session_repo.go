package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/ramisoul84/resreview-server/internal/domain"
	"github.com/ramisoul84/resreview-server/pkg/logger"
)

type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, sessionID string) (*domain.Session, error)
	ListByUserID(ctx context.Context, userID string) ([]domain.SessionResponse, error)
	Update(ctx context.Context, sessionID, userID, name string) error
	Delete(ctx context.Context, sessionID, userID string) error
}

type sessionRepository struct {
	db  *sqlx.DB
	log logger.Logger
}

func NewSessionRepository(db *sqlx.DB) SessionRepository {
	return &sessionRepository{db: db, log: logger.Get()}
}

func (r *sessionRepository) Create(ctx context.Context, session *domain.Session) error {
	log := r.log.WithFields(map[string]any{
		"layer":      "session_repo",
		"method":     "Create",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Msg("Creating new session")

	query := `
		INSERT INTO sessions (id, name, user_id, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id string
	err := r.db.QueryRowContext(ctx, query,
		session.ID,
		session.Name,
		session.UserID,
		session.CreatedAt,
	).Scan(&id)

	if err != nil {
		log.Error().Err(err).Msg("failed to create session")
		return err
	}

	log.Debug().Msg("session created successfully")
	return nil
}

func (r *sessionRepository) GetByID(ctx context.Context, sessionID string) (*domain.Session, error) {
	log := r.log.WithFields(map[string]any{
		"layer":      "session_repo",
		"method":     "GetByID",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("session_id", sessionID).Msg("Fetching session by ID")

	query := `SELECT id, name, user_id, created_at FROM sessions WHERE id = $1`

	var session domain.Session
	err := r.db.GetContext(ctx, &session, query, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug().Msg("session not found")
			return nil, domain.ErrSessionNotFound
		}
		log.Error().Err(err).Msg("failed to fetch session")
		return nil, err
	}

	log.Debug().Msg("session fetched successfully")
	return &session, nil
}

func (r *sessionRepository) ListByUserID(ctx context.Context, userID string) ([]domain.SessionResponse, error) {
	log := r.log.WithFields(map[string]any{
		"layer":      "session_repo",
		"method":     "ListByUserID",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("user_id", userID).Msg("Fetching sessions by user ID")

	query := `SELECT id, name, user_id, created_at FROM sessions WHERE user_id = $1 ORDER BY created_at DESC`

	var sessions []domain.SessionResponse
	err := r.db.SelectContext(ctx, &sessions, query, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch sessions")
		return nil, err
	}

	log.Debug().Int("count", len(sessions)).Msg("sessions fetched successfully")
	return sessions, nil
}

func (r *sessionRepository) Update(ctx context.Context, sessionID, userID, name string) error {
	log := r.log.WithFields(map[string]any{
		"layer":      "session_repo",
		"method":     "Update",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("session_id", sessionID).Msg("updating session")

	query := `UPDATE sessions SET name = $1 WHERE id = $2 AND user_id = $3`
	res, err := r.db.ExecContext(ctx, query, name, sessionID, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to update session")
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return domain.ErrSessionNotFound
	}

	log.Debug().Msg("session updated successfully")
	return nil
}

func (r *sessionRepository) Delete(ctx context.Context, sessionID, userID string) error {
	log := r.log.WithFields(map[string]any{
		"layer":      "session_repo",
		"method":     "Delete",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("session_id", sessionID).Msg("deleting session")

	query := `DELETE FROM sessions WHERE id = $1 AND user_id = $2`
	res, err := r.db.ExecContext(ctx, query, sessionID, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete session")
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return domain.ErrSessionNotFound
	}

	log.Debug().Msg("session deleted successfully")
	return nil
}
