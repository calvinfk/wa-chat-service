package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
)

func APIKeyAuth(apiKey string) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		headers := ctx.GetHeaders()
		keys := headers["X-API-Key"]
		if len(keys) == 0 {
			ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"message": "API key is required",
				"data":    nil,
				"errors":  nil,
			})
			return nil
		}
		providedKey := keys[0]
		if providedKey == "" {
			ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"message": "API key is required",
				"data":    nil,
				"errors":  nil,
			})
			return nil
		}
		if providedKey != apiKey {
			ctx.Status(http.StatusForbidden).JSON(fiber.Map{
				"code":    http.StatusForbidden,
				"message": "Invalid API key",
				"data":    nil,
				"errors":  nil,
			})
			return nil
		}
		return ctx.Next()
	}
}
