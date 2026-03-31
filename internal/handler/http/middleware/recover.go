package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gofiber/fiber/v3"
)

// Recovers from any panics that occur during request handling.
// It logs the panic message and stack trace, and responds with a 500 Internal Server Error status and a generic error message to the client.
func Recover() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				log.Println("[ERROR][internal/handler/http/middleware/recover.go][Recover] Recovered from panic:", r, "\nStack trace:\n", string(stack))
				ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"code":    http.StatusInternalServerError,
					"message": "Internal server error",
					"data":    nil,
					"errors":  nil,
				})
				return
			}
		}()
		return ctx.Next()
	}
}
