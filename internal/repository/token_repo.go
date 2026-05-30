package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ramisoul84/resreview-server/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type TokenRepository interface {
	Set(ctx context.Context, userID, jti, fingerprint string, ttl time.Duration) error
	Get(ctx context.Context, userID, jti string) (string, error)
	Delete(ctx context.Context, userID, jti string) error
}

type tokenRepository struct {
	client *redis.Client
	log    logger.Logger
}

func NewTokenRepository(client *redis.Client) TokenRepository {
	return &tokenRepository{client: client, log: logger.Get()}
}

func (r *tokenRepository) Set(ctx context.Context, userID, jti, fingerprint string, ttl time.Duration) error {
	if err := r.client.Set(ctx, refreshKey(userID, jti), fingerprint, ttl).Err(); err != nil {
		return fmt.Errorf("token_repo.Set: %w", err)
	}
	return nil
}

func (r *tokenRepository) Get(ctx context.Context, userID, jti string) (string, error) {
	return r.client.Get(ctx, refreshKey(userID, jti)).Result()
}

func (r *tokenRepository) Delete(ctx context.Context, userID, jti string) error {
	if err := r.client.Del(ctx, refreshKey(userID, jti)).Err(); err != nil {
		return fmt.Errorf("token_repo.DeleteRefreshToken: %w", err)
	}
	return nil
}

// Helper
func refreshKey(userID, jti string) string {
	return fmt.Sprintf("refresh:%s:%s", userID, jti)
}
