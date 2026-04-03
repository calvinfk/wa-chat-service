package message_usecase

import (
	"context"
	"fmt"
	"log"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/pkg/formatter"
	whatsapp_business_component "wa_chat_service/pkg/meta/whatsapp_business/component"
)

type MessageUsecase struct {
	messageRepository repository.Message
	chatRepository    repository.Chat
	whatsappService   service.WhatsappService
}

func NewMessageUsecase(messageRepository repository.Message, chatRepository repository.Chat, whatsappService service.WhatsappService) *MessageUsecase {
	return &MessageUsecase{
		messageRepository: messageRepository,
		chatRepository:    chatRepository,
		whatsappService:   whatsappService,
	}
}

func (u *MessageUsecase) SendMessage(ctx context.Context, inputData dto.MessageSendRequest) (model.Message, bool, error) {
	var err error
	var response model.Message
	// create chat header if not exist
	chat := model.Chat{
		DocumentID:  fmt.Sprintf("%s-%s", inputData.RecipientID, inputData.PhoneNumberID),
		ChatType:    "individual",
		DisplayName: inputData.RecipientName,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}
	_, err = u.chatRepository.Insert(ctx, nil, chat)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to insert chat:", err)
		return response, true, err
	}
	component, err := whatsapp_business_component.New(inputData.Type, inputData.Payload)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to validate message component:", err)
		return response, false, err
	}
	sendResponse, err := u.whatsappService.SendMessage(ctx, inputData.PhoneNumberID, inputData.RecipientID, component)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to send message:", err)
		return response, true, err
	}
	payloadData, err := formatter.AnyToJsonString(component.GetPayload())
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to convert payload to JSON")
	}
	message := model.Message{
		DocumentID:      sendResponse.Messages[0].ID,
		ChatID:          chat.DocumentID,
		MessageType:     string(component.GetType()),
		MessageCategory: "-",
		SenderName:      inputData.SenderName,
		Payload:         payloadData,
		Content:         component.GetMessage(),
		Status:          "-",
		CreatedAt:       time.Now().Unix(),
		UpdatedAt:       time.Now().Unix(),
	}
	response, err = u.messageRepository.Insert(ctx, nil, message)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to insert message:", err)
		return response, true, err
	}
	return response, false, nil
}
