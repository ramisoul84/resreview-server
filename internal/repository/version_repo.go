package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/ramisoul84/resreview-server/internal/domain"
	"github.com/ramisoul84/resreview-server/pkg/logger"
)

type VersionRepository interface {
	Create(ctx context.Context, version *domain.Version) error
	GetByID(ctx context.Context, versionID string) (*domain.Version, error)
	ListAll(ctx context.Context) ([]domain.VersionResponse, error)
	ListByProductID(ctx context.Context, productID string) ([]domain.VersionResponse, error)
	Update(ctx context.Context, versionID, label, name string) error
	UpdateURL(ctx context.Context, versionID, url string) error
	Delete(ctx context.Context, versionID string) error
}

type versionRepository struct {
	db  *sqlx.DB
	log logger.Logger
}

func NewVersionRepository(db *sqlx.DB) VersionRepository {
	return &versionRepository{db: db, log: logger.Get()}
}

func (r *versionRepository) Create(ctx context.Context, version *domain.Version) error {
	log := r.log.WithFields(map[string]any{
		"layer":      "version_repo",
		"method":     "Create",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Msg("Creating new version")

	query := `
		INSERT INTO versions (id, label, name, product_id, url, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id string
	err := r.db.QueryRowContext(ctx, query,
		version.ID,
		version.Label,
		version.Name,
		version.ProductID,
		version.URL,
		version.CreatedAt,
	).Scan(&id)

	if err != nil {
		log.Error().Err(err).Msg("failed to create version")
		return err
	}

	log.Debug().Msg("version created successfully")
	return nil
}

func (r *versionRepository) GetByID(ctx context.Context, versionID string) (*domain.Version, error) {
	log := r.log.WithFields(map[string]any{
		"layer":      "version_repo",
		"method":     "GetByID",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("version_id", versionID).Msg("Fetching version by ID")

	query := `SELECT id, label, name, product_id, url, created_at FROM versions WHERE id = $1`

	var version domain.Version
	err := r.db.GetContext(ctx, &version, query, versionID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug().Msg("version not found")
			return nil, domain.ErrVersionNotFound
		}
		log.Error().Err(err).Msg("failed to fetch version")
		return nil, err
	}

	log.Debug().Msg("version fetched successfully")
	return &version, nil
}

func (r *versionRepository) ListByProductID(ctx context.Context, productID string) ([]domain.VersionResponse, error) {
	log := r.log.WithFields(map[string]any{
		"layer":      "version_repo",
		"method":     "ListByProductID",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("product_id", productID).Msg("Fetching versions by product ID")

	query := `SELECT id, label, name, product_id, url, created_at FROM versions WHERE product_id = $1 ORDER BY created_at DESC`

	var versions []domain.VersionResponse
	err := r.db.SelectContext(ctx, &versions, query, productID)
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch versions")
		return nil, err
	}

	log.Debug().Int("count", len(versions)).Msg("versions fetched successfully")
	return versions, nil
}

func (r *versionRepository) ListAll(ctx context.Context) ([]domain.VersionResponse, error) {
	log := r.log.WithFields(map[string]any{
		"layer":      "version_repo",
		"method":     "ListAll",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Msg("Fetching all versions")

	query := `SELECT id, label, name, product_id, url, created_at FROM versions ORDER BY created_at DESC`

	var versions []domain.VersionResponse
	err := r.db.SelectContext(ctx, &versions, query)
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch all versions")
		return nil, err
	}

	log.Debug().Int("count", len(versions)).Msg("all versions fetched successfully")
	return versions, nil
}

func (r *versionRepository) Update(ctx context.Context, versionID, label, name string) error {
	log := r.log.WithFields(map[string]any{
		"layer":      "version_repo",
		"method":     "Update",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("version_id", versionID).Msg("updating version")

	query := `UPDATE versions SET label = $1, name = $2 WHERE id = $3`
	res, err := r.db.ExecContext(ctx, query, label, name, versionID)
	if err != nil {
		log.Error().Err(err).Msg("failed to update version")
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return domain.ErrVersionNotFound
	}

	log.Debug().Msg("version updated successfully")
	return nil
}

func (r *versionRepository) UpdateURL(ctx context.Context, versionID, url string) error {
	log := r.log.WithFields(map[string]any{
		"layer":      "version_repo",
		"method":     "UpdateURL",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("version_id", versionID).Msg("updating version url")

	query := `UPDATE versions SET url = $1 WHERE id = $2`
	res, err := r.db.ExecContext(ctx, query, url, versionID)
	if err != nil {
		log.Error().Err(err).Msg("failed to update version url")
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return domain.ErrVersionNotFound
	}

	log.Debug().Msg("version url updated successfully")
	return nil
}

func (r *versionRepository) Delete(ctx context.Context, versionID string) error {
	log := r.log.WithFields(map[string]any{
		"layer":      "version_repo",
		"method":     "Delete",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("version_id", versionID).Msg("deleting version")

	query := `DELETE FROM versions WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, versionID)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete version")
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return domain.ErrVersionNotFound
	}

	log.Debug().Msg("version deleted successfully")
	return nil
}
