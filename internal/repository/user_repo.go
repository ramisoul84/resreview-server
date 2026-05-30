package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/ramisoul84/resreview-server/internal/domain"
	"github.com/ramisoul84/resreview-server/pkg/logger"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)
	List(ctx context.Context) ([]domain.UserResponse, error)
	UpdateLastLoginAt(ctx context.Context, userID string) error
	UpdateIsAdmin(ctx context.Context, userID string, isAdmin bool) error
	UpdateProfile(ctx context.Context, userID, name, color string) error
}

type userRepository struct {
	db  *sqlx.DB
	log logger.Logger
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db, log: logger.Get()}
}

func (r *userRepository) CreateUser(ctx context.Context, user *domain.User) error {
	log := r.log.WithFields(map[string]any{
		"layer":      "user_repo",
		"method":     "CreateUser",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Msg("Creating New User")

	query := `
		INSERT INTO users (id, name, email, password_hash, color, is_admin, created_at, last_login_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (email) DO NOTHING
		RETURNING id
	`

	var id string
	err := r.db.QueryRowContext(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.Color,
		user.IsAdmin,
		user.CreatedAt,
		user.LastLoginAt,
	).Scan(&id)

	if err == sql.ErrNoRows {
		log.Debug().Msg("user already exists")
		return domain.ErrUserExists
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to create user")
		return err
	}

	log.Debug().Msg("user created successfully")
	return nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	log := r.log.WithFields(map[string]any{
		"layer":      "user_repo",
		"method":     "GetUserByEmail",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Msg("Fetching user by email")

	query := `SELECT id, name, email, password_hash, color, is_admin, created_at, last_login_at 
	          FROM users 
	          WHERE email = $1`

	var user domain.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug().Msg("user not found")
			return nil, domain.ErrUserNotFound
		}
		log.Error().Err(err).Msg("failed to fetch user")
		return nil, err
	}

	log.Debug().Msg("user fetched successfully")
	return &user, nil
}

func (r *userRepository) List(ctx context.Context) ([]domain.UserResponse, error) {
	log := r.log.WithFields(map[string]any{
		"layer":      "user_repo",
		"method":     "List",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Msg("Fetching all users")

	query := `SELECT id, name, email, color, is_admin, created_at, last_login_at 
	          FROM users 
	          ORDER BY created_at DESC`

	var users []domain.UserResponse
	err := r.db.SelectContext(ctx, &users, query)
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch users")
		return nil, err
	}

	log.Debug().Int("count", len(users)).Msg("users fetched successfully")
	return users, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	log := r.log.WithFields(map[string]any{
		"layer":      "user_repo",
		"method":     "GetUserByID",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Msg("Fetching user by ID")

	query := `SELECT id, name, email, password_hash, color, is_admin, created_at, last_login_at 
	          FROM users 
	          WHERE id = $1`

	var user domain.User
	err := r.db.GetContext(ctx, &user, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug().Msg("user not found")
			return nil, domain.ErrUserNotFound
		}
		log.Error().Err(err).Msg("failed to fetch user")
		return nil, err
	}

	log.Debug().Msg("user fetched successfully")
	return &user, nil
}

func (r *userRepository) UpdateIsAdmin(ctx context.Context, userID string, isAdmin bool) error {
	log := r.log.WithFields(map[string]any{
		"layer":      "user_repo",
		"method":     "UpdateIsAdmin",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Bool("is_admin", isAdmin).Msg("updating user admin status")

	query := `UPDATE users SET is_admin = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, isAdmin, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to update user admin status")
		return err
	}

	log.Debug().Msg("user admin status updated successfully")
	return nil
}

func (r *userRepository) UpdateProfile(ctx context.Context, userID, name, color string) error {
	log := r.log.WithFields(map[string]any{
		"layer":      "user_repo",
		"method":     "UpdateProfile",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("user_id", userID).Msg("updating user profile")

	query := `UPDATE users SET name = $1, color = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, name, color, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to update user profile")
		return err
	}

	log.Debug().Msg("user profile updated successfully")
	return nil
}

func (r *userRepository) UpdateLastLoginAt(ctx context.Context, userID string) error {
	log := r.log.WithFields(map[string]any{
		"layer":      "user_repo",
		"method":     "UpdateLastLoginAt",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Msg("updating last login at")

	query := `UPDATE users SET last_login_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to update last login at")
		return err
	}

	log.Debug().Msg("last login at updated successfully")
	return nil
}
