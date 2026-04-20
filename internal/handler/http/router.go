package handler_http

import (
	"net/http"
	"wa_chat_service/config"
	"wa_chat_service/internal/handler/http/middleware"
	http_v1 "wa_chat_service/internal/handler/http/v1"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func NewRouter(app *fiber.App, config *config.Config, handlerHTTPV1 http_v1.HandlerHTTPV1) {
	// Set up middleware
	app.Use(
		logger.New(),
		middleware.Cors(&config.CORS),
		middleware.AccessToken(handlerHTTPV1.AccessTokenService, handlerHTTPV1.EncryptService, handlerHTTPV1.ZSLog),
		// middleware.ActivityLog(handlerHTTPV1.ActivityLogUsecase),
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
	http_v1.New(apiV1, handlerHTTPV1, config)

}
