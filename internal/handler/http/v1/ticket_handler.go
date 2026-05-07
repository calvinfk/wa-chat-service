package http_v1

import (
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/handler/http/middleware"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/api_response"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type TicketHandler struct {
	ticketUsecase usecase.Ticket
	zsLog         *zap.SugaredLogger
}

func NewTicketHandler(ticketUsecase usecase.Ticket, zsLog *zap.SugaredLogger) HandlerV1 {
	return &TicketHandler{
		ticketUsecase: ticketUsecase,
		zsLog:         zsLog,
	}
}

func (h *TicketHandler) RegisterRoute(api fiber.Router) {
	ticketGroup := api.Group("/ticket")
	{
		ticketGroup.Post("/close", middleware.Protected(), middleware.Role(model.UserRoleAdmin), h.closeTicket)
		ticketGroup.Post("/assign-agent", middleware.Protected(), middleware.Role(model.UserRoleAdmin), h.assignAgent)
		ticketGroup.Get("/analytics", middleware.Protected(), middleware.Role(model.UserRoleAdmin), h.getAnalytics)
	}
}

func (h *TicketHandler) closeTicket(ctx fiber.Ctx) error {
	var requestData dto.TicketCloseRequest
	if err := ctx.Bind().Body(&requestData); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	authData := ctx.Locals("token_sub").(dto.AuthData)
	serverError, err := h.ticketUsecase.CloseTicket(ctx.Context(), authData.TenantID, requestData)
	if err != nil {
		code, response := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(response)
	}
	code, response := api_response.NewApiResponse("Successfully closed ticket", nil)
	return ctx.Status(code).JSON(response)
}

func (h *TicketHandler) assignAgent(ctx fiber.Ctx) error {
	var requestData dto.TicketAssignAgentRequest
	if err := ctx.Bind().Body(&requestData); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	authData := ctx.Locals("token_sub").(dto.AuthData)
	serverError, err := h.ticketUsecase.AssignAgent(ctx.Context(), authData.TenantID, requestData)
	if err != nil {
		code, response := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(response)
	}
	code, response := api_response.NewApiResponse("Successfully assigned agent", nil)
	return ctx.Status(code).JSON(response)
}

func (h *TicketHandler) getAnalytics(ctx fiber.Ctx) error {
	var requestData dto.TicketGetAnalyticsRequest
	if err := ctx.Bind().Query(&requestData); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	authData := ctx.Locals("token_sub").(dto.AuthData)
	data, serverError, err := h.ticketUsecase.GetTicketAnalytics(ctx.Context(), authData.TenantID, requestData)
	if err != nil {
		code, response := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(response)
	}
	code, response := api_response.NewApiResponse("Successfully retrieved ticket analytics", data)
	return ctx.Status(code).JSON(response)
}
