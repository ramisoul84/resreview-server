package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ramisoul84/resreview-server/internal/config"
	"github.com/ramisoul84/resreview-server/internal/domain"
	"github.com/ramisoul84/resreview-server/internal/repository"
	"github.com/ramisoul84/resreview-server/pkg/logger"
)

type accessClaims struct {
	jwt.RegisteredClaims
	IsAdmin bool `json:"is_admin"`
}

type refreshClaims struct {
	jwt.RegisteredClaims
}

type TokenService interface {
	IssueAccessToken(ctx context.Context, userID string, IsAdmin bool) (string, error)
	ValidateAccessToken(ctx context.Context, token string) (*accessClaims, error)
	IssueRefreshToken(ctx context.Context, userID, fingerprint string) (string, error)
	ValidateRefreshToken(ctx context.Context, tokenStr, fingerprint string) (*refreshClaims, error)
	RevokeRefreshToken(ctx context.Context, userID, jti string) error
}

type tokenService struct {
	repo repository.TokenRepository
	cfg  *config.Config
	log  logger.Logger
}

func NewTokenService(repo repository.TokenRepository, cfg *config.Config) *tokenService {
	return &tokenService{repo: repo, cfg: cfg, log: logger.Get()}
}

func Fingerprint(userAgent string) string {
	h := sha256.Sum256([]byte(userAgent))
	return hex.EncodeToString(h[:])
}

func (s *tokenService) IssueAccessToken(ctx context.Context, userID string, IsAdmin bool) (string, error) {
	now := time.Now()

	accessExpiry := now.Add(s.cfg.JWT.AccessTokenExpiry)
	accessStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.cfg.JWT.Issuer,
		},
		IsAdmin: IsAdmin,
	}).SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return "", fmt.Errorf("token_service.IssueAccess: sign: %w", err)
	}

	return accessStr, nil
}

func (s *tokenService) ValidateAccessToken(ctx context.Context, tokenStr string) (*accessClaims, error) {
	claims := &accessClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, s.keyFunc)
	if err != nil {
		return nil, mapJWTError(err)
	}
	if !token.Valid {
		return nil, domain.ErrTokenInvalid
	}
	return claims, nil
}

func (s *tokenService) keyFunc(token *jwt.Token) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return []byte(s.cfg.JWT.Secret), nil
}

func mapJWTError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, jwt.ErrTokenExpired) {
		return domain.ErrTokenExpired
	}
	if strings.Contains(err.Error(), "expired") {
		return domain.ErrTokenExpired
	}
	return domain.ErrTokenInvalid
}

func (s *tokenService) IssueRefreshToken(ctx context.Context, userID, fingerprint string) (string, error) {
	log := s.log.WithFields(map[string]any{
		"layer":      "token_service",
		"method":     "IssueRefreshToken",
		"request_id": ctx.Value("request_id"),
	})

	now := time.Now()
	jti := uuid.NewString()
	expiry := now.Add(s.cfg.JWT.RefreshTokenExpiry)

	refreshStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiry),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.cfg.JWT.Issuer,
		},
	}).SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return "", fmt.Errorf("token_service.IssueRefresh: sign: %w", err)
	}

	if err := s.repo.Set(ctx, userID, jti, fingerprint, s.cfg.JWT.RefreshTokenExpiry); err != nil {
		log.Error().Err(err).Msg("failed to store refresh token in redis")
		return "", fmt.Errorf("token_service.IssueRefresh: store: %w", err)
	}

	log.Debug().Str("jti", jti).Msg("refresh token issued")
	return refreshStr, nil
}

func (s *tokenService) ValidateRefreshToken(ctx context.Context, tokenStr, fingerprint string) (*refreshClaims, error) {
	log := s.log.WithFields(map[string]any{
		"layer":      "token_service",
		"method":     "ValidateRefreshToken",
		"request_id": ctx.Value("request_id"),
	})

	claims := &refreshClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, s.keyFunc)
	if err != nil {
		return nil, mapJWTError(err)
	}
	if !token.Valid {
		return nil, domain.ErrTokenInvalid
	}

	stored, err := s.repo.Get(ctx, claims.Subject, claims.ID)
	if err != nil {
		log.Debug().Str("jti", claims.ID).Err(err).Msg("refresh token not found in redis")
		return nil, domain.ErrTokenNotFound
	}

	if stored != fingerprint {
		log.Warn().Str("jti", claims.ID).Msg("fingerprint mismatch — possible token theft")
		_ = s.repo.Delete(ctx, claims.Subject, claims.ID)
		return nil, domain.ErrTokenRevoked
	}

	return claims, nil
}

func (s *tokenService) RevokeRefreshToken(ctx context.Context, userID, jti string) error {
	log := s.log.WithFields(map[string]any{
		"layer":      "token_service",
		"method":     "RevokeRefreshToken",
		"request_id": ctx.Value("request_id"),
	})
	log.Debug().Str("jti", jti).Msg("revoking refresh token")
	return s.repo.Delete(ctx, userID, jti)
}
