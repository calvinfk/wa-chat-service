package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
)

func Protected() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if jwtErrorMessage := ctx.Get("jwt_error_message"); jwtErrorMessage != "" {
			ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"message": jwtErrorMessage,
				"data":    nil,
				"errors":  nil,
			})
			return nil
		} else if userID := ctx.Get("userID"); userID == "" {
			ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"message": "Unauthorized",
				"data":    nil,
				"errors":  nil,
			})
			return nil
		}
		return ctx.Next()
	}
}
