package chat_usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/meta/whatsapp_business"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

// SendMessage validates and sends an outbound WhatsApp message, then persists it.
//
// Flow summary:
// - Resolves recipient context from ChatID or TicketID.
// - Creates a WhatsApp client when one is not injected.
// - Validates message component and resolves media references when present.
// - Sends the message through WhatsApp API.
// - Persists the sent message through SaveMessage.
//
// Returns:
// - serverError=true when the failure is considered internal/server-side.
// - error describing the failure.
func (u *ChatUsecase) SendMessage(ctx context.Context, whatsappClient *whatsapp_business.Client, tenantID string, inputData dto.MessageSendRequest) (bool, error) {
	var err error
	var phoneNumberId string
	var recipientId string
	var recipientName string
	if inputData.ChatID != nil && *inputData.ChatID != "" {
		chat, err := u.chatRepository.GetByID(ctx, *inputData.ChatID)
		if err != nil {
			u.zsLog.Errorf("[SendMessage] Failed to get chat by ID: %v", err)
			return err != errs.ErrGenericNotFound, err
		}
		phoneNumberId = chat.PhoneNumberId
		recipientId = chat.RecipientId
		recipientName = chat.RecipientName
	} else if inputData.TicketID != nil && *inputData.TicketID != "" {
		ticket, err := u.ticketRepository.GetByID(ctx, *inputData.TicketID)
		if err != nil {
			u.zsLog.Errorf("[SendMessage] Failed to get ticket by ID: %v", err)
			return err != errs.ErrGenericNotFound, err
		}
		phoneNumberId = ticket.PhoneNumberId
		recipientId = ticket.RecipientId
		recipientName = ticket.RecipientName
	}

	if whatsappClient == nil {
		whatsappClient, _, err = u.waBusinessAccountUsecase.GetWhatsappClient(ctx, tenantID, phoneNumberId)
		if err != nil {
			u.zsLog.Errorf("[SendMessage] Failed to get WhatsApp client: %v", err)
			return true, err
		}
	}
	component, err := whatsapp_business.NewComponent(inputData.Type, inputData.Payload)
	if err != nil {
		u.zsLog.Errorf("[SendMessage] Failed to validate message component: %v", err)
		return false, err
	}
	var sto *model.StorageMedia
	// If media is present, try to resolve it as an internal signed link first.
	// Otherwise, validate external URL and persist metadata for later retrieval.
	if media := whatsapp_business.GetMedia(component); media != nil {
		if media.Link != nil {
			mediaToken, err := u.storageMediaUsecase.ParsePublicURL(*media.Link)
			if err == nil {
				// get file URL from signed URL
				storageMediaID, serverError, err := u.storageMediaUsecase.ParseMediaToken(mediaToken)
				if err != nil {
					u.zsLog.Errorf("[SendMessage] Failed to parse media token: %v", err)
					return serverError, err
				}
				storageMedia, err := u.storageMediaRepository.GetByID(ctx, storageMediaID)
				if err != nil {
					u.zsLog.Errorf("[SendMessage] Failed to get storage media by document ID: %v", err)
					return true, err
				}
				sto = &storageMedia
			} else {
				// check if link is accessible
				resp, err := http.Head(*media.Link)
				if err != nil || resp.StatusCode != http.StatusOK {
					u.zsLog.Errorf("[SendMessage] Media link is not accessible: %v", err)
					return false, fmt.Errorf("media link is not accessible")
				}
				urlHeaders := resp.Header
				mimeType := urlHeaders.Get("Content-Type")
				extension := whatsapp_business.ParseMediaExtension(mimeType)
				if extension == "" {
					u.zsLog.Errorf("[SendMessage] Unsupported media type: %v", mimeType)
					return true, fmt.Errorf("unsupported media type: %s", mimeType)
				}
				// TODO: check file size is allowed or not
				originalFileName := utils.GetFileNameFromURL(urlHeaders, *media.Link)
				newStorageMediaID, err := uuid.NewV7()
				if err != nil {
					u.zsLog.Errorf("[SendMessage] Failed to generate storage media ID: %v", err)
					return true, err
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
					return true, err
				}
				sto = &storageMedia
			}
		}
	}
	sendResponse, httpCode, err := whatsappClient.SendMessage(recipientId, "individual", component)
	if err != nil {
		u.zsLog.Errorf("[SendMessage] Failed to send message: %v", err)
		return httpCode >= http.StatusInternalServerError, err
	}
	payloadData, err := utils.AnyToJsonString(component.GetPayload())
	if err != nil {
		u.zsLog.Errorf("[SendMessage] Failed to convert payload to JSON")
	}
	var storageMediaID *string
	if sto != nil {
		storageMediaID = &sto.DocumentID
	}

	newID, err := uuid.NewV7()
	if err != nil {
		u.zsLog.Errorf("[SendMessage] Failed to generate message ID: %v", err)
		return true, err
	}
	newIDStr := newID.String()
	req := dto.MessageSaveRequest{
		ID:              &newIDStr,
		ChatID:          inputData.ChatID,
		TicketID:        inputData.TicketID,
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
	}
	serverError, err := u.SaveMessage(ctx, tenantID, req)
	if err != nil {
		u.zsLog.Errorf("[SendMessage] Failed to save message: %v", err)
		return serverError, err
	}

	return false, nil
}

// GetMessagesByChatID returns paginated chat messages with role-based access control.
//
// Access rules:
// - Admin: can view all messages.
// - Agent: can view messages only when assigned to the chat (or chat is unassigned).
// - Supervisor: can view only chats assigned to agents under their supervision.
//
// Storage media references are expanded with generated public URLs when available.
func (u *ChatUsecase) GetMessagesByChatID(ctx context.Context, authData dto.AuthData, requestData filter_request.FilterRequest[dto.MessageGetByChatIDRequest]) (filter_request.FilterResponse[dto.MessageResponse], bool, error) {
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

// SaveMessage persists an inbound/outbound message to chat storage.
//
// Behavior summary:
// - If inputData.ID is provided, treats the operation as an upsert path.
// - Delegates to ticket flow when TicketID is provided or tenant chat type is "ticket".
// - Creates chat and a one-time "chat_created" system message for first-message scenarios.
// - Persists records in Firestore transaction, then updates search index after commit.
func (u *ChatUsecase) SaveMessage(ctx context.Context, tenantID string, inputData dto.MessageSaveRequest) (bool, error) {
	if inputData.ID != nil {
		if inputData.ChatID != nil {
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
			// Update the last message of chat and insert new message
			serverError, err := u.txManager.DoFirestore(ctx, func(ctx context.Context, txFirestore *firestore.Transaction) (bool, error) {
				var err error
				if inputData.LastMessage != "" {
					err = u.chatRepository.UpdateLastMessage(ctx, txFirestore, *inputData.ChatID, inputData.LastMessage)
					if err != nil {
						u.zsLog.Errorf("[SaveMessage] Failed to update chat last message: %v", err)
						return true, err
					}
				}
				_, err = u.messageRepository.Upsert(ctx, nil, message)
				if err != nil {
					u.zsLog.Errorf("[SaveMessage] Failed to upsert existing message: %v", err)
					return true, err
				}
				_, err = u.searchMessageRepository.AddDocuments(ctx, []model.Message{message})
				if err != nil {
					u.zsLog.Errorf("[SaveMessage] Failed to add existing message to search index: %v", err)
				}
				return false, nil
			})
			return serverError, err
		} else if inputData.TicketID != nil {
			return u.ticketUsecase.SaveTicketMessage(ctx, tenantID, inputData)
		}
	}
	// TODO: check if chat belongs to tenant
	// check tenant chat type
	tenant, err := u.tenantRepository.GetByID(ctx, tenantID)
	if err != nil {
		u.zsLog.Errorf("[SaveMessage] Failed to get tenant: %v", err)
		return true, err
	}
	// check if message from a broadcast, if yes, don't save message to ticket, because broadcast message is not related to any ticket.
	payloadMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(inputData.Payload), &payloadMap)
	if err != nil {
		u.zsLog.Errorf("[SaveMessage] Failed to unmarshal payload: %v", err)
		return true, err
	}
	isBroadcast := false
	if _, ok := payloadMap["button"]; ok {
		buttonMap := payloadMap["button"].(map[string]interface{})
		if payload, ok := buttonMap["payload"]; ok {
			payloadStr := payload.(string)
			if strings.HasPrefix(payloadStr, "broadcast_") {
				isBroadcast = true
			}
		}
	}

	if tenant.ChatType == "ticket" && !isBroadcast {
		return u.ticketUsecase.SaveTicketMessage(ctx, tenantID, inputData)
	}
	var chat model.Chat
	chatID := fmt.Sprintf("%s-%s", inputData.RecipientId, inputData.PhoneNumberId)
	chat, err = u.chatRepository.GetByID(ctx, chatID)
	if err != nil && err != errs.ErrGenericNotFound {
		u.zsLog.Errorf("[SaveMessage] Failed to get chat by ID: %v", err)
		return true, err
	}
	if chat.DocumentID == "" {
		// create new chat if not exist
		chat = model.Chat{
			DocumentID:    chatID,
			TenantID:      tenantID,
			PhoneNumberId: inputData.PhoneNumberId,
			RecipientId:   inputData.RecipientId,
			RecipientName: inputData.RecipientName,
			CreatedAt:     inputData.CreatedAt,
			UpdatedAt:     inputData.CreatedAt,
		}
	}
	chat.ChatType = "individual"
	if chat.RecipientName == "" {
		chat.RecipientName = inputData.RecipientId
	}
	if inputData.UserLastMessageAt != nil {
		chat.UserLastMessageAt = inputData.UserLastMessageAt
	}
	if inputData.LastMessage != "" {
		chat.LastMessage = inputData.LastMessage
	}

	messageLogID, err := uuid.NewV7()
	if err != nil {
		u.zsLog.Errorf("[SaveMessage] Failed to generate message log ID: %v", err)
		return true, err
	}
	logData := model.MessageSystemData{
		Type:    "chat_created",
		Message: "Chat created",
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
		if message.CreatedAt.IsZero() {
			message.CreatedAt = time.Now()
		}
	}
	var messagesToIndex []model.Message
	// Keep chat/message persistence atomic.
	// Search indexing is performed after transaction commit as best-effort.
	serverError, err := u.txManager.DoFirestore(ctx, func(ctx context.Context, txFirestore *firestore.Transaction) (bool, error) {
		_, isNewChat, err := u.chatRepository.Upsert(ctx, txFirestore, chat)
		if err != nil {
			u.zsLog.Errorf("[SaveMessage] Failed to upsert chat: %v", err)
			return true, err
		}
		if isNewChat {
			messageLog, err = u.messageRepository.Upsert(ctx, txFirestore, messageLog)
			if err != nil {
				u.zsLog.Errorf("[SaveMessage] Failed to upsert message: %v", err)
				return true, err
			}
			// TODO: check if adding system log message to search index is necessary or not
			messagesToIndex = append(messagesToIndex, messageLog)
		}
		_, err = u.messageRepository.Upsert(ctx, txFirestore, message)
		if err != nil {
			u.zsLog.Errorf("[SaveMessage] Failed to save message: %v", err)
			return true, err
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

// GetMessageByWamid finds a message by WhatsApp message ID within a recipient chat scope.
//
// It derives chat ID from recipientId and phoneNumberId, validates tenant ownership,
// then queries the message by WAMID.
func (u *ChatUsecase) GetMessageByWamid(ctx context.Context, tenantID string, phoneNumberId string, recipientId string, wamid string) (model.Message, bool, error) {
	chatID := fmt.Sprintf("%s-%s", recipientId, phoneNumberId)
	chat, err := u.chatRepository.GetByID(ctx, chatID)
	if err != nil {
		u.zsLog.Errorf("[GetMessageByWamid] Failed to get chat by ID: %v", err)
		return model.Message{}, err != errs.ErrGenericNotFound, err
	}
	if chat.TenantID != tenantID {
		u.zsLog.Warnf("[GetMessageByWamid] WhatsApp business tenant ID %s does not match auth tenant ID %s", chat.TenantID, tenantID)
		return model.Message{}, false, errs.ErrGenericForbidden
	}
	message, err := u.messageRepository.GetMessageByWamid(ctx, chatID, wamid)
	if err != nil {
		u.zsLog.Errorf("[GetMessageByWamid] Failed to get message by wamid: %v", err)
		return model.Message{}, err != errs.ErrGenericNotFound, err
	}
	return message, false, nil
}

// CheckCanSendMessage validates whether an authenticated user can send a message.
//
// Context can be provided using either chatID or ticketID. Authorization is role-based:
// - Admin: always allowed.
// - Agent: allowed only when assigned to the target chat/ticket.
// - Supervisor: allowed only when assigned agent is under their supervision.
func (u *ChatUsecase) CheckCanSendMessage(ctx context.Context, authData dto.AuthData, chatID *string, ticketID *string) (bool, bool, error) {
	var assignedAgentID *string
	if chatID != nil && *chatID != "" {
		chat, err := u.chatRepository.GetByID(ctx, *chatID)
		if err != nil {
			u.zsLog.Errorf("[CheckCanSendMessage] Failed to get chat by ID: %v", err)
			return false, true, err
		}
		if chat.TenantID != authData.TenantID {
			u.zsLog.Warnf("[CheckCanSendMessage] WhatsApp business tenant ID %s does not match auth tenant ID %s", chat.TenantID, authData.TenantID)
			return false, false, errs.ErrGenericForbidden
		}
		assignedAgentID = chat.AgentID
	} else if ticketID != nil && *ticketID != "" {
		ticket, err := u.ticketRepository.GetByID(ctx, *ticketID)
		if err != nil {
			u.zsLog.Errorf("[CheckCanSendMessage] Failed to get ticket by ID: %v", err)
			return false, err != errs.ErrGenericNotFound, err
		}
		if ticket.TenantID != authData.TenantID {
			u.zsLog.Warnf("[CheckCanSendMessage] Ticket tenant ID %s does not match auth tenant ID %s", ticket.TenantID, authData.TenantID)
			return false, false, errs.ErrGenericForbidden
		}
		assignedAgentID = ticket.AgentID
	} else {
		u.zsLog.Warnf("[CheckCanSendMessage] Both chat ID and ticket ID are empty, cannot check if user can send message")
		return false, false, fmt.Errorf("both chat ID and ticket ID are empty, cannot check if user can send message")
	}
	switch authData.Role {
	case model.UserRoleAdmin:
		// Admin can send message to all chats
		return true, false, nil
	case model.UserRoleAgent:
		// Agent can only send message if they are assigned to the chat
		if assignedAgentID != nil && *assignedAgentID == authData.UserID {
			return true, false, nil
		}
		u.zsLog.Warnf("[CheckCanSendMessage] Agent user ID %s is not assigned to chat ID %s", authData.UserID, chatID)
		return false, false, errs.ErrGenericForbidden
	case model.UserRoleSupervisor:
		if assignedAgentID == nil {
			u.zsLog.Warnf("[CheckCanSendMessage] Chat ID %s is not assigned to any agent, supervisor user ID %s cannot send message to the chat", chatID, authData.UserID)
			return false, false, errs.ErrGenericForbidden
		}
		// Supervisor can send message in the chats that they supervise
		agents, err := u.userRepository.GetBySupervisorID(ctx, authData.UserID)
		if err != nil {
			u.zsLog.Errorf("[CheckCanSendMessage] Failed to get agents by supervisor ID: %v", err)
			return false, true, err
		}
		for _, agent := range agents {
			if agent.DocumentID == *assignedAgentID {
				return true, false, nil
			}
		}
		u.zsLog.Warnf("[CheckCanSendMessage] Supervisor user ID %s does not supervise agent assigned to chat ID %s", authData.UserID, chatID)
		return false, false, errs.ErrGenericForbidden
	default:
		u.zsLog.Warnf("[CheckCanSendMessage] Unauthorized role %s for checking can send message", authData.Role)
		return false, false, fmt.Errorf("unauthorized role %s for checking can send message", authData.Role)
	}

}
