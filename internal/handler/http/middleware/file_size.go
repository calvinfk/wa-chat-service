package middleware

import "github.com/gofiber/fiber/v3"

func FileSizeLimit(maxSize int) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if ctx.Request().Header.ContentLength() > maxSize {
			return ctx.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"code":    fiber.StatusRequestEntityTooLarge,
				"data":    nil,
				"message": "File size exceeds the maximum limit",
				"errors":  nil,
			})
		}
		return ctx.Next()
	}
}
