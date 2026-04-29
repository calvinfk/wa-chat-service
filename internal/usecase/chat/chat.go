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
	"go.uber.org/zap"
)

type ChatUsecase struct {
	chatRepository              repository.Chat
	waPhoneRepository           repository.WaPhone
	waBusinessAccountRepository repository.WaBusinessAccount
	userRepository              repository.User
	messageRepository           repository.Message
	txManager                   *utils.TxManager
	zsLog                       *zap.SugaredLogger
}

func NewChatUsecase(chatRepository repository.Chat, waPhoneRepository repository.WaPhone, waBusinessAccountRepository repository.WaBusinessAccount, userRepository repository.User, messageRepository repository.Message, txManager *utils.TxManager, zsLog *zap.SugaredLogger) *ChatUsecase {
	return &ChatUsecase{
		chatRepository:              chatRepository,
		waPhoneRepository:           waPhoneRepository,
		waBusinessAccountRepository: waBusinessAccountRepository,
		userRepository:              userRepository,
		messageRepository:           messageRepository,
		txManager:                   txManager,
		zsLog:                       zsLog,
	}
}

func (uc *ChatUsecase) GetChatByPhoneNumberID(ctx context.Context, tenantID string, requestData filter_request.FilterRequest[dto.ChatGetByPhoneNumberIdRequest]) (filter_request.FilterResponse[dto.ChatGetByPhoneNumberIdResponse], bool, error) {
	// TODO: check if phone number belongs to tenant
	response, err := uc.chatRepository.GetChatByPhoneNumberID(ctx, requestData)
	if err != nil {
		uc.zsLog.Errorf("[GetChatByPhoneNumberID] error while getting chat by phone number id: %v", err)
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
