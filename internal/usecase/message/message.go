package message_usecase

import (
	"context"
	"fmt"
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
	"wa_chat_service/pkg/meta/whatsapp_business"
	"wa_chat_service/pkg/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type MessageUsecase struct {
	messageRepository       repository.Message
	chatRepository          repository.Chat
	storageMediaRepository  repository.StorageMedia
	searchMessageRepository repository.SearchMessage
	storageMediaUsecase     usecase.StorageMedia
	tenantUsecase           usecase.Tenant
	whatsappService         service.WhatsappBusiness
	googleStorageService    service.GoogleStorage
	zslog                   *zap.SugaredLogger
}

func NewMessageUsecase(
	messageRepository repository.Message,
	chatRepository repository.Chat,
	storageMediaRepository repository.StorageMedia,
	searchMessageRepository repository.SearchMessage,
	storageMediaUsecase usecase.StorageMedia,
	tenantUsecase usecase.Tenant,
	whatsappService service.WhatsappBusiness,
	googleStorageService service.GoogleStorage,
	zslog *zap.SugaredLogger,
) *MessageUsecase {
	return &MessageUsecase{
		messageRepository:       messageRepository,
		chatRepository:          chatRepository,
		storageMediaRepository:  storageMediaRepository,
		storageMediaUsecase:     storageMediaUsecase,
		searchMessageRepository: searchMessageRepository,
		tenantUsecase:           tenantUsecase,
		whatsappService:         whatsappService,
		googleStorageService:    googleStorageService,
		zslog:                   zslog,
	}
}

func (u *MessageUsecase) SendMessage(ctx context.Context, whatsappClient *whatsapp_business.Client, tenantID string, inputData dto.MessageSendRequest) (model.Message, bool, error) {
	var err error
	var response model.Message
	if whatsappClient == nil {
		whatsappClient, tenantID, err = u.tenantUsecase.GetWhatsappClient(ctx, inputData.PhoneNumberID)
		if err != nil {
			u.zslog.Errorf("[SendMessage] Failed to get WhatsApp client: %v", err)
			return response, true, err
		}
	}
	component, err := whatsapp_business.NewComponent(inputData.Type, inputData.Payload)
	if err != nil {
		u.zslog.Errorf("[SendMessage] Failed to validate message component: %v", err)
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
		u.zslog.Errorf("[SendMessage] Failed to Upsert chat: %v", err)
		return response, true, err
	}
	var sto *model.StorageMedia
	if media := whatsapp_business.GetMedia(component); media != nil {
		if media.Link != nil {
			isValid, err := u.googleStorageService.IsValidSignedURL(ctx, *media.Link)
			if err != nil {
				u.zslog.Errorf("[SendMessage] Failed to validate media link: %v", err)
				return response, true, err
			}
			if isValid {
				// get file URL from signed URL
				fileURL, err := u.googleStorageService.ParseSignedURLToFileURL(ctx, *media.Link)
				if err != nil {
					u.zslog.Errorf("[SendMessage] Failed to parse signed URL to file URL: %v", err)
					return response, true, err
				}
				_, filePath, err := u.googleStorageService.ParseGoogleStorageURL(fileURL)
				if err != nil {
					u.zslog.Errorf("[SendMessage] Failed to parse file URL: %v", err)
					return response, true, err
				}
				fileName := filePath[strings.LastIndex(filePath, "/")+1:]
				fileNameWithoutExt := fileName[:strings.LastIndex(fileName, ".")]
				storageMedia, err := u.storageMediaRepository.GetByDocumentID(ctx, fileNameWithoutExt)
				if err != nil {
					u.zslog.Errorf("[SendMessage] Failed to get storage media by document ID: %v", err)
					return response, true, err
				}
				sto = &storageMedia
			} else {
				// check if link is accessible
				resp, err := http.Head(*media.Link)
				if err != nil || resp.StatusCode != http.StatusOK {
					u.zslog.Errorf("[SendMessage] Media link is not accessible: %v", err)
					return response, false, fmt.Errorf("media link is not accessible")
				}
				urlHeaders := resp.Header
				mimeType := urlHeaders.Get("Content-Type")
				extension := whatsapp_business.ParseMediaExtension(mimeType)
				if extension == "" {
					u.zslog.Errorf("[SendMessage] Unsupported media type: %v", mimeType)
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
					originalFileName = utils.GetFileNameFromURL(*media.Link)
				}
				newStorageMediaID, err := uuid.NewV7()
				if err != nil {
					u.zslog.Errorf("[SendMessage] Failed to generate storage media ID: %v", err)
					return response, true, err
				}
				if originalFileName == "" {
					originalFileName = fmt.Sprintf("%s%s", newStorageMediaID.String(), whatsapp_business.ParseMediaExtension(mimeType))
				}
				storageMedia := model.StorageMedia{
					DocumentID:       newStorageMediaID.String(),
					TenantID:         tenantID,
					OriginalName:     originalFileName,
					URL:              media.Link,
					IsURLFromStorage: false,
					MimeType:         mimeType,
					CreatedAt:        time.Now(),
				}
				_, err = u.storageMediaRepository.Insert(ctx, nil, storageMedia)
				if err != nil {
					u.zslog.Errorf("[SendMessage] Failed to create storage media: %v", err)
					return response, true, err
				}
				sto = &storageMedia
			}
		}
	}
	sendResponse, httpCode, err := whatsappClient.SendMessage(inputData.RecipientID, "individual", component)
	if err != nil {
		u.zslog.Errorf("[SendMessage] Failed to send message: %v", err)
		return response, httpCode >= http.StatusInternalServerError, err
	}
	payloadData, err := utils.AnyToJsonString(component.GetPayload())
	if err != nil {
		u.zslog.Errorf("[SendMessage] Failed to convert payload to JSON")
	}
	var storageMediaID *string
	if sto != nil {
		storageMediaID = &sto.DocumentID
	}
	messageID, err := uuid.NewV7()
	if err != nil {
		u.zslog.Errorf("[SendMessage] Failed to generate message ID: %v", err)
		return response, true, err
	}
	message := model.Message{
		DocumentID:      messageID.String(),
		Wamid:           sendResponse.Messages[0].ID,
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
		u.zslog.Errorf("[SendMessage] Failed to upsert message: %v", err)
		return response, true, err
	}
	err = u.searchMessageRepository.AddDocuments(ctx, []model.Message{response})
	if err != nil {
		u.zslog.Errorf("[SendMessage] Failed to add message to search index: %v", err)
	}

	return response, false, nil
}

func (u *MessageUsecase) GetMessagesByChatID(ctx context.Context, requestData filter_request.FilterRequest[dto.MessageGetByChatIDRequest]) (filter_request.FilterResponse[dto.MessageResponse], bool, error) {
	var response filter_request.FilterResponse[dto.MessageResponse]
	messages, totalItems, paginate, err := u.searchMessageRepository.GetFiltered(ctx, requestData)
	if err != nil {
		u.zslog.Errorf("[GetMessagesByChatID] Failed to get messages by chat ID: %v", err)
		return response, true, err
	}
	response.Page = paginate.Page
	response.PageSize = paginate.PageSize
	response.TotalItems = totalItems
	response.TotalPages = (totalItems + int64(paginate.PageSize) - 1) / int64(paginate.PageSize)

	var results []dto.MessageResponse
	if len(messages) != 0 {
		// get storage media for messages
		var storageMediaIds []string
		var storageMediaMap map[string]model.StorageMedia
		for _, message := range messages {
			if message.StorageMediaID != nil {
				storageMediaIds = append(storageMediaIds, *message.StorageMediaID)
			}
		}
		if len(storageMediaIds) > 0 {
			storageMediaMap, err = u.storageMediaRepository.GetByDocumentIDs(ctx, storageMediaIds)
			if err != nil {
				u.zslog.Errorf("[GetMessagesByChatID] Failed to get storage medias by IDs: %v", err)
				return response, true, err
			}
		}
		for _, message := range messages {
			var storageMediaResponse *dto.StorageMediaResponse
			if message.StorageMediaID != nil {
				storageMedia, ok := storageMediaMap[*message.StorageMediaID]
				if ok {
					var accessURL *string
					if storageMedia.IsURLFromStorage {
						url, err := u.googleStorageService.GenerateV4GetObjectSignedURL(*storageMedia.URL, 0)
						if err != nil {
							u.zslog.Errorf("[GetMessagesByChatID] Failed to generate signed URL: %v", err)
						} else {
							accessURL = &url
						}
					}
					media := dto.StorageMediaResponse{}.FromModel(storageMedia, accessURL)
					storageMediaResponse = &media
				}
			}
			results = append(results, dto.MessageResponse{}.FromModel(message, storageMediaResponse))
		}
	}
	response.Results = results
	return response, false, nil
}

func (u *MessageUsecase) SaveMessage(ctx context.Context, inputData dto.MessageSaveRequest) (bool, error) {
	message := model.Message{
		Wamid:           inputData.Wamid,
		ChatID:          inputData.ChatID,
		MessageType:     inputData.MessageType,
		MessageCategory: inputData.MessageCategory,
		SenderName:      inputData.SenderName,
		Payload:         inputData.Payload,
		StorageMediaID:  inputData.StorageMediaID,
		Status:          inputData.Status,
		Error:           inputData.Error,
		CreatedAt:       inputData.CreatedAt,
		SentAt:          inputData.SentAt,
		DeliveredAt:     inputData.DeliveredAt,
		ReadAt:          inputData.ReadAt,
	}
	if inputData.ID != nil {
		message.DocumentID = *inputData.ID
	}
	_, err := u.messageRepository.Upsert(ctx, nil, message)
	if err != nil {
		u.zslog.Errorf("[SaveMessage] Failed to save message: %v", err)
		return true, err
	}
	err = u.searchMessageRepository.AddDocuments(ctx, []model.Message{message})
	if err != nil {
		u.zslog.Errorf("[SaveMessage] Failed to add message to search index: %v", err)
	}
	return false, nil
}
