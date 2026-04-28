package message_usecase

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/meta/whatsapp_business"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type MessageUsecase struct {
	messageRepository        repository.Message
	chatRepository           repository.Chat
	storageMediaRepository   repository.StorageMedia
	searchMessageRepository  repository.SearchMessage
	tenantRepository         repository.Tenant
	storageMediaUsecase      usecase.StorageMedia
	waBusinessAccountUsecase usecase.WaBusinessAccount
	whatsappService          service.WhatsappBusiness
	googleStorageService     service.GoogleStorage
	txManager                *utils.TxManager
	zsLog                    *zap.SugaredLogger
}

func NewMessageUsecase(
	messageRepository repository.Message,
	chatRepository repository.Chat,
	storageMediaRepository repository.StorageMedia,
	searchMessageRepository repository.SearchMessage,
	tenantRepository repository.Tenant,
	storageMediaUsecase usecase.StorageMedia,
	waBusinessAccountUsecase usecase.WaBusinessAccount,
	whatsappService service.WhatsappBusiness,
	googleStorageService service.GoogleStorage,
	txManager *utils.TxManager,
	zsLog *zap.SugaredLogger,
) *MessageUsecase {
	return &MessageUsecase{
		messageRepository:        messageRepository,
		chatRepository:           chatRepository,
		storageMediaRepository:   storageMediaRepository,
		tenantRepository:         tenantRepository,
		searchMessageRepository:  searchMessageRepository,
		storageMediaUsecase:      storageMediaUsecase,
		waBusinessAccountUsecase: waBusinessAccountUsecase,
		whatsappService:          whatsappService,
		googleStorageService:     googleStorageService,
		txManager:                txManager,
		zsLog:                    zsLog,
	}
}

func (u *MessageUsecase) SendMessage(ctx context.Context, whatsappClient *whatsapp_business.Client, tenantID string, inputData dto.MessageSendRequest) (model.Message, bool, error) {
	var err error
	var response model.Message
	if whatsappClient == nil {
		whatsappClient, _, err = u.waBusinessAccountUsecase.GetWhatsappClient(ctx, tenantID, inputData.PhoneNumberId)
		if err != nil {
			u.zsLog.Errorf("[SendMessage] Failed to get WhatsApp client: %v", err)
			return response, true, err
		}
	}
	tenant, err := u.tenantRepository.GetByID(ctx, tenantID)
	if err != nil {
		u.zsLog.Errorf("[SendMessage] Failed to get tenant: %v", err)
		return response, true, err
	}
	component, err := whatsapp_business.NewComponent(inputData.Type, inputData.Payload)
	if err != nil {
		u.zsLog.Errorf("[SendMessage] Failed to validate message component: %v", err)
		return response, false, err
	}
	// TODO: save ticketing
	// create chat header if not exist
	chat := model.Chat{
		PhoneNumberId: inputData.PhoneNumberId,
		RecipientId:   inputData.RecipientId,
		LastMessage:   component.GetMessage(),
		DisplayName:   inputData.RecipientName,
		ChatStatus:    model.ChatStatusOpen,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if tenant.ChatType == "ticket" {
		chatID, err := uuid.NewV7()
		if err != nil {
			u.zsLog.Errorf("[SendMessage] Failed to generate chat ID: %v", err)
			return response, true, err
		}
		chat.DocumentID = chatID.String()
		chat.ChatType = "ticket"
	} else {
		chat.DocumentID = fmt.Sprintf("%s-%s", inputData.RecipientId, inputData.PhoneNumberId)
		chat.ChatType = "individual"
	}
	_, err = u.chatRepository.Upsert(ctx, nil, chat)
	if err != nil {
		u.zsLog.Errorf("[SendMessage] Failed to Upsert chat: %v", err)
		return response, true, err
	}
	var sto *model.StorageMedia
	if media := whatsapp_business.GetMedia(component); media != nil {
		if media.Link != nil {
			mediaToken, err := u.storageMediaUsecase.ParsePublicURL(*media.Link)
			if err == nil {
				// get file URL from signed URL
				storageMediaID, serverError, err := u.storageMediaUsecase.ParseMediaToken(mediaToken)
				if err != nil {
					u.zsLog.Errorf("[SendMessage] Failed to parse media token: %v", err)
					return response, serverError, err
				}
				storageMedia, err := u.storageMediaRepository.GetByDocumentID(ctx, storageMediaID)
				if err != nil {
					u.zsLog.Errorf("[SendMessage] Failed to get storage media by document ID: %v", err)
					return response, true, err
				}
				sto = &storageMedia
			} else {
				// check if link is accessible
				resp, err := http.Head(*media.Link)
				if err != nil || resp.StatusCode != http.StatusOK {
					u.zsLog.Errorf("[SendMessage] Media link is not accessible: %v", err)
					return response, false, fmt.Errorf("media link is not accessible")
				}
				urlHeaders := resp.Header
				mimeType := urlHeaders.Get("Content-Type")
				extension := whatsapp_business.ParseMediaExtension(mimeType)
				if extension == "" {
					u.zsLog.Errorf("[SendMessage] Unsupported media type: %v", mimeType)
					return response, true, fmt.Errorf("unsupported media type: %s", mimeType)
				}
				// TODO: check file size is allowed or not
				originalFileName := utils.GetFileNameFromURL(urlHeaders, *media.Link)
				newStorageMediaID, err := uuid.NewV7()
				if err != nil {
					u.zsLog.Errorf("[SendMessage] Failed to generate storage media ID: %v", err)
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
					u.zsLog.Errorf("[SendMessage] Failed to create storage media: %v", err)
					return response, true, err
				}
				sto = &storageMedia
			}
		}
	}
	sendResponse, httpCode, err := whatsappClient.SendMessage(inputData.RecipientId, "individual", component)
	if err != nil {
		u.zsLog.Errorf("[SendMessage] Failed to send message: %v", err)
		return response, httpCode >= http.StatusInternalServerError, err
	}
	payloadData, err := utils.AnyToJsonString(component.GetPayload())
	if err != nil {
		u.zsLog.Errorf("[SendMessage] Failed to convert payload to JSON")
	}
	var storageMediaID *string
	if sto != nil {
		storageMediaID = &sto.DocumentID
	}
	messageID, err := uuid.NewV7()
	if err != nil {
		u.zsLog.Errorf("[SendMessage] Failed to generate message ID: %v", err)
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
		u.zsLog.Errorf("[SendMessage] Failed to upsert message: %v", err)
		return response, true, err
	}
	err = u.searchMessageRepository.AddDocuments(ctx, []model.Message{response})
	if err != nil {
		u.zsLog.Errorf("[SendMessage] Failed to add message to search index: %v", err)
	}

	return response, false, nil
}

func (u *MessageUsecase) GetMessagesByChatID(ctx context.Context, tenantID string, requestData filter_request.FilterRequest[dto.MessageGetByChatIDRequest]) (filter_request.FilterResponse[dto.MessageResponse], bool, error) {
	// TODO: check if chat belongs to tenant
	var response filter_request.FilterResponse[dto.MessageResponse]
	messages, totalItems, paginate, err := u.searchMessageRepository.GetFiltered(ctx, requestData)
	if err != nil {
		u.zsLog.Errorf("[GetMessagesByChatID] Failed to get messages by chat ID: %v", err)
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
				u.zsLog.Errorf("[GetMessagesByChatID] Failed to get storage medias by IDs: %v", err)
				return response, true, err
			}
		}
		for _, message := range messages {
			var storageMediaResponse *dto.StorageMediaResponse
			if message.StorageMediaID != nil {
				storageMedia, ok := storageMediaMap[*message.StorageMediaID]
				if ok {
					var accessURL *string
					if storageMedia.URL != nil || storageMedia.MediaId != nil {
						url, err := u.storageMediaUsecase.GeneratePublicURL(storageMedia.DocumentID)
						if err != nil {
							u.zsLog.Errorf("[GetMessagesByChatID] Failed to get access URL for storage media ID %s: %v", storageMedia.DocumentID, err)
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

func (u *MessageUsecase) SaveMessage(ctx context.Context, tenantID string, inputData dto.MessageSaveRequest) (bool, error) {
	// TODO: check if chat belongs to tenant
	// check tenant chat type
	tenant, err := u.tenantRepository.GetByID(ctx, tenantID)
	if err != nil {
		u.zsLog.Errorf("[SaveMessage] Failed to get tenant: %v", err)
		return true, err
	}
	var chatID string
	var chat model.Chat

	switch tenant.ChatType {
	case "ticket":
		chat, err = u.chatRepository.GetOpenedTicketChatByPhoneNumberID(ctx, inputData.PhoneNumberId, inputData.RecipientId)
		switch err {
		case nil:
			chatID = chat.DocumentID
		case errs.ErrGenericNotFound:
			newChatID, genErr := uuid.NewV7()
			if genErr != nil {
				u.zsLog.Errorf("[SaveMessage] Failed to generate chat ID: %v", genErr)
				return true, genErr
			}
			chatID = newChatID.String()
		default:
			u.zsLog.Errorf("[SaveMessage] Failed to get opened ticket chat: %v", err)
			return true, err
		}
	default:
		chatID = fmt.Sprintf("%s-%s", inputData.RecipientId, inputData.PhoneNumberId)
	}

	if chat.DocumentID == "" {
		// create new chat if not exist
		chat = model.Chat{
			DocumentID:    chatID,
			PhoneNumberId: inputData.PhoneNumberId,
			RecipientId:   inputData.RecipientId,
			DisplayName:   inputData.DisplayName,
			ChatStatus:    model.ChatStatusOpen,
			CreatedAt:     inputData.CreatedAt,
		}
	}
	// if previous chat type is empty, set chat type based on tenant chat type
	if chat.ChatType == "" {
		if tenant.ChatType == "ticket" {
			chat.ChatType = "ticket"
		} else {
			chat.ChatType = "individual"
		}
	}
	if chat.DisplayName == "" {
		chat.DisplayName = inputData.RecipientId
	}
	chat.LastMessage = inputData.LastMessage
	chat.UpdatedAt = time.Now()

	serverError, err := u.txManager.DoFirestore(ctx, func(ctx context.Context, txFirestore *firestore.Transaction) (bool, error) {
		_, err = u.chatRepository.Upsert(ctx, txFirestore, chat)
		if err != nil {
			u.zsLog.Errorf("[SaveMessage] Failed to upsert chat: %v", err)
			return true, err
		}

		message := model.Message{
			Wamid:           inputData.Wamid,
			ChatID:          chatID,
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
		_, err = u.messageRepository.Upsert(ctx, txFirestore, message)
		if err != nil {
			u.zsLog.Errorf("[SaveMessage] Failed to save message: %v", err)
			return true, err
		}
		err = u.searchMessageRepository.AddDocuments(ctx, []model.Message{message})
		if err != nil {
			u.zsLog.Errorf("[SaveMessage] Failed to add message to search index: %v", err)
		}
		return false, nil
	})
	return serverError, err
}
