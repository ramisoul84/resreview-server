package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const (
	LocalUserID      = "user_id"
	LocalIsAdmin     = "is_admin"
	LocalAccessToken = "access_token"
)

func JWTAuth(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := extractBearer(c)
		if raw == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization token",
			})
		}

		claims, err := parseJWT(raw, secret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}
		c.Locals(LocalUserID, claims.Subject)
		c.Locals(LocalIsAdmin, claims.IsAdmin)
		c.Locals(LocalAccessToken, raw)
		return c.Next()
	}
}

type accessClaims struct {
	jwt.RegisteredClaims
	IsAdmin bool `json:"is_admin"`
}

func parseJWT(tokenStr, secret string) (*accessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &accessClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	claims, ok := token.Claims.(*accessClaims)
	if !ok {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "invalid token claims")
	}
	return claims, nil
}

func extractBearer(c *fiber.Ctx) string {
	header := c.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}
	return ""
}
