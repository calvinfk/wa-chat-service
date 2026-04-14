package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
)

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
