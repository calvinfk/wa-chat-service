package chat_usecase

import (
	"context"
	"fmt"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/meilisearch/meilisearch-go"
	"go.uber.org/zap"
)

type ChatUsecase struct {
	chatRepository              repository.Chat
	waPhoneRepository           repository.WaPhone
	waBusinessAccountRepository repository.WaBusinessAccount
	userRepository              repository.User
	messageRepository           repository.Message
	tenantRepository            repository.Tenant
	searchMessageRepository     repository.SearchMessage
	txManager                   *utils.TxManager
	zsLog                       *zap.SugaredLogger
}

func NewChatUsecase(chatRepository repository.Chat, waPhoneRepository repository.WaPhone, waBusinessAccountRepository repository.WaBusinessAccount, userRepository repository.User, messageRepository repository.Message, tenantRepository repository.Tenant, searchMessageRepository repository.SearchMessage, txManager *utils.TxManager, zsLog *zap.SugaredLogger) *ChatUsecase {
	return &ChatUsecase{
		chatRepository:              chatRepository,
		waPhoneRepository:           waPhoneRepository,
		waBusinessAccountRepository: waBusinessAccountRepository,
		userRepository:              userRepository,
		messageRepository:           messageRepository,
		tenantRepository:            tenantRepository,
		searchMessageRepository:     searchMessageRepository,
		txManager:                   txManager,
		zsLog:                       zsLog,
	}
}

func (uc *ChatUsecase) GetChatByPhoneNumberId(ctx context.Context, tenantID string, requestData filter_request.FilterRequest[dto.ChatGetByPhoneNumberIdRequest]) (filter_request.FilterResponse[dto.ChatGetByPhoneNumberIdResponse], bool, error) {
	// TODO: check if phone number belongs to tenant
	response, err := uc.chatRepository.GetChatByPhoneNumberId(ctx, requestData)
	if err != nil {
		uc.zsLog.Errorf("[GetChatByPhoneNumberId] error while getting chat by phone number id: %v", err)
		return filter_request.FilterResponse[dto.ChatGetByPhoneNumberIdResponse]{}, true, err
	}
	return response, false, nil
}

func (uc *ChatUsecase) CloseTicket(ctx context.Context, tenantID string, requestData dto.ChatCloseTicketRequest) (bool, error) {
	chat, err := uc.chatRepository.GetByID(ctx, requestData.ChatID)
	if err != nil {
		uc.zsLog.Errorf("[CloseTicket] error while getting chat by id: %v", err)
		if err == errs.ErrGenericNotFound {
			return false, err
		}
		return true, err
	}
	phone, err := uc.waPhoneRepository.GetByPhoneNumberId(ctx, chat.PhoneNumberId)
	if err != nil {
		uc.zsLog.Errorf("[CloseTicket] error while getting phone by id: %v", err)
		return true, err
	}
	waba, err := uc.waBusinessAccountRepository.GetByID(ctx, phone.WaBusinessAccountID)
	if err != nil {
		uc.zsLog.Errorf("[CloseTicket] error while getting wa business account by id: %v", err)
		return true, err
	}
	if waba.TenantID != tenantID {
		uc.zsLog.Errorf("[CloseTicket] tenant id mismatch: %s vs %s", waba.TenantID, tenantID)
		return false, errs.ErrGenericForbidden
	}

	if chat.ChatStatus == model.ChatStatusClosed {
		return false, nil
	}
	chat.ChatStatus = model.ChatStatusClosed
	chat.UpdatedAt = time.Now()

	logPayload, err := utils.AnyToJsonString(model.MessageSystemData{
		Type:    "close_ticket",
		Message: "Ticket closed",
	})
	if err != nil {
		uc.zsLog.Errorf("[CloseTicket] error while marshalling log payload: %v", err)
		return true, err
	}
	messageID, err := uuid.NewV7()
	if err != nil {
		uc.zsLog.Errorf("[CloseTicket] error while generating message id: %v", err)
		return true, err
	}
	message := model.Message{
		DocumentID:  messageID.String(),
		ChatID:      chat.DocumentID,
		MessageType: "system_flag",
		Payload:     logPayload,
	}
	serverError, err := uc.txManager.DoFirestore(ctx, func(ctx context.Context, txFirestore *firestore.Transaction) (bool, error) {
		_, _, err = uc.chatRepository.Upsert(ctx, txFirestore, chat)
		if err != nil {
			uc.zsLog.Errorf("[CloseTicket] error while upserting chat: %v", err)
			return true, err
		}
		_, err = uc.messageRepository.Upsert(ctx, txFirestore, message)
		if err != nil {
			uc.zsLog.Errorf("[CloseTicket] error while creating system message: %v", err)
			return true, err
		}
		taskInfo, err := uc.searchMessageRepository.AddDocuments(ctx, []model.Message{message})
		if err != nil {
			uc.zsLog.Errorf("[CloseTicket] error while adding message to search index: %v", err)
			return true, err
		}
		for {
			time.Sleep(100 * time.Millisecond)
			task, err := uc.searchMessageRepository.GetTaskStatus(ctx, taskInfo.TaskUID)
			if err != nil {
				uc.zsLog.Errorf("[CloseTicket] error while getting search index task status: %v", err)
				return true, err
			}
			if task.Status != meilisearch.TaskStatusProcessing {
				if task.Status == meilisearch.TaskStatusFailed {
					uc.zsLog.Errorf("[CloseTicket] search index task failed: %v", task.Error)
					return true, fmt.Errorf("search index task failed: %v", task.Error)
				}
				break
			}
		}
		return false, nil
	})
	if err != nil {
		return serverError, err
	}
	return false, nil
}

func (uc *ChatUsecase) AssignAgent(ctx context.Context, tenantID string, requestData dto.ChatAssignAgentRequest) (bool, error) {
	chat, err := uc.chatRepository.GetByID(ctx, requestData.ChatID)
	if err != nil {
		uc.zsLog.Errorf("[AssignAgent] error while getting chat by id: %v", err)
		if err == errs.ErrGenericNotFound {
			return false, err
		}
		return true, err
	}
	if chat.ChatStatus == model.ChatStatusClosed {
		return false, fmt.Errorf("cannot assign agent to closed chat")
	}
	if chat.AgentID != nil && *chat.AgentID == requestData.AgentID {
		return false, nil
	}
	phone, err := uc.waPhoneRepository.GetByPhoneNumberId(ctx, chat.PhoneNumberId)
	if err != nil {
		uc.zsLog.Errorf("[AssignAgent] error while getting phone by id: %v", err)
		return true, err
	}
	waba, err := uc.waBusinessAccountRepository.GetByID(ctx, phone.WaBusinessAccountID)
	if err != nil {
		uc.zsLog.Errorf("[AssignAgent] error while getting wa business account by id: %v", err)
		return true, err
	}
	if waba.TenantID != tenantID {
		uc.zsLog.Errorf("[AssignAgent] tenant id mismatch: %s vs %s", waba.TenantID, tenantID)
		return false, errs.ErrGenericForbidden
	}
	agent, err := uc.userRepository.GetByID(ctx, requestData.AgentID)
	if err != nil {
		uc.zsLog.Errorf("[AssignAgent] error while getting agent by id: %v", err)
		return true, err
	}
	if agent.TenantID != tenantID {
		uc.zsLog.Errorf("[AssignAgent] tenant id mismatch: %s vs %s", agent.TenantID, tenantID)
		return false, fmt.Errorf("agent does not belong to tenant")
	}
	if agent.Role != "agent" {
		uc.zsLog.Errorf("[AssignAgent] user role mismatch: %s is not an agent", agent.Role)
		return false, fmt.Errorf("user is not an agent")
	}

	var prevAgentName string

	if chat.AgentID != nil {
		prevAgent, err := uc.userRepository.GetByID(ctx, *chat.AgentID)
		if err != nil {
			uc.zsLog.Errorf("[AssignAgent] error while getting previous agent by id: %v", err)
			return true, err
		}
		prevAgentName = prevAgent.Name
	}

	chat.AgentID = &requestData.AgentID
	chat.UpdatedAt = time.Now()

	messageID, err := uuid.NewV7()
	if err != nil {
		uc.zsLog.Errorf("[AssignAgent] error while generating message id: %v", err)
		return true, err
	}
	var logPayload string
	if prevAgentName != "" {
		logPayload, err = utils.AnyToJsonString(model.MessageSystemData{
			Type:    "move_agent",
			Message: fmt.Sprintf("Agent changed from %s to %s", prevAgentName, agent.Name),
		})
	} else {
		logPayload, err = utils.AnyToJsonString(model.MessageSystemData{
			Type:    "move_agent",
			Message: fmt.Sprintf("Agent %s assigned to chat", agent.Name),
		})
	}
	if err != nil {
		uc.zsLog.Errorf("[AssignAgent] error while marshalling log payload: %v", err)
		return true, err
	}
	message := model.Message{
		DocumentID:  messageID.String(),
		ChatID:      chat.DocumentID,
		MessageType: "system_flag",
		Payload:     logPayload,
	}
	serverError, err := uc.txManager.DoFirestore(ctx, func(ctx context.Context, txFirestore *firestore.Transaction) (bool, error) {
		_, _, err = uc.chatRepository.Upsert(ctx, txFirestore, chat)
		if err != nil {
			uc.zsLog.Errorf("[AssignAgent] error while upserting chat: %v", err)
			return true, err
		}
		_, err = uc.messageRepository.Upsert(ctx, txFirestore, message)
		if err != nil {
			uc.zsLog.Errorf("[AssignAgent] error while creating system message: %v", err)
			return true, err
		}
		task, err := uc.searchMessageRepository.AddDocuments(ctx, []model.Message{message})
		if err != nil {
			uc.zsLog.Errorf("[CloseTicket] error while adding message to search index: %v", err)
			return true, err
		}
		for {
			time.Sleep(100 * time.Millisecond)
			task, err := uc.searchMessageRepository.GetTaskStatus(ctx, task.TaskUID)
			if err != nil {
				uc.zsLog.Errorf("[AssignAgent] error while getting search index task status: %v", err)
				return true, err
			}
			if task.Status != meilisearch.TaskStatusProcessing {
				if task.Status == meilisearch.TaskStatusFailed {
					uc.zsLog.Errorf("[AssignAgent] search index task failed: %v", task.Error)
					return true, fmt.Errorf("search index task failed: %v", task.Error)
				}
				break
			}
		}
		return false, nil
	})
	if err != nil {
		return serverError, err
	}
	return false, nil
}

func (uc *ChatUsecase) GetByID(ctx context.Context, chatID string) (model.Chat, bool, error) {
	chat, err := uc.chatRepository.GetByID(ctx, chatID)
	if err != nil {
		uc.zsLog.Errorf("[GetByID] error while getting chat by id: %v", err)
		if err == errs.ErrGenericNotFound {
			return model.Chat{}, false, err
		}
		return model.Chat{}, true, err
	}
	return chat, false, nil
}

func (uc *ChatUsecase) CreateChat(ctx context.Context, tenantID string, requestData dto.ChatCreateRequest) (model.Chat, bool, error) {
	phone, err := uc.waPhoneRepository.GetByPhoneNumberId(ctx, requestData.PhoneNumberId)
	if err != nil {
		uc.zsLog.Errorf("[CreateChat] error while getting phone by id: %v", err)
		return model.Chat{}, true, err
	}
	waba, err := uc.waBusinessAccountRepository.GetByID(ctx, phone.WaBusinessAccountID)
	if err != nil {
		uc.zsLog.Errorf("[CreateChat] error while getting wa business account by id: %v", err)
		return model.Chat{}, true, err
	}
	if waba.TenantID != tenantID {
		uc.zsLog.Errorf("[CreateChat] tenant id mismatch: %s vs %s", waba.TenantID, tenantID)
		return model.Chat{}, false, errs.ErrGenericForbidden
	}
	tenant, err := uc.tenantRepository.GetByID(ctx, tenantID)
	if err != nil {
		uc.zsLog.Errorf("[CreateChat] error while getting tenant by id: %v", err)
		return model.Chat{}, true, err
	}
	if tenant.ChatType == "ticket" {
		// check if there's an open ticket for the same phone number and recipient
		chat, err := uc.chatRepository.GetOpenedTicketChatByPhoneNumberId(ctx, requestData.PhoneNumberId, requestData.RecipientId)
		if err != nil && err != errs.ErrGenericNotFound {
			uc.zsLog.Errorf("[CreateChat] error while getting open chats by phone number id: %v", err)
			return model.Chat{}, true, err
		} else if err == nil {
			return chat, false, nil
		}
	} else {
		chatID := fmt.Sprintf("%s-%s", requestData.PhoneNumberId, requestData.RecipientId)
		chat, err := uc.chatRepository.GetByID(ctx, chatID)
		if err != nil && err != errs.ErrGenericNotFound {
			uc.zsLog.Errorf("[CreateChat] error while getting chat by phone number id: %v", err)
			return model.Chat{}, true, err
		} else if err == nil {
			return chat, false, nil
		}
	}
	// create new chat
	newChat := model.Chat{
		PhoneNumberId: requestData.PhoneNumberId,
		RecipientId:   requestData.RecipientId,
		RecipientName: requestData.RecipientName,
		ChatStatus:    model.ChatStatusOpen,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if tenant.ChatType == "ticket" {
		newChat.DocumentID = uuid.New().String()
		newChat.ChatType = "ticket"
	} else {
		newChat.DocumentID = fmt.Sprintf("%s-%s", requestData.PhoneNumberId, requestData.RecipientId)
		newChat.ChatType = "individual"
	}
	logData := model.MessageSystemData{}
	if tenant.ChatType == "ticket" {
		logData.Type = "ticket_created"
		logData.Message = fmt.Sprintf("Ticket created with ID %s", newChat.DocumentID)
	} else {
		logData.Type = "chat_created"
		logData.Message = "Chat created"
	}
	logPayload, err := utils.AnyToJsonString(logData)
	messageLog := model.Message{
		DocumentID:  uuid.New().String(),
		ChatID:      newChat.DocumentID,
		MessageType: "system_flag",
		Payload:     logPayload,
		CreatedAt:   time.Now(),
	}
	var createdChat model.Chat
	serverError, err := uc.txManager.DoFirestore(ctx, func(ctx context.Context, txFirestore *firestore.Transaction) (bool, error) {
		createdChat, _, err = uc.chatRepository.Upsert(ctx, txFirestore, newChat)
		if err != nil {
			uc.zsLog.Errorf("[CreateChat] error while creating chat: %v", err)
			return true, err
		}
		_, err = uc.messageRepository.Upsert(ctx, txFirestore, messageLog)
		if err != nil {
			uc.zsLog.Errorf("[CreateChat] error while creating system message: %v", err)
			return true, err
		}
		task, err := uc.searchMessageRepository.AddDocuments(ctx, []model.Message{messageLog})
		if err != nil {
			uc.zsLog.Errorf("[CreateChat] error while adding message to search index: %v", err)
			return true, err
		}
		for {
			time.Sleep(100 * time.Millisecond)
			task, err := uc.searchMessageRepository.GetTaskStatus(ctx, task.TaskUID)
			if err != nil {
				uc.zsLog.Errorf("[CreateChat] error while getting search index task status: %v", err)
				return true, err
			}
			if task.Status != meilisearch.TaskStatusProcessing {
				if task.Status == meilisearch.TaskStatusFailed {
					uc.zsLog.Errorf("[CreateChat] search index task failed: %v", task.Error)
					return true, fmt.Errorf("search index task failed: %v", task.Error)
				}
				break
			}
		}
		return false, nil
	})
	if err != nil {
		return model.Chat{}, serverError, err
	}
	return createdChat, false, nil
}
