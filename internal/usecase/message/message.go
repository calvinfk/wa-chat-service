package message_usecase

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
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
	userRepository           repository.User
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
	userRepository repository.User,
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
		userRepository:           userRepository,
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
	tenant, err := u.tenantRepository.GetByID(ctx, tenantID)
	if err != nil {
		u.zsLog.Errorf("[SendMessage] Failed to get tenant: %v", err)
		return response, true, err
	}
	var chatExists bool
	chat, err := u.chatRepository.GetByID(ctx, inputData.ChatID)
	if err != nil && err != errs.ErrGenericNotFound {
		u.zsLog.Errorf("[SendMessage] Failed to get chat by ID: %v", err)
		return response, true, err
	}
	var phoneNumberId string
	var recipientId string
	var recipientName string

	if err != nil && err == errs.ErrGenericNotFound {
		if tenant.ChatType == "ticket" {
			u.zsLog.Errorf("[SendMessage] No opened ticket chat found for ID %s: %v", inputData.ChatID, err)
			return response, false, fmt.Errorf("no opened ticket chat found for ID %s", inputData.ChatID)
		}
		// check if id is in format {recipient_id}-{phone_number_id} for individual chat
		n, scanErr := fmt.Sscanf(inputData.ChatID, "%s-%s", &recipientId, &phoneNumberId)
		if scanErr != nil || n != 2 {
			u.zsLog.Errorf("[SendMessage] Chat ID %s is not in valid format: %v", inputData.ChatID, scanErr)
			return response, false, fmt.Errorf("chat ID %s is not in valid format", inputData.ChatID)
		}
		// check if id is numeric
		if _, err := strconv.Atoi(recipientId); err != nil {
			u.zsLog.Errorf("[SendMessage] Recipient ID %s is not numeric: %v", recipientId, err)
			return response, false, fmt.Errorf("recipient ID %s is not numeric", recipientId)
		}
		if _, err := strconv.Atoi(phoneNumberId); err != nil {
			u.zsLog.Errorf("[SendMessage] Phone number ID %s is not numeric: %v", phoneNumberId, err)
			return response, false, fmt.Errorf("phone number ID %s is not numeric", phoneNumberId)
		}
	} else {
		phoneNumberId = chat.PhoneNumberId
		recipientId = chat.RecipientId
		recipientName = chat.RecipientName
		chatExists = true
	}
	if whatsappClient == nil {
		whatsappClient, _, err = u.waBusinessAccountUsecase.GetWhatsappClient(ctx, tenantID, chat.PhoneNumberId)
		if err != nil {
			u.zsLog.Errorf("[SendMessage] Failed to get WhatsApp client: %v", err)
			return response, true, err
		}
	}
	component, err := whatsapp_business.NewComponent(inputData.Type, inputData.Payload)
	if err != nil {
		u.zsLog.Errorf("[SendMessage] Failed to validate message component: %v", err)
		return response, false, err
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
				storageMedia, err := u.storageMediaRepository.GetByID(ctx, storageMediaID)
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
				_, err = u.storageMediaRepository.Upsert(ctx, nil, storageMedia)
				if err != nil {
					u.zsLog.Errorf("[SendMessage] Failed to upsert storage media: %v", err)
					return response, true, err
				}
				sto = &storageMedia
			}
		}
	}
	sendResponse, httpCode, err := whatsappClient.SendMessage(chat.RecipientId, "individual", component)
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

	// create a new chat for the message if chat does not exist and associate the message with the new chat
	var chatID *string
	if chatExists {
		chatID = &chat.DocumentID
	}
	serverError, err := u.SaveMessage(ctx, tenantID, dto.MessageSaveRequest{
		ChatID:          chatID,
		Wamid:           sendResponse.Messages[0].ID,
		PhoneNumberId:   phoneNumberId,
		RecipientId:     recipientId,
		RecipientName:   recipientName,
		LastMessage:     component.GetMessage(),
		MessageType:     string(component.GetType()),
		MessageCategory: "-",
		SenderName:      inputData.SenderName,
		Payload:         payloadData,
		StorageMediaID:  storageMediaID,
		Status:          "-",
		CreatedAt:       time.Now(),
	})
	if err != nil {
		u.zsLog.Errorf("[SendMessage] Failed to save message: %v", err)
		return response, serverError, err
	}

	return response, false, nil
}

func (u *MessageUsecase) GetMessagesByChatID(ctx context.Context, authData dto.AuthData, requestData filter_request.FilterRequest[dto.MessageGetByChatIDRequest]) (filter_request.FilterResponse[dto.MessageResponse], bool, error) {
	chat, err := u.chatRepository.GetByID(ctx, requestData.SpecificFilter.ChatID)
	if err != nil {
		u.zsLog.Errorf("[GetMessagesByChatID] Failed to get chat by ID: %v", err)
		return filter_request.FilterResponse[dto.MessageResponse]{}, true, err
	}
	if chat.TenantID != authData.TenantID {
		u.zsLog.Warnf("[GetMessagesByChatID] WhatsApp business tenant ID %s does not match auth tenant ID %s", chat.TenantID, authData.TenantID)
		return filter_request.FilterResponse[dto.MessageResponse]{}, false, fmt.Errorf("no chat found for ID %s", requestData.SpecificFilter.ChatID)
	}
	switch authData.Role {
	case model.UserRoleAdmin:
		// Admin can view all messages
	case model.UserRoleAgent:
		// Agent can only view messages if they are assigned to the chat
		if chat.AgentID != nil && *chat.AgentID != authData.UserID {
			u.zsLog.Warnf("[GetMessagesByChatID] Agent user ID %s is not assigned to chat ID %s", authData.UserID, chat.DocumentID)
			return filter_request.FilterResponse[dto.MessageResponse]{}, false, errs.ErrGenericForbidden
		}
	case model.UserRoleSupervisor:
		if chat.AgentID == nil {
			u.zsLog.Warnf("[GetMessagesByChatID] Chat ID %s is not assigned to any agent, supervisor user ID %s cannot view the messages", chat.DocumentID, authData.UserID)
			return filter_request.FilterResponse[dto.MessageResponse]{}, false, errs.ErrGenericForbidden
		}
		// Supervisor can view all messages in the chats that they supervise
		agents, err := u.userRepository.GetBySupervisorID(ctx, authData.UserID)
		if err != nil {
			u.zsLog.Errorf("[GetMessagesByChatID] Failed to get agents by supervisor ID: %v", err)
			return filter_request.FilterResponse[dto.MessageResponse]{}, true, err
		}
		agentIDs := make(map[string]bool)
		for _, agent := range agents {
			agentIDs[agent.DocumentID] = true
		}
		if !agentIDs[*chat.AgentID] {
			u.zsLog.Warnf("[GetMessagesByChatID] Supervisor user ID %s does not supervise agent assigned to chat ID %s", authData.UserID, chat.DocumentID)
			return filter_request.FilterResponse[dto.MessageResponse]{}, false, errs.ErrGenericForbidden
		}
	default:
		u.zsLog.Warnf("[GetMessagesByChatID] Unauthorized role %s for getting messages by chat ID", authData.Role)
		return filter_request.FilterResponse[dto.MessageResponse]{}, false, fmt.Errorf("unauthorized role %s for getting messages by chat ID", authData.Role)
	}
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
			storageMediaMap, err = u.storageMediaRepository.GetByIDs(ctx, storageMediaIds)
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
	if inputData.ID != nil && inputData.ChatID != nil {
		message := model.Message{
			DocumentID:      *inputData.ID,
			Wamid:           inputData.Wamid,
			ChatID:          *inputData.ChatID,
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
		_, err := u.messageRepository.Upsert(ctx, nil, message)
		if err != nil {
			u.zsLog.Errorf("[SaveMessage] Failed to upsert existing message: %v", err)
			return true, err
		}
		_, err = u.searchMessageRepository.AddDocuments(ctx, []model.Message{message})
		if err != nil {
			u.zsLog.Errorf("[SaveMessage] Failed to add existing message to search index: %v", err)
		}
		return false, err
	}
	// TODO: check if chat belongs to tenant
	// check tenant chat type
	tenant, err := u.tenantRepository.GetByID(ctx, tenantID)
	if err != nil {
		u.zsLog.Errorf("[SaveMessage] Failed to get tenant: %v", err)
		return true, err
	}
	var chatID string
	var chat model.Chat

	if inputData.ChatID != nil {
		chat, err = u.chatRepository.GetByID(ctx, *inputData.ChatID)
		if err != nil {
			if err == errs.ErrGenericNotFound {
				u.zsLog.Warnf("[SaveMessage] No chat found for ID %s, will create new chat if tenant chat type is not ticket", *inputData.ChatID)
				return false, err
			}
			u.zsLog.Errorf("[SaveMessage] Failed to get chat by ID: %v", err)
			return true, err
		}
		chatID = chat.DocumentID
	} else {
		switch tenant.ChatType {
		case "ticket":
			chat, err = u.chatRepository.GetRunningTicketChat(ctx, inputData.PhoneNumberId, inputData.RecipientId)
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
	}
	if chat.DocumentID == "" {
		// create new chat if not exist
		chat = model.Chat{
			DocumentID:    chatID,
			TenantID:      tenantID,
			PhoneNumberId: inputData.PhoneNumberId,
			RecipientId:   inputData.RecipientId,
			RecipientName: inputData.RecipientName,
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
	if chat.RecipientName == "" {
		chat.RecipientName = inputData.RecipientId
	}
	if inputData.UserLastMessageAt != nil {
		chat.UserLastMessageAt = inputData.UserLastMessageAt
	}
	chat.LastMessage = inputData.LastMessage
	chat.UpdatedAt = time.Now()

	messageLogID, err := uuid.NewV7()
	if err != nil {
		u.zsLog.Errorf("[SaveMessage] Failed to generate message log ID: %v", err)
		return true, err
	}
	// create system log message for new chat if chat is new, or chat type is ticket but no opened ticket chat found
	// if chat is not new and chat type is not ticket, it means the chat is already created from previous message, so no need to create system log message
	logData := model.MessageSystemData{}
	if tenant.ChatType == "ticket" {
		logData.Type = "ticket_created"
		logData.Message = fmt.Sprintf("Ticket created with ID %s", chat.DocumentID)
	} else {
		logData.Type = "chat_created"
		logData.Message = "Chat created"
	}
	logPayload, err := utils.AnyToJsonString(logData)
	if err != nil {
		u.zsLog.Errorf("[SaveMessage] Failed to convert log payload to JSON: %v", err)
		return true, err
	}
	messageLog := model.Message{
		DocumentID:      messageLogID.String(),
		Wamid:           "",
		ChatID:          chatID,
		MessageType:     "",
		MessageCategory: "system_flag",
		SenderName:      "",
		Payload:         logPayload,
		StorageMediaID:  nil,
		StorageMedia:    nil,
		Status:          "",
		CreatedAt:       inputData.CreatedAt.Add(-1 * time.Second), // set created time before the first message to make sure the log message is always before the first message in the chat
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
	} else {
		newMessageID, genErr := uuid.NewV7()
		if genErr != nil {
			u.zsLog.Errorf("[SaveMessage] Failed to generate message ID: %v", genErr)
			return true, genErr
		}
		message.DocumentID = newMessageID.String()
		// if message ID is not provided, it means the message is created from incoming webhook, so we need to set created time to current time if it's not provided to make sure the message is sorted correctly in the chat
		if message.CreatedAt.IsZero() {
			message.CreatedAt = time.Now()
		}
	}
	var messagesToIndex []model.Message
	serverError, err := u.txManager.DoFirestore(ctx, func(ctx context.Context, txFirestore *firestore.Transaction) (bool, error) {
		_, isNewChat, err := u.chatRepository.Upsert(ctx, txFirestore, chat)
		if err != nil {
			u.zsLog.Errorf("[SaveMessage] Failed to upsert chat: %v", err)
			return true, err
		}
		_, err = u.messageRepository.Upsert(ctx, txFirestore, message)
		if err != nil {
			u.zsLog.Errorf("[SaveMessage] Failed to save message: %v", err)
			return true, err
		}
		if isNewChat {
			messageLog, err = u.messageRepository.Upsert(ctx, txFirestore, messageLog)
			if err != nil {
				u.zsLog.Errorf("[SaveMessage] Failed to upsert message: %v", err)
				return true, err
			}
		}
		// TODO: check if adding system log message to search index is necessary or not
		if isNewChat {
			messagesToIndex = append(messagesToIndex, messageLog)
		}
		messagesToIndex = append(messagesToIndex, message)
		return false, nil
	})
	_, err = u.searchMessageRepository.AddDocuments(ctx, messagesToIndex)
	if err != nil {
		u.zsLog.Errorf("[SaveMessage] Failed to add message to search index: %v", err)
	}
	return serverError, err
}

func (u *MessageUsecase) GetByWamid(ctx context.Context, tenantID string, phoneNumberId string, recipientId string, wamid string) (model.Message, bool, error) {
	var message model.Message
	tenant, err := u.tenantRepository.GetByID(ctx, tenantID)
	if err != nil {
		u.zsLog.Errorf("[GetByWamid] Failed to get tenant by phone number ID: %v", err)
		return model.Message{}, true, err
	}
	var chatID string
	if tenant.ChatType == "ticket" {
		chat, err := u.chatRepository.GetRunningTicketChat(ctx, phoneNumberId, recipientId)
		if err != nil && err != errs.ErrGenericNotFound {
			u.zsLog.Errorf("[GetByWamid] Failed to get running ticket chat: %v", err)
			return model.Message{}, true, err
		}
		if err == nil {
			chatID = chat.DocumentID
		}
	}
	defaultChatID := fmt.Sprintf("%s-%s", recipientId, phoneNumberId)
	if chatID == "" {
		chatID = defaultChatID
	}
	var getErr error
	// if tenant uses ticketing, it will use the recent open ticket association first to find the message,
	// if not, it will search in the default chat associated with the phone number
	for range 2 {
		message, getErr = u.messageRepository.GetByWamid(ctx, chatID, wamid)
		if getErr == nil {
			break
		}
		if getErr != errs.ErrGenericNotFound {
			return model.Message{}, true, getErr
		}
		if chatID == defaultChatID {
			return model.Message{}, true, errs.ErrGenericNotFound
		}
		chatID = defaultChatID
	}
	return message, false, nil
}

func (u *MessageUsecase) CheckCanSendMessage(ctx context.Context, authData dto.AuthData, chatID string) (bool, bool, error) {
	chat, err := u.chatRepository.GetByID(ctx, chatID)
	if err != nil {
		u.zsLog.Errorf("[CheckCanSendMessage] Failed to get chat by ID: %v", err)
		return false, true, err
	}
	if chat.TenantID != authData.TenantID {
		u.zsLog.Warnf("[CheckCanSendMessage] WhatsApp business tenant ID %s does not match auth tenant ID %s", chat.TenantID, authData.TenantID)
		return false, false, fmt.Errorf("no chat found for ID %s", chatID)
	}
	switch authData.Role {
	case model.UserRoleAdmin:
		// Admin can send message to all chats
		return true, false, nil
	case model.UserRoleAgent:
		// Agent can only send message if they are assigned to the chat
		if chat.AgentID != nil && *chat.AgentID == authData.UserID {
			return true, false, nil
		}
		u.zsLog.Warnf("[CheckCanSendMessage] Agent user ID %s is not assigned to chat ID %s", authData.UserID, chat.DocumentID)
		return false, false, errs.ErrGenericForbidden
	case model.UserRoleSupervisor:
		if chat.AgentID == nil {
			u.zsLog.Warnf("[CheckCanSendMessage] Chat ID %s is not assigned to any agent, supervisor user ID %s cannot send message", chat.DocumentID, authData.UserID)
			return false, false, errs.ErrGenericForbidden
		}
		// Supervisor can send message in the chats that they supervise
		agents, err := u.userRepository.GetBySupervisorID(ctx, authData.UserID)
		if err != nil {
			u.zsLog.Errorf("[CheckCanSendMessage] Failed to get agents by supervisor ID: %v", err)
			return false, true, err
		}
		for _, agent := range agents {
			if agent.DocumentID == *chat.AgentID {
				return true, false, nil
			}
		}
		u.zsLog.Warnf("[CheckCanSendMessage] Supervisor user ID %s does not supervise agent assigned to chat ID %s", authData.UserID, chat.DocumentID)
		return false, false, errs.ErrGenericForbidden
	default:
		u.zsLog.Warnf("[CheckCanSendMessage] Unauthorized role %s for checking can send message", authData.Role)
		return false, false, fmt.Errorf("unauthorized role %s for checking can send message", authData.Role)
	}
}
