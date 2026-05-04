package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
)

// Protected is a middleware that checks if the token parsing middleware has set an error message or sub in the context.
// If there is an error message, it returns a 401 Unauthorized response with the error message.
// If there is no sub, it returns a 401 Unauthorized response with a generic "Unauthorized" message.
func Protected() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if tokenErrorMessage := ctx.Locals("token_error_message"); tokenErrorMessage != nil {
			ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"message": tokenErrorMessage,
				"data":    nil,
				"errors":  nil,
			})
			return nil
		} else if sub := ctx.Locals("token_sub"); sub == nil {
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
