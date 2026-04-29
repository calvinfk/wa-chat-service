package http_v1

import (
	"encoding/json"
	"fmt"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/handler/http/middleware"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/api_response"
	"wa_chat_service/pkg/filter_request"

	"github.com/gofiber/fiber/v3"
)

type ChatHandler struct {
	messageUsecase usecase.Message
	chatUsecase    usecase.Chat
}

func NewChatHandler(messageUsecase usecase.Message, chatUsecase usecase.Chat) HandlerV1 {
	return &ChatHandler{
		messageUsecase: messageUsecase,
		chatUsecase:    chatUsecase,
	}
}

func (h *ChatHandler) RegisterRoute(api fiber.Router) {
	chatGroup := api.Group("/chat")
	{
		chatGroup.Post("/send", middleware.Protected(), h.sendMessage)
		chatGroup.Get("/by-phone-number-id", middleware.Protected(), h.getChatByPhoneNumberId)
		chatGroup.Get("/messages", middleware.Protected(), middleware.Role("admin", "agent"), h.getMessagesByChatID)
		chatGroup.Post("/close-ticket", middleware.Protected(), middleware.Role("admin"), h.closeTicket)
		chatGroup.Post("/assign-agent", middleware.Protected(), middleware.Role("admin"), h.assignAgent)
		chatGroup.Post("/create", middleware.Protected(), h.createChat)
	}
}

func (h *ChatHandler) sendMessage(ctx fiber.Ctx) error {
	var requestData dto.MessageSendRequest
	if err := ctx.Bind().Body(&requestData); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}

	jsonBody := ctx.Body()
	var additionalData map[string]any
	if err := json.Unmarshal(jsonBody, &additionalData); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}

	messageData, ok := additionalData[requestData.Type]
	if !ok {
		code, response := api_response.NewErrorApiResponse(false, fmt.Errorf("%s is required", requestData.Type))
		return ctx.Status(code).JSON(response)
	}
	if messageData == nil {
		code, response := api_response.NewErrorApiResponse(false, fmt.Errorf("%s is required", requestData.Type))
		return ctx.Status(code).JSON(response)
	}
	requestData.Payload = messageData

	authData := ctx.Locals("token_sub").(dto.AuthData)
	_, serverError, err := h.messageUsecase.SendMessage(ctx.Context(), nil, authData.TenantID, requestData)
	if err != nil {
		code, response := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(response)
	}
	code, response := api_response.NewApiResponse("Successfully sent message", nil)
	return ctx.Status(code).JSON(response)
}

func (h *ChatHandler) getChatByPhoneNumberId(ctx fiber.Ctx) error {
	var requestData filter_request.FilterRequest[dto.ChatGetByPhoneNumberIdRequest]
	if err := ctx.Bind().Query(&requestData.SpecificFilter); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	if err := ctx.Bind().Query(&requestData); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	authData := ctx.Locals("token_sub").(dto.AuthData)
	chats, serverError, err := h.chatUsecase.GetChatByPhoneNumberId(ctx.Context(), authData.TenantID, requestData)
	if err != nil {
		code, response := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(response)
	}
	code, response := api_response.NewApiResponse("Successfully retrieved chats", chats)
	return ctx.Status(code).JSON(response)
}

func (h *ChatHandler) getMessagesByChatID(ctx fiber.Ctx) error {
	var requestData filter_request.FilterRequest[dto.MessageGetByChatIDRequest]
	if err := ctx.Bind().Query(&requestData.SpecificFilter); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	if err := ctx.Bind().Query(&requestData); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	authData := ctx.Locals("token_sub").(dto.AuthData)
	messages, serverError, err := h.messageUsecase.GetMessagesByChatID(ctx.Context(), authData.TenantID, requestData)
	if err != nil {
		code, response := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(response)
	}
	code, response := api_response.NewApiResponse("Successfully retrieved messages", messages)
	return ctx.Status(code).JSON(response)
}

func (h *ChatHandler) closeTicket(ctx fiber.Ctx) error {
	var requestData dto.ChatCloseTicketRequest
	if err := ctx.Bind().Body(&requestData); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	authData := ctx.Locals("token_sub").(dto.AuthData)
	serverError, err := h.chatUsecase.CloseTicket(ctx.Context(), authData.TenantID, requestData)
	if err != nil {
		code, response := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(response)
	}
	code, response := api_response.NewApiResponse("Successfully closed ticket", nil)
	return ctx.Status(code).JSON(response)
}

func (h *ChatHandler) assignAgent(ctx fiber.Ctx) error {
	var requestData dto.ChatAssignAgentRequest
	if err := ctx.Bind().Body(&requestData); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	authData := ctx.Locals("token_sub").(dto.AuthData)
	serverError, err := h.chatUsecase.AssignAgent(ctx.Context(), authData.TenantID, requestData)
	if err != nil {
		code, response := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(response)
	}
	code, response := api_response.NewApiResponse("Successfully assigned agent", nil)
	return ctx.Status(code).JSON(response)
}

func (uc *ChatHandler) createChat(ctx fiber.Ctx) error {
	var requestData dto.ChatCreateRequest
	if err := ctx.Bind().Body(&requestData); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	authData := ctx.Locals("token_sub").(dto.AuthData)
	data, serverError, err := uc.chatUsecase.CreateChat(ctx.Context(), authData.TenantID, requestData)
	if err != nil {
		code, response := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(response)
	}
	code, response := api_response.NewApiResponse("Successfully created chat", data)
	return ctx.Status(code).JSON(response)
}
