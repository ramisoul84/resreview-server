package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/ramisoul84/resreview-server/internal/domain"
)

type PhotoRepository interface {
	FindByID(ctx context.Context, photoID string) (*domain.Photo, error)
	FindByVersionID(ctx context.Context, versionID string) (*domain.Photo, error)
	Create(ctx context.Context, photo domain.Photo) (*domain.Photo, error)
	Delete(ctx context.Context, photoID, versionID string) error
}

type photoRepository struct {
	db *sqlx.DB
}

func NewPhotoRepository(db *sqlx.DB) PhotoRepository {
	return &photoRepository{db: db}
}

func (r *photoRepository) FindByID(ctx context.Context, photoID string) (*domain.Photo, error) {
	var row domain.Photo
	err := r.db.GetContext(ctx, &row, `
		SELECT photo_id, version_id, url, key,  created_at
		FROM photos
		WHERE photo_id = $1
	`, photoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPhotoNotFound
		}
		return nil, fmt.Errorf("photo_repo.FindByID: %w", err)
	}
	return &row, nil
}

func (r *photoRepository) FindByVersionID(ctx context.Context, versionID string) (*domain.Photo, error) {
	var row domain.Photo
	err := r.db.GetContext(ctx, &row, `
		SELECT photo_id, version_id, url, key,  created_at
		FROM photos
		WHERE version_id = $1
	`, versionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPhotoNotFound
		}
		return nil, fmt.Errorf("photo_repo.FindByVersionID: %w", err)
	}
	return &row, nil
}

func (r *photoRepository) Create(ctx context.Context, photo domain.Photo) (*domain.Photo, error) {
	var row domain.Photo
	err := r.db.GetContext(ctx, &row, `
		INSERT INTO photos (version_id, url, key)
		VALUES ($1, $2, $3)
		RETURNING photo_id, version_id, url, key, created_at
	`, photo.PhotoID, photo.VersionID, photo.URL, photo.Key, photo.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("photo_repo.Insert: %w", err)
	}
	return &row, nil
}

func (r *photoRepository) Delete(ctx context.Context, photoID, versionID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM photos WHERE photo_id = $1 AND user_id = $2`, photoID, versionID)
	if err != nil {
		return fmt.Errorf("photo_repo.Delete: %w", err)
	}
	return nil
}
