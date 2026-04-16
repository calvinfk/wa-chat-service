package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// Recovers from any panics that occur during request handling.
// It logs the panic message and stack trace, and responds with a 500 Internal Server Error status and a generic error message to the client.
func Recover(zsLog *zap.SugaredLogger) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				zsLog.Errorf("[ERROR][internal/handler/http/middleware/recover.go][Recover] Recovered from panic: %v\nStack trace:\n%s", r, string(stack))
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
