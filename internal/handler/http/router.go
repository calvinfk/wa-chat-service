package handler_http

import (
	"wa_chat_service/config"
	"wa_chat_service/internal/handler/http/middleware"
	http_v1 "wa_chat_service/internal/handler/http/v1"
	"wa_chat_service/pkg/api_response"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func NewRouter(app *fiber.App, config *config.Config, handlerHTTPV1 http_v1.HandlerHTTPV1) {
	// Set up middleware
	app.Use(
		logger.New(),
		middleware.Cors(&config.CORS),
		middleware.AccessToken(handlerHTTPV1.AccessTokenService, handlerHTTPV1.EncryptService, handlerHTTPV1.ZSLog),
	)

	api := app.Group("api")

	api.Get("/ping", func(ctx fiber.Ctx) error {
		httpCode, apiResponse := api_response.NewApiResponse("pong", nil)
		return ctx.Status(httpCode).JSON(apiResponse)
	})

	api.Get("ping-protected", middleware.Protected(), func(ctx fiber.Ctx) error {
		httpCode, apiResponse := api_response.NewApiResponse("pong-protected", nil)
		return ctx.Status(httpCode).JSON(apiResponse)
	})

	apiV1 := api.Group("v1")
	http_v1.New(apiV1, handlerHTTPV1, config)

}
