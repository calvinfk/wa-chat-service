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
	textComponent := &whatsapp_business_component.Text{
		Body: inputData.Content,
	}
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

	sendResponse, err := u.whatsappService.SendMessage(ctx, inputData.PhoneNumberID, inputData.RecipientID, textComponent)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to send message:", err)
		return response, true, err
	}
	message := model.Message{
		DocumentID:      sendResponse.Messages[0].ID,
		ChatID:          chat.DocumentID,
		MessageType:     textComponent.GetType(),
		MessageCategory: "-",
		SenderName:      inputData.SenderName,
		Content:         textComponent.Body,
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
