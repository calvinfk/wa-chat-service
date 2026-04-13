package http_internal

import (
	"net/http"
	"wa_chat_service/config"
	"wa_chat_service/internal/handler/http/middleware"
	http_v1 "wa_chat_service/internal/handler/http/v1"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func NewRouter(app *fiber.App, config *config.Config, routerHandlerV1 http_v1.RouterHandlerV1) {
	// Set up middleware
	app.Use(
		logger.New(),
		middleware.Recover(),
		middleware.Cors(&config.CORS),
		middleware.OptionsRoute(),
		middleware.FileSizeLimit(16*1024*1024), // 16MB
		middleware.AccessToken(routerHandlerV1.AccessTokenService, routerHandlerV1.EncryptService),
		// middleware.ActivityLog(routerHandlerV1.ActivityLogUsecase),
	)

	api := app.Group("api")

	api.Get("/ping", func(ctx fiber.Ctx) error {
		return ctx.Status(http.StatusOK).JSON(fiber.Map{
			"code":    http.StatusOK,
			"data":    nil,
			"message": "pong",
			"errors":  nil,
		})
	})

	api.Get("ping-protected", middleware.Protected(), func(ctx fiber.Ctx) error {
		return ctx.Status(http.StatusOK).JSON(fiber.Map{
			"code":    http.StatusOK,
			"data":    nil,
			"message": "pong-protected",
			"errors":  nil,
		})
	})

	apiV1 := api.Group("v1")
	http_v1.NewApiV1Routes(apiV1, routerHandlerV1, config)

}
