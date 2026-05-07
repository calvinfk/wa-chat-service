package http_v1

import (
	"encoding/json"
	"fmt"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/handler/http/middleware"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/api_response"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/filter_request"

	"github.com/gofiber/fiber/v3"
)

type ChatHandler struct {
	chatUsecase usecase.Chat
	userUsecase usecase.User
}

func NewChatHandler(chatUsecase usecase.Chat, userUsecase usecase.User) HandlerV1 {
	return &ChatHandler{
		chatUsecase: chatUsecase,
		userUsecase: userUsecase,
	}
}

func (h *ChatHandler) RegisterRoute(api fiber.Router) {
	chatGroup := api.Group("/chat")
	{
		chatGroup.Post("/send", middleware.Protected(), h.sendMessage)
		chatGroup.Get("/by-phone-number-id", middleware.Protected(), h.getChatByPhoneNumberId)
		chatGroup.Get("/messages", middleware.Protected(), h.getMessagesByChatID)
		chatGroup.Post("/create", middleware.Protected(), h.createChat)
		chatGroup.Post("/close", middleware.Protected(), middleware.Role(model.UserRoleAgent), h.closeChat)
		chatGroup.Post("/assign-agent", middleware.Protected(), middleware.Role(model.UserRoleAdmin), h.assignAgent)
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

	// check if user can send message to the chat
	canSend, serverError, err := h.chatUsecase.CheckCanSendMessage(ctx.Context(), authData, requestData.ChatID, requestData.TicketID)
	if err != nil {
		code, response := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(response)
	}
	if !canSend {
		code, response := api_response.NewErrorApiResponse(false, errs.ErrGenericForbidden)
		return ctx.Status(code).JSON(response)
	}
	// resolve sender name
	if requestData.SenderName == "" {
		user, serverError, err := h.userUsecase.GetByID(ctx.Context(), authData.TenantID, dto.UserGetByIDRequest{
			ID: authData.UserID,
		})
		if err != nil {
			code, response := api_response.NewErrorApiResponse(serverError, err)
			return ctx.Status(code).JSON(response)
		}
		requestData.SenderName = user.Name
	}

	serverError, err = h.chatUsecase.SendMessage(ctx.Context(), nil, authData.TenantID, requestData)
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
	chats, serverError, err := h.chatUsecase.GetChatByPhoneNumberId(ctx.Context(), authData, requestData)
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
	messages, serverError, err := h.chatUsecase.GetMessagesByChatID(ctx.Context(), authData, requestData)
	if err != nil {
		code, response := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(response)
	}
	code, response := api_response.NewApiResponse("Successfully retrieved messages", messages)
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

func (h *ChatHandler) createChat(ctx fiber.Ctx) error {
	var requestData dto.ChatCreateRequest
	if err := ctx.Bind().Body(&requestData); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	authData := ctx.Locals("token_sub").(dto.AuthData)
	data, serverError, err := h.chatUsecase.CreateChat(ctx.Context(), authData.TenantID, requestData)
	if err != nil {
		code, response := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(response)
	}
	code, response := api_response.NewApiResponse("Successfully created chat", data)
	return ctx.Status(code).JSON(response)
}

func (h *ChatHandler) closeChat(ctx fiber.Ctx) error {
	var requestData dto.ChatCloseRequest
	if err := ctx.Bind().Body(&requestData); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	authData := ctx.Locals("token_sub").(dto.AuthData)
	serverError, err := h.chatUsecase.CloseChat(ctx.Context(), authData.TenantID, requestData)
	if err != nil {
		code, response := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(response)
	}
	code, response := api_response.NewApiResponse("Successfully closed chat", nil)
	return ctx.Status(code).JSON(response)
}
