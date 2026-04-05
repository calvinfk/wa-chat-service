package message_usecase

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/formatter"
	"wa_chat_service/pkg/meta/whatsapp_business"
	whatsapp_business_component "wa_chat_service/pkg/meta/whatsapp_business/component"
)

type MessageUsecase struct {
	messageRepository     repository.Message
	chatRepository        repository.Chat
	phoneNumberRepository repository.PhoneNumber
	whatsappService       service.WhatsappService
	encryptService        service.Encrypt
}

func NewMessageUsecase(messageRepository repository.Message, chatRepository repository.Chat, phoneNumberRepository repository.PhoneNumber, whatsappService service.WhatsappService, encryptService service.Encrypt) *MessageUsecase {
	return &MessageUsecase{
		messageRepository:     messageRepository,
		chatRepository:        chatRepository,
		phoneNumberRepository: phoneNumberRepository,
		whatsappService:       whatsappService,
		encryptService:        encryptService,
	}
}

func (u *MessageUsecase) SendMessage(ctx context.Context, inputData dto.MessageSendRequest) (model.Message, bool, error) {
	var err error
	var response model.Message
	phoneNumber, err := u.phoneNumberRepository.GetByPhoneNumberID(ctx, inputData.PhoneNumberID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to get phone number:", err)
		return response, true, err
	}
	decyptedAccessToken, err := u.encryptService.Decrypt(phoneNumber.AccessToken)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to decrypt access token:", err)
		return response, true, err
	}
	whatsappClient := whatsapp_business.New(decyptedAccessToken, phoneNumber.WabaID, phoneNumber.PhoneNumberID)
	component, err := whatsapp_business_component.New(inputData.Type, inputData.Payload)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to validate message component:", err)
		return response, false, err
	}
	// create chat header if not exist
	chat := model.Chat{
		DocumentID:    fmt.Sprintf("%s-%s", inputData.RecipientID, inputData.PhoneNumberID),
		PhoneNumberID: inputData.PhoneNumberID,
		RecipientID:   inputData.RecipientID,
		ChatType:      "individual",
		LastMessage:   component.GetMessage(),
		DisplayName:   inputData.RecipientName,
		CreatedAt:     time.Now().Unix(),
		UpdatedAt:     time.Now().Unix(),
	}
	_, err = u.chatRepository.Insert(ctx, nil, chat)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to insert chat:", err)
		return response, true, err
	}
	sendResponse, httpCode, err := u.whatsappService.SendMessage(ctx, whatsappClient, inputData.RecipientID, component)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to send message:", err)
		return response, httpCode != http.StatusBadRequest, err
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
	response, err = u.messageRepository.Upsert(ctx, nil, message)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to upsert message:", err)
		return response, true, err
	}
	return response, false, nil
}

func (u *MessageUsecase) GetTemplateList(ctx context.Context, inputData dto.TemplateListRequest) ([]any, bool, error) {
	phoneNumber, err := u.phoneNumberRepository.GetByPhoneNumberID(ctx, inputData.PhoneNumberID)
	if err != nil {
		if err.Error() == "no more items in iterator" {
			log.Println("[ERROR][internal/usecase/message/message.go][GetTemplateList] Phone number not found:", err)
			return nil, false, fmt.Errorf("phone number not found")
		}
		log.Println("[ERROR][internal/usecase/message/message.go][GetTemplateList] Failed to get phone number:", err)
		return nil, true, err
	}
	decyptedAccessToken, err := u.encryptService.Decrypt(phoneNumber.AccessToken)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][GetTemplateList] Failed to decrypt access token:", err)
		return nil, true, err
	}
	whatsappClient := whatsapp_business.New(decyptedAccessToken, phoneNumber.WabaID, phoneNumber.PhoneNumberID)
	templateList, httpCode, err := u.whatsappService.GetTemplateList(ctx, whatsappClient)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][GetTemplateList] Failed to get template list:", err)
		return nil, httpCode != http.StatusBadRequest, err
	}
	return templateList, false, nil
}

func (u *MessageUsecase) GetMessagesByChatID(ctx context.Context, requestData filter_request.FilterRequest[dto.MessageGetByChatIDRequest]) (filter_request.FilterResponse[dto.MessageGetByChatIDResponse], bool, error) {
	var response filter_request.FilterResponse[dto.MessageGetByChatIDResponse]
	messages, err := u.messageRepository.GetMessageByChatID(ctx, requestData)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][GetMessagesByChatID] Failed to get messages by chat ID:", err)
		return response, true, err
	}
	return messages, false, nil
}
