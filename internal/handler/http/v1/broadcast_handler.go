package http_v1

import (
	"net/http"
	"strings"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/handler/http/middleware"
	"wa_chat_service/internal/service"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/api_response"
	"wa_chat_service/pkg/filter_request"

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
		broadcastRoute.Post("/upsert", middleware.Protected(), h.upsertBroadcast)
		broadcastRoute.Post("/schedule", middleware.Protected(), h.scheduleBroadcast)
		broadcastRoute.Post("/send", middleware.Jwt(h.encryptService, h.jwtService, http.StatusOK, true), h.sendBroadcast)
		broadcastRoute.Put("/cancel", middleware.Protected(), h.cancelBroadcast)
		broadcastRoute.Get("/get-filtered", middleware.Protected(), h.getFilteredBroadcast)
		broadcastRoute.Get("/get-recipients-filtered", middleware.Protected(), h.getFilteredBroadcastRecipients)
	}
}

func (h *BroadcastHandler) upsertBroadcast(ctx fiber.Ctx) error {
	var inputData dto.BroadcastUpsertRequest
	if err := ctx.Bind().All(&inputData); err != nil {
		httpCode, apiResponse := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	data, serverError, err := h.broadcastUsecase.UpsertBroadcast(ctx.Context(), inputData)
	httpCode, apiResponse := api_response.NewApiResponse(serverError, err, "Successfully insert/update broadcast", data)
	return ctx.Status(httpCode).JSON(apiResponse)
}

func (h *BroadcastHandler) scheduleBroadcast(ctx fiber.Ctx) error {
	var inputData dto.BroadcastScheduleRequest
	if err := ctx.Bind().All(&inputData); err != nil {
		httpCode, apiResponse := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	serverError, err := h.broadcastUsecase.ScheduleBroadcast(ctx.Context(), inputData)
	httpCode, apiResponse := api_response.NewApiResponse(serverError, err, "Successfully schedule broadcast", nil)
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

func (h *BroadcastHandler) cancelBroadcast(ctx fiber.Ctx) error {
	var inputData dto.BroadcastCancelRequest
	if err := ctx.Bind().All(&inputData); err != nil {
		httpCode, apiResponse := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	serverError, err := h.broadcastUsecase.CancelBroadcast(ctx.Context(), inputData)
	httpCode, apiResponse := api_response.NewApiResponse(serverError, err, "Successfully cancel broadcast", nil)
	return ctx.Status(httpCode).JSON(apiResponse)
}

func (h *BroadcastHandler) getFilteredBroadcast(ctx fiber.Ctx) error {
	var inputData filter_request.FilterRequest[dto.BroadcastGetFilteredRequest]
	if err := ctx.Bind().Query(&inputData.SpecificFilter); err != nil {
		httpCode, apiResponse := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	if err := ctx.Bind().Query(&inputData); err != nil {
		httpCode, apiResponse := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	if err := inputData.Validate(); err != nil {
		httpCode, apiResponse := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	data, serverError, err := h.broadcastUsecase.GetFilteredBroadcast(ctx.Context(), inputData)
	httpCode, apiResponse := api_response.NewApiResponse(serverError, err, "Successfully get filtered broadcast", data)
	return ctx.Status(httpCode).JSON(apiResponse)
}

func (h *BroadcastHandler) getFilteredBroadcastRecipients(ctx fiber.Ctx) error {
	var inputData filter_request.FilterRequest[dto.BroadcastGetRecipientsFilteredRequest]
	if err := ctx.Bind().Query(&inputData.SpecificFilter); err != nil {
		httpCode, apiResponse := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	if err := ctx.Bind().Query(&inputData); err != nil {
		httpCode, apiResponse := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	if err := inputData.Validate(); err != nil {
		httpCode, apiResponse := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	data, serverError, err := h.broadcastUsecase.GetFilteredBroadcastRecipients(ctx.Context(), inputData)
	httpCode, apiResponse := api_response.NewApiResponse(serverError, err, "Successfully get filtered broadcast recipients", data)
	return ctx.Status(httpCode).JSON(apiResponse)
}
