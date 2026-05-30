package middleware

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ramisoul84/resreview-server/internal/domain"
)

const LocalRequestID = "request_id"

func InjectRequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get(fiber.HeaderXRequestID)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set(fiber.HeaderXRequestID, requestID)
		c.Locals(LocalRequestID, requestID)

		ctx := context.WithValue(c.UserContext(), domain.RequestIDKey, requestID)
		c.SetUserContext(ctx)

		return c.Next()
	}
}
