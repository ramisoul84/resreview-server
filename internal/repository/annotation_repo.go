package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/ramisoul84/resreview-server/internal/domain"
	"github.com/ramisoul84/resreview-server/pkg/logger"
)

type AnnotationRepository interface {
	ListByVersionID(ctx context.Context, versionID string) ([]domain.Annotation, error)
	Create(ctx context.Context, ann *domain.Annotation) error
	Update(ctx context.Context, ann *domain.Annotation) error
	Delete(ctx context.Context, id string) error
}

type annotationRepository struct {
	db  *sqlx.DB
	log logger.Logger
}

func NewAnnotationRepository(db *sqlx.DB) AnnotationRepository {
	return &annotationRepository{db: db, log: logger.Get()}
}

func (r *annotationRepository) ListByVersionID(ctx context.Context, versionID string) ([]domain.Annotation, error) {
	log := r.log.WithFields(map[string]any{
		"layer":      "annotation_repo",
		"method":     "ListByVersionID",
		"request_id": ctx.Value("request_id"),
	})

	query := `SELECT id, version_id, type, data::text, user_id, session_id, color, stroke_w, stroke_style, x, y, title, text, created_at, updated_at FROM annotations WHERE version_id = $1 ORDER BY created_at`

	var rows []domain.Annotation
	if err := r.db.SelectContext(ctx, &rows, query, versionID); err != nil {
		log.Error().Err(err).Msg("failed to list annotations")
		return nil, err
	}
	return rows, nil
}

func (r *annotationRepository) Create(ctx context.Context, ann *domain.Annotation) error {
	log := r.log.WithFields(map[string]any{
		"layer":      "annotation_repo",
		"method":     "Create",
		"request_id": ctx.Value("request_id"),
	})

	query := `INSERT INTO annotations (id, version_id, type, data, user_id, session_id, color, stroke_w, stroke_style, x, y, title, text, created_at, updated_at) VALUES ($1, $2, $3, $4::jsonb, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING id`

	var id string
	err := r.db.QueryRowContext(ctx, query,
		ann.ID, ann.VersionID, ann.Type, ann.Data, ann.UserID, ann.SessionID,
		ann.Color, ann.StrokeW, ann.StrokeStyle, ann.X, ann.Y, ann.Title, ann.Text,
		ann.CreatedAt, ann.UpdatedAt,
	).Scan(&id)
	if err != nil {
		log.Error().Err(err).Msg("failed to create annotation")
		return err
	}
	ann.ID = id
	return nil
}

func (r *annotationRepository) Update(ctx context.Context, ann *domain.Annotation) error {
	log := r.log.WithFields(map[string]any{
		"layer":      "annotation_repo",
		"method":     "Update",
		"request_id": ctx.Value("request_id"),
	})

	// First read existing to preserve fields not being updated
	var existing domain.Annotation
	err := r.db.GetContext(ctx, &existing, `SELECT id, version_id, type, data::text, user_id, session_id, color, stroke_w, stroke_style, x, y, title, text, created_at, updated_at FROM annotations WHERE id = $1`, ann.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrAnnotationNotFound
		}
		log.Error().Err(err).Msg("failed to fetch annotation for update")
		return err
	}

	data := ann.Data
	if data == "" {
		data = existing.Data
	}
	x := ann.X
	if x == 0 && ann.Data == "" {
		x = existing.X
	}
	y := ann.Y
	if y == 0 && ann.Data == "" {
		y = existing.Y
	}
	title := ann.Title
	if title == "" {
		title = existing.Title
	}
	text := ann.Text
	if text == "" {
		text = existing.Text
	}

	query := `UPDATE annotations SET data = $1::jsonb, x = $2, y = $3, title = $4, text = $5, updated_at = $6 WHERE id = $7`
	res, err := r.db.ExecContext(ctx, query, data, x, y, title, text, ann.UpdatedAt, ann.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to update annotation")
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return domain.ErrAnnotationNotFound
	}
	return nil
}

func (r *annotationRepository) Delete(ctx context.Context, id string) error {
	log := r.log.WithFields(map[string]any{
		"layer":      "annotation_repo",
		"method":     "Delete",
		"request_id": ctx.Value("request_id"),
	})

	query := `DELETE FROM annotations WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete annotation")
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return domain.ErrAnnotationNotFound
	}
	return nil
}

