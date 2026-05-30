package handler

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// ─── Validation ───────────────────────────────────────────────────────────────
func validationError(c *fiber.Ctx, err error) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
		"error":  "validation failed",
		"fields": formatValidationErrors(err),
	})
}

func formatValidationErrors(err error) map[string]string {
	fields := make(map[string]string)
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			fields[fe.Field()] = validationMessage(fe)
		}
	}
	return fields
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return "too short (min length: " + fe.Param() + ")"
	case "max":
		return "too long (max length: " + fe.Param() + ")"
	case "oneof":
		return "must be one of: " + fe.Param()
	default:
		return "invalid value"
	}
}
