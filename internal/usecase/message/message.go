package message_usecase

import (
	"context"
	"fmt"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/formatter"
	"wa_chat_service/pkg/meta/whatsapp_business"
	whatsapp_business_component "wa_chat_service/pkg/meta/whatsapp_business/component"

	"github.com/google/uuid"
)

type MessageUsecase struct {
	messageRepository      repository.Message
	chatRepository         repository.Chat
	phoneNumberRepository  repository.PhoneNumber
	storageMediaRepository repository.StorageMedia
	whatsappService        service.WhatsappService
	encryptService         service.Encrypt
	googleFirebaseService  service.GoogleFirebase
}

func NewMessageUsecase(
	messageRepository repository.Message,
	chatRepository repository.Chat,
	phoneNumberRepository repository.PhoneNumber,
	storageMediaRepository repository.StorageMedia,
	whatsappService service.WhatsappService,
	encryptService service.Encrypt,
	googleFirebaseService service.GoogleFirebase,
) *MessageUsecase {
	return &MessageUsecase{
		messageRepository:      messageRepository,
		chatRepository:         chatRepository,
		phoneNumberRepository:  phoneNumberRepository,
		storageMediaRepository: storageMediaRepository,
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
	var storageMediaID *string
	var sto *model.StorageMedia
	if media := whatsapp_business_component.GetMedia(component); media != nil {
		if media.Link != nil {
			storedMedia, err := u.storageMediaRepository.GetByAccessURL(ctx, *media.Link)
			if err == nil {
				storageMediaID = &storedMedia.DocumentID
				sto = &storedMedia
			} else {
				// create new storage media record with original media link as access URL
				newMediaID, err := uuid.NewV7()
				if err != nil {
					log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to generate new media ID:", err)
					return response, true, err
				}
				// download the file
				fileData, urlHeaders, err := formatter.DownloadFile(*media.Link)
				if err != nil {
					log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to download media file:", err)
					return response, true, err
				}
				// upload to firebase storage
				var filename string
				contentDisposition := urlHeaders.Get("Content-Disposition")
				if contentDisposition == "" {
					// check the url path for filename if Content-Disposition header is not present
					urlParts := strings.Split(*media.Link, "/")
					if len(urlParts) > 0 {
						urlParts[len(urlParts)-1] = strings.Split(urlParts[len(urlParts)-1], "?")[0] // remove query params
						// check if the last part of the URL path has a valid filename format (e.g., has an extension)
						if strings.Contains(urlParts[len(urlParts)-1], ".") {
							filename = urlParts[len(urlParts)-1]
						}
					}
				} else {
					// extract filename from Content-Disposition header
					_, params, err := mime.ParseMediaType(contentDisposition)
					if err == nil {
						filename = params["filename"]
					}
				}
				if filename == "" {
					filename = fmt.Sprintf("%s.%s", newMediaID.String(), strings.Split(urlHeaders.Get("Content-Type"), "/")[1]) // default filename if not provided
				}
				filePath := "whatsapp-media/" + newMediaID.String() + filepath.Ext(filename)
				url, attrs, err := u.googleFirebaseService.UploadFile(ctx, filePath, fileData)
				if err != nil {
					log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to upload media file to storage:", err)
					return response, true, err
				}
				newMedia := model.StorageMedia{
					DocumentID:   newMediaID.String(),
					OriginalName: filename,
					MimeType:     attrs.ContentType,
					URL:          url,
					CreatedAt:    time.Now().Unix(),
				}
				_, err = u.storageMediaRepository.Insert(ctx, nil, newMedia)
				if err != nil {
					log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to insert new storage media record:", err)
				}
				storageMediaID = &newMedia.DocumentID
				sto = &newMedia
				log.Println("[INFO][internal/usecase/message/message.go][SendMessage] Successfully stored media with ID:", newMedia.DocumentID)
			}
		} else if media.ID != nil {
			// check if media with the given media ID already exists in storage before attempting to download from Meta
			storedMedia, err := u.storageMediaRepository.GetByMediaID(ctx, *media.ID)
			if err == nil {
				storageMediaID = &storedMedia.DocumentID
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
				// proceed with the download and upload logic
				fileData, urlHeaders, err := formatter.DownloadFile(downloadURL)
				if err != nil {
					log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to download media file from Meta:", err)
					return response, true, err
				}
				// upload to firebase storage
				newMediaID, err := uuid.NewV7()
				if err != nil {
					log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to generate new media ID:", err)
					return response, true, err
				}
				mimeType := urlHeaders.Get("Content-Type")
				fileExtension := whatsapp_business.ParseMediaExtension(mimeType)
				var filename string
				contentDisposition := urlHeaders.Get("Content-Disposition")
				if contentDisposition != "" {
					// extract filename from Content-Disposition header
					_, params, err := mime.ParseMediaType(contentDisposition)
					if err == nil {
						filename = params["filename"]
					}
				}
				if filename == "" {
					filename = fmt.Sprintf("%s.%s", newMediaID.String(), fileExtension) // default filename
				}
				filePath := "whatsapp-media/" + filename
				url, _, err := u.googleFirebaseService.UploadFile(ctx, filePath, fileData)
				if err != nil {
					log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to upload media file to storage:", err)
					return response, true, err
				}
				newMedia := model.StorageMedia{
					DocumentID:   newMediaID.String(),
					OriginalName: filename,
					MimeType:     mimeType,
					URL:          url,
					MediaID:      media.ID,
					CreatedAt:    time.Now().Unix(),
				}
				_, err = u.storageMediaRepository.Insert(ctx, nil, newMedia)
				if err != nil {
					log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to insert new storage media record:", err)
				}
				storageMediaID = &newMedia.DocumentID
				sto = &newMedia
				log.Println("[INFO][internal/usecase/message/message.go][SendMessage] Successfully stored media with ID:", newMedia.DocumentID)
			}
		}
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
