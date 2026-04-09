package http_v1

import (
	"encoding/json"
	"fmt"
	"wa_chat_service/internal/dto"
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
		chatGroup.Post("/send", h.sendMessage)
		chatGroup.Get("/by-phone-number-id", h.getChatByPhoneNumberID)
		chatGroup.Get("/messages", h.getMessagesByChatID)
	}
}

func (h *ChatHandler) sendMessage(ctx fiber.Ctx) error {
	var requestData dto.MessageSendRequest
	if err := ctx.Bind().Body(&requestData); err != nil {
		code, response := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(code).JSON(response)
	}

	jsonBody := ctx.Body()
	var additionalData map[string]any
	if err := json.Unmarshal(jsonBody, &additionalData); err != nil {
		code, response := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(code).JSON(response)
	}

	messageData, ok := additionalData[requestData.Type]
	if !ok {
		code, response := api_response.NewApiResponse(false, fmt.Errorf("%s is required", requestData.Type), "", nil)
		return ctx.Status(code).JSON(response)
	}
	if messageData == nil {
		code, response := api_response.NewApiResponse(false, fmt.Errorf("%s is required", requestData.Type), "", nil)
		return ctx.Status(code).JSON(response)
	}
	requestData.Payload = messageData

	_, serverError, err := h.messageUsecase.SendMessage(ctx.Context(), requestData)
	code, response := api_response.NewApiResponse(serverError, err, "Successfully sent message", nil)
	return ctx.Status(code).JSON(response)
}

func (h *ChatHandler) getChatByPhoneNumberID(ctx fiber.Ctx) error {
	var requestData filter_request.FilterRequest[dto.ChatGetByPhoneNumberIDRequest]
	if err := ctx.Bind().Query(&requestData.SpecificFilter); err != nil {
		code, response := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(code).JSON(response)
	}
	if err := ctx.Bind().Query(&requestData); err != nil {
		code, response := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(code).JSON(response)
	}
	chats, serverError, err := h.chatUsecase.GetChatByPhoneNumberID(ctx.Context(), requestData)
	code, response := api_response.NewApiResponse(serverError, err, "Successfully retrieved chats", chats)
	return ctx.Status(code).JSON(response)
}

func (h *ChatHandler) getMessagesByChatID(ctx fiber.Ctx) error {
	var requestData filter_request.FilterRequest[dto.MessageGetByChatIDRequest]
	if err := ctx.Bind().Query(&requestData.SpecificFilter); err != nil {
		code, response := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(code).JSON(response)
	}
	if err := ctx.Bind().Query(&requestData); err != nil {
		code, response := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(code).JSON(response)
	}
	messages, serverError, err := h.messageUsecase.GetMessagesByChatID(ctx.Context(), requestData)
	code, response := api_response.NewApiResponse(serverError, err, "Successfully retrieved messages", messages)
	return ctx.Status(code).JSON(response)
}
