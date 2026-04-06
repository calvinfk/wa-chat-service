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
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/formatter"
	"wa_chat_service/pkg/meta/whatsapp_business"
	whatsapp_business_component "wa_chat_service/pkg/meta/whatsapp_business/component"
)

type MessageUsecase struct {
	messageRepository      repository.Message
	chatRepository         repository.Chat
	phoneNumberRepository  repository.PhoneNumber
	storageMediaRepository repository.StorageMedia
	storageMediaUsecase    usecase.StorageMedia
	whatsappService        service.WhatsappService
	encryptService         service.Encrypt
	googleFirebaseService  service.GoogleFirebase
}

func NewMessageUsecase(
	messageRepository repository.Message,
	chatRepository repository.Chat,
	phoneNumberRepository repository.PhoneNumber,
	storageMediaRepository repository.StorageMedia,
	storageMediaUsecase usecase.StorageMedia,
	whatsappService service.WhatsappService,
	encryptService service.Encrypt,
	googleFirebaseService service.GoogleFirebase,
) *MessageUsecase {
	return &MessageUsecase{
		messageRepository:      messageRepository,
		chatRepository:         chatRepository,
		phoneNumberRepository:  phoneNumberRepository,
		storageMediaRepository: storageMediaRepository,
		storageMediaUsecase:    storageMediaUsecase,
		whatsappService:        whatsappService,
		encryptService:         encryptService,
		googleFirebaseService:  googleFirebaseService,
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
	_, err = u.chatRepository.Upsert(ctx, nil, chat)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to Upsert chat:", err)
		return response, true, err
	}
	var sto *model.StorageMedia
	if media := whatsapp_business_component.GetMedia(component); media != nil {
		if media.Link != nil {
			storedMedia, err := u.storageMediaRepository.GetByAccessURL(ctx, *media.Link)
			if err == nil {
				sto = &storedMedia
			} else {
				newMedia, serverError, err := u.storageMediaUsecase.StoreMediaFromURL(ctx, *media.Link)
				if err != nil {
					log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to handle media by URL:", err)
					return response, serverError, err
				}
				sto = &newMedia
			}
		} else if media.ID != nil {
			// check if media with the given media ID already exists in storage before attempting to download from Meta
			storedMedia, err := u.storageMediaRepository.GetByMediaID(ctx, *media.ID)
			if err == nil {
				sto = &storedMedia
			} else {
				// download from meta then upload to firebase storage
				downloadURL, httpCode, err := u.whatsappService.GetMediaURL(ctx, whatsappClient, *media.ID)
				if err != nil {
					log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to get media download URL:", err)
					return response, true, err
				}
				if httpCode != http.StatusOK {
					log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to get media download URL, HTTP code:", httpCode)
					return response, true, fmt.Errorf("failed to get media download URL, HTTP code: %d", httpCode)
				}
				newMedia, serverError, err := u.storageMediaUsecase.StoreMediaFromURL(ctx, downloadURL)
				if err != nil {
					log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to download media from Meta:", err)
					return response, serverError, err
				}
				sto = &newMedia
			}
		}
	}
	sendResponse, httpCode, err := u.whatsappService.SendMessage(ctx, whatsappClient, inputData.RecipientID, component)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to send message:", err)
		return response, httpCode >= http.StatusInternalServerError, err
	}
	payloadData, err := formatter.AnyToJsonString(component.GetPayload())
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to convert payload to JSON")
	}
	var storageMediaID *string
	if sto != nil {
		storageMediaID = &sto.DocumentID
	}
	message := model.Message{
		DocumentID:      sendResponse.Messages[0].ID,
		ChatID:          chat.DocumentID,
		MessageType:     string(component.GetType()),
		MessageCategory: "-",
		SenderName:      inputData.SenderName,
		Payload:         payloadData,
		StorageMediaID:  storageMediaID,
		StorageMedia:    sto,
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
			return nil, false, errs.ErrGenericNotFound
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
		return nil, httpCode >= http.StatusInternalServerError, err
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
