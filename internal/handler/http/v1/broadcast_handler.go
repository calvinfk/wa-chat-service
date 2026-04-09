package http_v1

import (
	"wa_chat_service/internal/handler/http/middleware"
	"wa_chat_service/internal/service"

	"github.com/gofiber/fiber/v3"
)

type BroadcastHandler struct {
	encryptService service.Encrypt
	jwtService     service.JWT
}

func NewBroadcastHandler(encryptService service.Encrypt, jwtService service.JWT) *BroadcastHandler {
	return &BroadcastHandler{
		encryptService: encryptService,
		jwtService:     jwtService,
	}
}

func (h *BroadcastHandler) RegisterRoute(api fiber.Router) {
	broadcastRoute := api.Group("/broadcast")
	{
		broadcastRoute.Post("/send", middleware.Jwt(h.encryptService, h.jwtService), h.sendBroadcast)
	}
}

func (h *BroadcastHandler) sendBroadcast(ctx fiber.Ctx) error {
	return ctx.JSON(fiber.Map{
		"message": "Broadcast sent successfully",
	})
}
