package handler

import (
	"context"
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/ramisoul84/resreview-server/internal/config"
	"github.com/ramisoul84/resreview-server/internal/domain"
	"github.com/ramisoul84/resreview-server/internal/server/http/middleware"
	"github.com/ramisoul84/resreview-server/internal/service"
	"github.com/ramisoul84/resreview-server/pkg/logger"
)

const refreshCookieName = "refresh_token"

type AuthHandler interface {
	Register(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	Refresh(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
}

type authHandler struct {
	service  service.AuthService
	tokenSvc service.TokenService
	validate *validator.Validate
	cfg      *config.Config
	log      logger.Logger
}

func NewAuthHandler(svc service.AuthService, tokenSvc service.TokenService, cfg *config.Config) AuthHandler {
	return &authHandler{
		service:  svc,
		tokenSvc: tokenSvc,
		validate: validator.New(),
		cfg:      cfg,
		log:      logger.Get(),
	}
}

// POST /api/v1/auth/register
func (h *authHandler) Register(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "auth_handler",
		"method": "Register",
	})

	log.Debug().Msg("Register request received")

	var req domain.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("invalid register payload")
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		log.Debug().Err(err).Msg("register payload validation failed")
		return validationError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	err := h.service.Register(ctx, &req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserExists):
			log.Debug().Msg("user already exists")
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "user with this email already exists",
			})
		case errors.Is(err, context.DeadlineExceeded):
			log.Error().Err(err).Msg("request timeout")
			return c.Status(fiber.StatusGatewayTimeout).JSON(fiber.Map{
				"error": "request timeout",
			})
		default:
			log.Error().Err(err).Msg("unexpected error during registration")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "registration failed",
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "user registered successfully",
	})
}

// POST /api/v1/auth/login
func (h *authHandler) Login(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "auth_handler",
		"method": "Login",
	})

	log.Debug().Msg("Login request received")

	var req domain.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("invalid login payload")
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		log.Debug().Err(err).Msg("login payload validation failed")
		return validationError(c, err)
	}

	ctx := ctxWithTimeout(c, 30*time.Second)
	defer ctx.cancel()

	fingerprint := service.Fingerprint(c.Get(fiber.HeaderUserAgent))

	resp, err := h.service.Login(ctx.ctx, &req, fingerprint)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials), errors.Is(err, domain.ErrUserNotFound):
			log.Debug().Err(err).Msg("invalid credentials")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid credentials",
			})
		default:
			log.Error().Err(err).Msg("unexpected error during login")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "login failed",
			})
		}
	}

	h.setNoStore(c)
	h.setRefreshCookie(c, resp.RefreshToken, fingerprint)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "user logged in successfully",
		"access_token": resp.AccessToken,
		"id":           resp.ID,
		"name":         resp.Name,
		"email":        resp.Email,
		"color":        resp.Color,
		"is_admin":     resp.IsAdmin,
	})
}

// POST /api/v1/auth/refresh
func (h *authHandler) Refresh(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "auth_handler",
		"method": "Refresh",
	})

	log.Debug().Msg("Refresh request received")

	refreshToken := c.Cookies(refreshCookieName)
	if refreshToken == "" {
		log.Debug().Msg("no refresh token cookie")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "refresh token missing",
		})
	}

	ctx := ctxWithTimeout(c, 10*time.Second)
	defer ctx.cancel()

	fingerprint := service.Fingerprint(c.Get(fiber.HeaderUserAgent))

	resp, err := h.service.RefreshSession(ctx.ctx, refreshToken, fingerprint)
	if err != nil {
		log.Debug().Err(err).Msg("refresh failed")
		h.clearRefreshCookie(c)
		switch {
		case errors.Is(err, domain.ErrTokenExpired):
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "refresh token expired, please login again",
			})
		case errors.Is(err, domain.ErrTokenRevoked):
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "refresh token revoked, please login again",
			})
		case errors.Is(err, domain.ErrTokenNotFound):
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "refresh token not found, please login again",
			})
		default:
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid refresh token",
			})
		}
	}

	h.setNoStore(c)
	h.setRefreshCookie(c, resp.RefreshToken, fingerprint)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token": resp.AccessToken,
	})
}

// POST /api/v1/auth/logout
func (h *authHandler) Logout(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "auth_handler",
		"method": "Logout",
	})

	log.Debug().Msg("Logout request received")

	refreshToken := c.Cookies(refreshCookieName)
	if refreshToken == "" {
		h.clearRefreshCookie(c)
		return c.JSON(fiber.Map{"message": "logged out"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	claims, err := h.tokenSvc.ValidateRefreshToken(ctx, refreshToken, "")
	if err == nil {
		_ = h.tokenSvc.RevokeRefreshToken(ctx, claims.Subject, claims.ID)
	}

	h.clearRefreshCookie(c)

	return c.JSON(fiber.Map{"message": "logged out successfully"})
}

// ── Helper ────────────────────────────────────────────────────────────────

func (h *authHandler) setRefreshCookie(c *fiber.Ctx, token, fingerprint string) {
	expiresAt := time.Now().Add(h.cfg.JWT.RefreshTokenExpiry)

	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/api/v1/auth",
		HTTPOnly: true,
		Secure:   h.cfg.IsProduction(),
		SameSite: "Lax",
		Expires:  expiresAt,
		MaxAge:   int(h.cfg.JWT.RefreshTokenExpiry.Seconds()),
	})
}

func (h *authHandler) clearRefreshCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/api/v1/auth",
		HTTPOnly: true,
		Secure:   h.cfg.IsProduction(),
		SameSite: "Lax",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func (h *authHandler) setNoStore(c *fiber.Ctx) {
	c.Set(fiber.HeaderCacheControl, "no-store")
	c.Set(fiber.HeaderPragma, "no-cache")
	c.Set(fiber.HeaderExpires, "0")
}

type requestCtx struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func ctxWithTimeout(c *fiber.Ctx, timeout time.Duration) requestCtx {
	requestID, _ := c.Locals(middleware.LocalRequestID).(string)
	ctx, cancel := context.WithTimeout(c.Context(), timeout)
	ctx = context.WithValue(ctx, domain.RequestIDKey, requestID)
	return requestCtx{ctx: ctx, cancel: cancel}
}
