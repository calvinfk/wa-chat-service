package http_v1

import (
	"net/http"
	"strings"
	"wa_chat_service/internal/dto"
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
		broadcastRoute.Post("/schedule", middleware.Protected(), h.scheduleBroadcast)
		broadcastRoute.Post("/send", middleware.Jwt(h.encryptService, h.jwtService, http.StatusOK, true), h.sendBroadcast)
	}
}

func (h *BroadcastHandler) scheduleBroadcast(ctx fiber.Ctx) error {
	var inputData dto.BroadcastScheduleRequest
	if err := ctx.Bind().All(&inputData); err != nil {
		httpCode, apiResponse := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(httpCode).JSON(apiResponse)
	}

	serverError, err := h.broadcastUsecase.ScheduleBroadcast(ctx.Context(), inputData)
	httpCode, apiResponse := api_response.NewApiResponse(serverError, err, "Successfully scheduled broadcast", nil)
	return ctx.Status(httpCode).JSON(apiResponse)
}

func (h *BroadcastHandler) sendBroadcast(ctx fiber.Ctx) error {
	ctx.SendStatus(fiber.StatusOK)
	sub, ok := ctx.Value("jwt_sub").(string)
	if !ok {
		return nil
	}
	parts := strings.Split(sub, "_")
	task := parts[0]
	broadcastID := parts[1]
	if task != "broadcast" || broadcastID == "" {
		return nil
	}
	go func() {
		h.broadcastUsecase.SendBroadcast(ctx.Context(), broadcastID)
	}()
	return nil
}
