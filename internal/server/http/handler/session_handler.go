package handler

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/ramisoul84/resreview-server/internal/domain"
	"github.com/ramisoul84/resreview-server/internal/server/http/middleware"
	"github.com/ramisoul84/resreview-server/internal/service"
	"github.com/ramisoul84/resreview-server/pkg/logger"
)

type SessionHandler interface {
	CreateSession(c *fiber.Ctx) error
	GetSession(c *fiber.Ctx) error
	ListSessions(c *fiber.Ctx) error
	UpdateSession(c *fiber.Ctx) error
	DeleteSession(c *fiber.Ctx) error
}

type sessionHandler struct {
	service  service.SessionService
	validate *validator.Validate
	log      logger.Logger
}

func NewSessionHandler(svc service.SessionService) SessionHandler {
	return &sessionHandler{service: svc, validate: validator.New(), log: logger.Get()}
}

// POST /api/v1/sessions
func (h *sessionHandler) CreateSession(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "session_handler",
		"method": "CreateSession",
	})

	userID, ok := c.Locals(middleware.LocalUserID).(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req domain.CreateSessionRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("invalid create session payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return validationError(c, err)
	}

	session, err := h.service.CreateSession(c.Context(), userID, req.Name)
	if err != nil {
		log.Error().Err(err).Msg("failed to create session")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create session",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(session)
}

// GET /api/v1/sessions
func (h *sessionHandler) ListSessions(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "session_handler",
		"method": "ListSessions",
	})

	userID, ok := c.Locals(middleware.LocalUserID).(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	sessions, err := h.service.ListSessions(c.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list sessions")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list sessions",
		})
	}

	return c.JSON(sessions)
}

// GET /api/v1/sessions/:sessionId
func (h *sessionHandler) GetSession(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "session_handler",
		"method": "GetSession",
	})

	userID, ok := c.Locals(middleware.LocalUserID).(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	sessionID := c.Params("sessionId")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "session ID is required",
		})
	}

	session, err := h.service.GetSession(c.Context(), sessionID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSessionNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "session not found",
			})
		default:
			log.Error().Err(err).Msg("failed to get session")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get session",
			})
		}
	}

	return c.JSON(session)
}

// PUT /api/v1/sessions/:sessionId
func (h *sessionHandler) UpdateSession(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "session_handler",
		"method": "UpdateSession",
	})

	userID, ok := c.Locals(middleware.LocalUserID).(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	sessionID := c.Params("sessionId")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "session ID is required",
		})
	}

	var req domain.UpdateSessionRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("invalid update session payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return validationError(c, err)
	}

	err := h.service.UpdateSession(c.Context(), sessionID, userID, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSessionNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "session not found",
			})
		default:
			log.Error().Err(err).Msg("failed to update session")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to update session",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "session updated successfully",
	})
}

// DELETE /api/v1/sessions/:sessionId
func (h *sessionHandler) DeleteSession(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "session_handler",
		"method": "DeleteSession",
	})

	userID, ok := c.Locals(middleware.LocalUserID).(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	sessionID := c.Params("sessionId")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "session ID is required",
		})
	}

	err := h.service.DeleteSession(c.Context(), sessionID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSessionNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "session not found",
			})
		default:
			log.Error().Err(err).Msg("failed to delete session")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to delete session",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "session deleted successfully",
	})
}
