package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/ramisoul84/resreview-server/internal/domain"
	"github.com/ramisoul84/resreview-server/pkg/logger"
)

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, productID string) (*domain.Product, error)
	ListByUserID(ctx context.Context, userID string) ([]domain.ProductResponse, error)
	Update(ctx context.Context, productID, userID, name string) error
	Delete(ctx context.Context, productID, userID string) error
}

type productRepository struct {
	db  *sqlx.DB
	log logger.Logger
}

func NewProductRepository(db *sqlx.DB) ProductRepository {
	return &productRepository{db: db, log: logger.Get()}
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	log := r.log.WithFields(map[string]any{
		"layer":      "product_repo",
		"method":     "Create",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Msg("Creating new product")

	query := `
		INSERT INTO products (id, name, user_id, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id string
	err := r.db.QueryRowContext(ctx, query,
		product.ID,
		product.Name,
		product.UserID,
		product.CreatedAt,
	).Scan(&id)

	if err != nil {
		log.Error().Err(err).Msg("failed to create product")
		return err
	}

	log.Debug().Msg("product created successfully")
	return nil
}

func (r *productRepository) GetByID(ctx context.Context, productID string) (*domain.Product, error) {
	log := r.log.WithFields(map[string]any{
		"layer":      "product_repo",
		"method":     "GetByID",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("product_id", productID).Msg("Fetching product by ID")

	query := `SELECT id, name, user_id, created_at FROM products WHERE id = $1`

	var product domain.Product
	err := r.db.GetContext(ctx, &product, query, productID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug().Msg("product not found")
			return nil, domain.ErrProductNotFound
		}
		log.Error().Err(err).Msg("failed to fetch product")
		return nil, err
	}

	log.Debug().Msg("product fetched successfully")
	return &product, nil
}

func (r *productRepository) ListByUserID(ctx context.Context, userID string) ([]domain.ProductResponse, error) {
	log := r.log.WithFields(map[string]any{
		"layer":      "product_repo",
		"method":     "ListByUserID",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("user_id", userID).Msg("Fetching products by user ID")

	query := `SELECT id, name, user_id, created_at FROM products WHERE user_id = $1 ORDER BY created_at DESC`

	var products []domain.ProductResponse
	err := r.db.SelectContext(ctx, &products, query, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch products")
		return nil, err
	}

	log.Debug().Int("count", len(products)).Msg("products fetched successfully")
	return products, nil
}

func (r *productRepository) Update(ctx context.Context, productID, userID, name string) error {
	log := r.log.WithFields(map[string]any{
		"layer":      "product_repo",
		"method":     "Update",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("product_id", productID).Msg("updating product")

	query := `UPDATE products SET name = $1 WHERE id = $2 AND user_id = $3`
	res, err := r.db.ExecContext(ctx, query, name, productID, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to update product")
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return domain.ErrProductNotFound
	}

	log.Debug().Msg("product updated successfully")
	return nil
}

func (r *productRepository) Delete(ctx context.Context, productID, userID string) error {
	log := r.log.WithFields(map[string]any{
		"layer":      "product_repo",
		"method":     "Delete",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("product_id", productID).Msg("deleting product and its versions")

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to begin transaction")
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM versions WHERE product_id = $1`, productID)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete product versions")
		return err
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM products WHERE id = $1 AND user_id = $2`, productID, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete product")
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return domain.ErrProductNotFound
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("failed to commit transaction")
		return err
	}

	log.Debug().Msg("product and its versions deleted successfully")
	return nil
}
