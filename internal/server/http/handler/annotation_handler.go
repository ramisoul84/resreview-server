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

type AnnotationHandler interface {
	ListAnnotations(c *fiber.Ctx) error
	CreateAnnotation(c *fiber.Ctx) error
	UpdateAnnotation(c *fiber.Ctx) error
	DeleteAnnotation(c *fiber.Ctx) error
}

type annotationHandler struct {
	service  service.AnnotationService
	validate *validator.Validate
	log      logger.Logger
}

func NewAnnotationHandler(svc service.AnnotationService) AnnotationHandler {
	return &annotationHandler{service: svc, validate: validator.New(), log: logger.Get()}
}

// GET /api/v1/products/:productId/versions/:versionId/annotations
func (h *annotationHandler) ListAnnotations(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "annotation_handler",
		"method": "ListAnnotations",
	})

	versionID := c.Params("versionId")
	if versionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "version ID is required"})
	}

	annotations, err := h.service.ListByVersion(c.Context(), versionID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list annotations")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list annotations"})
	}

	return c.JSON(annotations)
}

// POST /api/v1/products/:productId/versions/:versionId/annotations
func (h *annotationHandler) CreateAnnotation(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "annotation_handler",
		"method": "CreateAnnotation",
	})

	userID, ok := c.Locals(middleware.LocalUserID).(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	versionID := c.Params("versionId")
	if versionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "version ID is required"})
	}

	var req domain.CreateAnnotationRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("invalid create annotation payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.validate.Struct(req); err != nil {
		return validationError(c, err)
	}

	annotation, err := h.service.CreateAnnotation(c.Context(), versionID, userID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrVersionNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "version not found"})
		default:
			log.Error().Err(err).Msg("failed to create annotation")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create annotation"})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(annotation)
}

// PUT /api/v1/products/:productId/versions/:versionId/annotations/:annotationId
func (h *annotationHandler) UpdateAnnotation(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "annotation_handler",
		"method": "UpdateAnnotation",
	})

	annotationID := c.Params("annotationId")
	if annotationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "annotation ID is required"})
	}

	var req domain.UpdateAnnotationRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("invalid update annotation payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.service.UpdateAnnotation(c.Context(), annotationID, req); err != nil {
		switch {
		case errors.Is(err, domain.ErrAnnotationNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "annotation not found"})
		default:
			log.Error().Err(err).Msg("failed to update annotation")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update annotation"})
		}
	}

	return c.JSON(fiber.Map{"message": "annotation updated"})
}

// DELETE /api/v1/products/:productId/versions/:versionId/annotations/:annotationId
func (h *annotationHandler) DeleteAnnotation(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "annotation_handler",
		"method": "DeleteAnnotation",
	})

	annotationID := c.Params("annotationId")
	if annotationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "annotation ID is required"})
	}

	if err := h.service.DeleteAnnotation(c.Context(), annotationID); err != nil {
		switch {
		case errors.Is(err, domain.ErrAnnotationNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "annotation not found"})
		default:
			log.Error().Err(err).Msg("failed to delete annotation")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete annotation"})
		}
	}

	return c.JSON(fiber.Map{"message": "annotation deleted"})
}
