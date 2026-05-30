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

type ProductHandler interface {
	CreateProduct(c *fiber.Ctx) error
	GetProduct(c *fiber.Ctx) error
	ListProducts(c *fiber.Ctx) error
	UpdateProduct(c *fiber.Ctx) error
	DeleteProduct(c *fiber.Ctx) error
}

type productHandler struct {
	service  service.ProductService
	validate *validator.Validate
	log      logger.Logger
}

func NewProductHandler(svc service.ProductService) ProductHandler {
	return &productHandler{service: svc, validate: validator.New(), log: logger.Get()}
}

// POST /api/v1/products
func (h *productHandler) CreateProduct(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "product_handler",
		"method": "CreateProduct",
	})

	userID, ok := c.Locals(middleware.LocalUserID).(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req domain.CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("invalid create product payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return validationError(c, err)
	}

	product, err := h.service.CreateProduct(c.Context(), userID, req.Name)
	if err != nil {
		log.Error().Err(err).Msg("failed to create product")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create product",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(product)
}

// GET /api/v1/products
func (h *productHandler) ListProducts(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "product_handler",
		"method": "ListProducts",
	})

	userID, ok := c.Locals(middleware.LocalUserID).(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	products, err := h.service.ListProducts(c.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list products")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list products",
		})
	}

	return c.JSON(products)
}

// GET /api/v1/products/:productId
func (h *productHandler) GetProduct(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "product_handler",
		"method": "GetProduct",
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

	product, err := h.service.GetProduct(c.Context(), productID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProductNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "product not found",
			})
		default:
			log.Error().Err(err).Msg("failed to get product")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get product",
			})
		}
	}

	return c.JSON(product)
}

// PUT /api/v1/products/:productId
func (h *productHandler) UpdateProduct(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "product_handler",
		"method": "UpdateProduct",
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

	var req domain.UpdateProductRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("invalid update product payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return validationError(c, err)
	}

	err := h.service.UpdateProduct(c.Context(), productID, userID, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProductNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "product not found",
			})
		default:
			log.Error().Err(err).Msg("failed to update product")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to update product",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "product updated successfully",
	})
}

// DELETE /api/v1/products/:productId
func (h *productHandler) DeleteProduct(c *fiber.Ctx) error {
	log := h.log.WithFields(map[string]any{
		"layer":  "product_handler",
		"method": "DeleteProduct",
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

	err := h.service.DeleteProduct(c.Context(), productID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProductNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "product not found",
			})
		default:
			log.Error().Err(err).Msg("failed to delete product")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to delete product",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "product deleted successfully",
	})
}
