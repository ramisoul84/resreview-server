package handler

import (
	"errors"

	"io"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/ramisoul84/resreview-server/internal/domain"
	"github.com/ramisoul84/resreview-server/internal/server/http/middleware"
	"github.com/ramisoul84/resreview-server/internal/service"
	"github.com/ramisoul84/resreview-server/pkg/logger"
)

type VersionHandler interface {
	CreateVersion(c *fiber.Ctx) error
	GetVersion(c *fiber.Ctx) error
	ListVersions(c *fiber.Ctx) error
	UpdateVersion(c *fiber.Ctx) error
	DeleteVersion(c *fiber.Ctx) error
	UploadImage(c *fiber.Ctx) error
}

type versionHandler struct {
	service  service.VersionService
	validate *validator.Validate
	log      logger.Logger
}

func NewVersionHandler(svc service.VersionService) VersionHandler {
	return &versionHandler{service: svc, validate: validator.New(), log: logger.Get()}
}

// POST /api/v1/products/:productId/versions
func (h *versionHandler) CreateVersion(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "version_handler",
		"method": "CreateVersion",
	})

	userID, ok := c.Locals(middleware.LocalUserID).(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	productID := c.Params("productId")
	if productID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product ID is required",
		})
	}

	var req domain.CreateVersionRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("invalid create version payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return validationError(c, err)
	}

	version, err := h.service.CreateVersion(c.Context(), productID, userID, req.Label, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProductNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "product not found",
			})
		default:
			log.Error().Err(err).Msg("failed to create version")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to create version",
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(version)
}

// GET /api/v1/products/:productId/versions
func (h *versionHandler) ListVersions(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "version_handler",
		"method": "ListVersions",
	})

	productID := c.Params("productId")
	if productID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product ID is required",
		})
	}

	versions, err := h.service.ListVersions(c.Context(), productID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list versions")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list versions",
		})
	}

	return c.JSON(versions)
}

// GET /api/v1/products/:productId/versions/:versionId
func (h *versionHandler) GetVersion(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "version_handler",
		"method": "GetVersion",
	})

	productID := c.Params("productId")
	if productID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product ID is required",
		})
	}

	versionID := c.Params("versionId")
	if versionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "version ID is required",
		})
	}

	version, err := h.service.GetVersion(c.Context(), versionID, productID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrVersionNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "version not found",
			})
		default:
			log.Error().Err(err).Msg("failed to get version")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get version",
			})
		}
	}

	return c.JSON(version)
}

// PUT /api/v1/products/:productId/versions/:versionId
func (h *versionHandler) UpdateVersion(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "version_handler",
		"method": "UpdateVersion",
	})

	userID, ok := c.Locals(middleware.LocalUserID).(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	productID := c.Params("productId")
	if productID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product ID is required",
		})
	}

	versionID := c.Params("versionId")
	if versionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "version ID is required",
		})
	}

	var req domain.UpdateVersionRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("invalid update version payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return validationError(c, err)
	}

	err := h.service.UpdateVersion(c.Context(), versionID, productID, userID, req.Label, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProductNotFound), errors.Is(err, domain.ErrVersionNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "version not found",
			})
		default:
			log.Error().Err(err).Msg("failed to update version")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to update version",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "version updated successfully",
	})
}

// POST /api/v1/products/:productId/versions/:versionId/upload
func (h *versionHandler) UploadImage(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "version_handler",
		"method": "UploadImage",
	})

	userID, ok := c.Locals(middleware.LocalUserID).(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	productID := c.Params("productId")
	if productID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product ID is required",
		})
	}

	versionID := c.Params("versionId")
	if versionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "version ID is required",
		})
	}

	file, err := c.FormFile("image")
	if err != nil {
		log.Warn().Err(err).Msg("missing image file")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "image file is required",
		})
	}

	f, err := file.Open()
	if err != nil {
		log.Error().Err(err).Msg("failed to open uploaded file")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to process image",
		})
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		log.Error().Err(err).Msg("failed to read uploaded file")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to read image",
		})
	}

	mimeType := file.Header.Get("Content-Type")

	version, err := h.service.UploadVersionImage(c.Context(), versionID, productID, userID, data, mimeType)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProductNotFound), errors.Is(err, domain.ErrVersionNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "version not found",
			})
		case errors.Is(err, domain.ErrInvalidMimeType):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid image type, only JPEG, PNG, and WebP are allowed",
			})
		case errors.Is(err, domain.ErrFileTooLarge):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "image file too large",
			})
		default:
			log.Error().Err(err).Msg("failed to upload version image")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to upload image",
			})
		}
	}

	return c.JSON(version)
}

// DELETE /api/v1/products/:productId/versions/:versionId
func (h *versionHandler) DeleteVersion(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "version_handler",
		"method": "DeleteVersion",
	})

	userID, ok := c.Locals(middleware.LocalUserID).(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	productID := c.Params("productId")
	if productID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product ID is required",
		})
	}

	versionID := c.Params("versionId")
	if versionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "version ID is required",
		})
	}

	err := h.service.DeleteVersion(c.Context(), versionID, productID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProductNotFound), errors.Is(err, domain.ErrVersionNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "version not found",
			})
		default:
			log.Error().Err(err).Msg("failed to delete version")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to delete version",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "version deleted successfully",
	})
}
