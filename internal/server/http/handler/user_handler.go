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

type UserHandler interface {
	ListUsers(c *fiber.Ctx) error
	UpdateRole(c *fiber.Ctx) error
	UpdateProfile(c *fiber.Ctx) error
}

type userHandler struct {
	service  service.UserService
	validate *validator.Validate
	log      logger.Logger
}

func NewUserHandler(service service.UserService) UserHandler {
	return &userHandler{service: service, validate: validator.New(), log: logger.Get()}
}

// GET /api/v1/users
func (h *userHandler) ListUsers(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "user_handler",
		"method": "ListUsers",
	})

	users, err := h.service.ListUsers(c.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to list users")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list users",
		})
	}

	return c.JSON(users)
}

// PATCH /api/v1/users/profile
func (h *userHandler) UpdateProfile(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "user_handler",
		"method": "UpdateProfile",
	})

	userID, ok := c.Locals(middleware.LocalUserID).(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req domain.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("invalid update profile payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return validationError(c, err)
	}

	err := h.service.UpdateProfile(c.Context(), userID, req.Name, req.Color)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "user not found",
			})
		default:
			log.Error().Err(err).Msg("failed to update profile")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to update profile",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "profile updated successfully",
	})
}

// PATCH /api/v1/users/:userId/role
func (h *userHandler) UpdateRole(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "user_handler",
		"method": "UpdateRole",
	})

	isAdmin, ok := c.Locals(middleware.LocalIsAdmin).(bool)
	if !ok || !isAdmin {
		log.Warn().Msg("non-admin user attempted to update role")
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "admin access required",
		})
	}

	userID := c.Params("userId")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user ID is required",
		})
	}

	var req domain.UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("invalid update role payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return validationError(c, err)
	}

	err := h.service.UpdateRole(c.Context(), userID, req.IsAdmin)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "user not found",
			})
		default:
			log.Error().Err(err).Msg("failed to update role")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to update role",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "role updated successfully",
	})
}
