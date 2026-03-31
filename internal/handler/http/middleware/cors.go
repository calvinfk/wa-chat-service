package middleware

import (
	"wa_chat_service/config"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

// Sets up CORS headers based on the provided configuration.
func Cors(cfg *config.CORS) fiber.Handler {
	if !cfg.CorsEnabled {
		return func(ctx fiber.Ctx) error {
			return ctx.Next()
		}
	}
	config := cors.Config{
		AllowMethods:     cfg.CorsAllowMethods,
		AllowHeaders:     cfg.CorsAllowHeaders,
		ExposeHeaders:    cfg.CorsExposeHeaders,
		AllowCredentials: cfg.CorsAllowCredentials,
	}
	if len(cfg.CorsAllowOrigins) > 0 {
		config.AllowOrigins = cfg.CorsAllowOrigins
	}
	return cors.New(config)
}
