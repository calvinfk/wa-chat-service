package message_usecase

import (
	"context"
	"fmt"
	"log"
	"mime"
	"net/http"
	"strings"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/formatter"
	"wa_chat_service/pkg/meta/whatsapp_business"

	"github.com/google/uuid"
)

type MessageUsecase struct {
	messageRepository      repository.Message
	chatRepository         repository.Chat
	storageMediaRepository repository.StorageMedia
	storageMediaUsecase    usecase.StorageMedia
	phoneNumberUsecase     usecase.PhoneNumber
	whatsappService        service.WhatsappBusiness
	googleStorageService   service.GoogleStorage
}

func NewMessageUsecase(
	messageRepository repository.Message,
	chatRepository repository.Chat,
	storageMediaRepository repository.StorageMedia,
	storageMediaUsecase usecase.StorageMedia,
	phoneNumberUsecase usecase.PhoneNumber,
	whatsappService service.WhatsappBusiness,
	googleStorageService service.GoogleStorage,
) *MessageUsecase {
	return &MessageUsecase{
		messageRepository:      messageRepository,
		chatRepository:         chatRepository,
		storageMediaRepository: storageMediaRepository,
		storageMediaUsecase:    storageMediaUsecase,
		phoneNumberUsecase:     phoneNumberUsecase,
		whatsappService:        whatsappService,
		googleStorageService:   googleStorageService,
	}
}

func (u *MessageUsecase) SendMessage(ctx context.Context, inputData dto.MessageSendRequest) (model.Message, bool, error) {
	var err error
	var response model.Message
	whatsappClient, err := u.phoneNumberUsecase.GetWhatsappClient(ctx, inputData.PhoneNumberID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to get WhatsApp client:", err)
		return response, true, err
	}
	component, err := whatsapp_business.NewComponent(inputData.Type, inputData.Payload)
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
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	_, err = u.chatRepository.Upsert(ctx, nil, chat)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to Upsert chat:", err)
		return response, true, err
	}
	var sto *model.StorageMedia
	if media := whatsapp_business.GetMedia(component); media != nil {
		if media.Link != nil {
			isValid, err := u.googleStorageService.IsValidSignedURL(ctx, *media.Link)
			if err != nil {
				log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to validate media link:", err)
				return response, true, err
			}
			if isValid {
				// get file URL from signed URL
				fileURL, err := u.googleStorageService.ParseSignedURLToFileURL(ctx, *media.Link)
				if err != nil {
					log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to parse signed URL to file URL:", err)
					return response, true, err
				}
				_, filePath, err := u.googleStorageService.ParseGoogleStorageURL(fileURL)
				if err != nil {
					log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to parse file URL:", err)
					return response, true, err
				}
				fileName := filePath[strings.LastIndex(filePath, "/")+1:]
				fileNameWithoutExt := fileName[:strings.LastIndex(fileName, ".")]
				storageMedia, err := u.storageMediaRepository.GetByDocumentID(ctx, fileNameWithoutExt)
				if err != nil {
					log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to get storage media by document ID:", err)
					return response, true, err
				}
				sto = &storageMedia
			} else {
				// check if link is accessible
				resp, err := http.Head(*media.Link)
				if err != nil || resp.StatusCode != http.StatusOK {
					log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Media link is not accessible:", err)
					return response, true, fmt.Errorf("media link is not accessible")
				}
				urlHeaders := resp.Header
				mimeType := urlHeaders.Get("Content-Type")
				extension := whatsapp_business.ParseMediaExtension(mimeType)
				if extension == "" {
					log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Unsupported media type:", mimeType)
					return response, true, fmt.Errorf("unsupported media type: %s", mimeType)
				}
				// TODO: check file size is allowed or not
				var originalFileName string
				contentDisposition := urlHeaders.Get("Content-Disposition")
				if contentDisposition != "" {
					_, params, err := mime.ParseMediaType(contentDisposition)
					if err == nil {
						originalFileName = params["filename"]
					}
				} else {
					originalFileName = formatter.GetFileNameFromURL(*media.Link)
				}
				newStorageMediaID, err := uuid.NewV7()
				if err != nil {
					log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to generate storage media ID:", err)
					return response, true, err
				}
				if originalFileName == "" {
					originalFileName = fmt.Sprintf("%s%s", newStorageMediaID.String(), whatsapp_business.ParseMediaExtension(mimeType))
				}
				storageMedia := model.StorageMedia{
					DocumentID:       newStorageMediaID.String(),
					OriginalName:     originalFileName,
					URL:              media.Link,
					IsURLFromStorage: false,
					MimeType:         mimeType,
					CreatedAt:        time.Now(),
				}
				_, err = u.storageMediaRepository.Insert(ctx, nil, storageMedia)
				if err != nil {
					log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to create storage media:", err)
					return response, true, err
				}
				sto = &storageMedia
			}
		}
	}
	sendResponse, httpCode, err := u.whatsappService.SendMessage(whatsappClient, inputData.RecipientID, component)
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
		CreatedAt:       time.Now(),
	}
	response, err = u.messageRepository.Upsert(ctx, nil, message)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to upsert message:", err)
		return response, true, err
	}
	return response, false, nil
}

func (u *MessageUsecase) GetTemplateList(ctx context.Context, inputData dto.TemplateListRequest) ([]whatsapp_business.TemplateResponse, bool, error) {
	whatsappClient, err := u.phoneNumberUsecase.GetWhatsappClient(ctx, inputData.PhoneNumberID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][GetTemplateList] Failed to get WhatsApp client:", err)
		return nil, true, err
	}
	templateList, httpCode, err := u.whatsappService.GetTemplateList(whatsappClient)
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
