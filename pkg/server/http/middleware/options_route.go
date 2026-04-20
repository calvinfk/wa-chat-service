package http_middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
)

func OptionsRoute() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if string(ctx.Request().Header.Method()) == http.MethodOptions {
			headers := ctx.GetHeaders()
			if headers["Origin"] != nil && headers["Access-Control-Request-Method"] != nil {
				ctx.Status(http.StatusNoContent)
				return nil
			}
		}
		return ctx.Next()
	}
}
