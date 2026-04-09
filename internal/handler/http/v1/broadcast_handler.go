package http_v1

import (
	"wa_chat_service/internal/handler/http/middleware"
	"wa_chat_service/internal/service"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/api_response"

	"github.com/gofiber/fiber/v3"
)

type BroadcastHandler struct {
	broadcastUsecase usecase.Broadcast
	encryptService   service.Encrypt
	jwtService       service.JWT
}

func NewBroadcastHandler(broadcastUsecase usecase.Broadcast, encryptService service.Encrypt, jwtService service.JWT) *BroadcastHandler {
	return &BroadcastHandler{
		broadcastUsecase: broadcastUsecase,
		encryptService:   encryptService,
		jwtService:       jwtService,
	}
}

func (h *BroadcastHandler) RegisterRoute(api fiber.Router) {
	broadcastRoute := api.Group("/broadcast")
	{
		broadcastRoute.Post("/schedule", h.scheduleBroadcast)
		broadcastRoute.Post("/send", middleware.Jwt(h.encryptService, h.jwtService), h.sendBroadcast)
	}
}

func (h *BroadcastHandler) scheduleBroadcast(ctx fiber.Ctx) error {
	serverError, err := h.broadcastUsecase.ScheduleBroadcast()
	httpCode, apiResponse := api_response.NewApiResponse(serverError, err, "Successfully scheduled broadcast", nil)
	return ctx.Status(httpCode).JSON(apiResponse)
}

func (h *BroadcastHandler) sendBroadcast(ctx fiber.Ctx) error {
	ctx.SendStatus(fiber.StatusOK)
	go func() {
		h.broadcastUsecase.SendBroadcast()
	}()
	return nil
}
