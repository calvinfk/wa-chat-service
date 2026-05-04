package middleware

import (
	"wa_chat_service/config"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

// Sets up CORS headers based on the provided configuration.
// Cors is used for controlling cross-origin requests to the API
func Cors(config *config.CORS) fiber.Handler {
	if !config.CorsEnabled {
		return func(ctx fiber.Ctx) error {
			return ctx.Next()
		}
	}
	corsConfig := cors.Config{
		AllowMethods:     config.CorsAllowMethods,
		AllowHeaders:     config.CorsAllowHeaders,
		ExposeHeaders:    config.CorsExposeHeaders,
		AllowCredentials: config.CorsAllowCredentials,
	}
	if len(config.CorsAllowOrigins) > 0 {
		corsConfig.AllowOrigins = config.CorsAllowOrigins
	}
	return cors.New(corsConfig)
}
