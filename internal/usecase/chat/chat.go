package chat_usecase

import (
	"context"
	"fmt"
	"strings"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/usecase"
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
	tenantRepository            repository.Tenant
	searchMessageRepository     repository.SearchMessage
	storageMediaRepository      repository.StorageMedia
	ticketRepository            repository.Ticket
	waBusinessAccountUsecase    usecase.WaBusinessAccount
	storageMediaUsecase         usecase.StorageMedia
	ticketUsecase               usecase.Ticket
	txManager                   *utils.TxManager
	zsLog                       *zap.SugaredLogger
}

func NewChatUsecase(chatRepository repository.Chat,
	waPhoneRepository repository.WaPhone,
	waBusinessAccountRepository repository.WaBusinessAccount,
	userRepository repository.User,
	messageRepository repository.Message,
	tenantRepository repository.Tenant,
	searchMessageRepository repository.SearchMessage,
	storageMediaRepository repository.StorageMedia,
	ticketRepository repository.Ticket,
	waBusinessAccountUsecase usecase.WaBusinessAccount,
	storageMediaUsecase usecase.StorageMedia,
	ticketUsecase usecase.Ticket,
	txManager *utils.TxManager,
	zsLog *zap.SugaredLogger,
) *ChatUsecase {
	return &ChatUsecase{
		chatRepository:              chatRepository,
		waPhoneRepository:           waPhoneRepository,
		waBusinessAccountRepository: waBusinessAccountRepository,
		userRepository:              userRepository,
		messageRepository:           messageRepository,
		tenantRepository:            tenantRepository,
		searchMessageRepository:     searchMessageRepository,
		storageMediaRepository:      storageMediaRepository,
		ticketRepository:            ticketRepository,
		waBusinessAccountUsecase:    waBusinessAccountUsecase,
		storageMediaUsecase:         storageMediaUsecase,
		ticketUsecase:               ticketUsecase,
		txManager:                   txManager,
		zsLog:                       zsLog,
	}
}

func (uc *ChatUsecase) GetChatByPhoneNumberId(ctx context.Context, authData dto.AuthData, requestData filter_request.FilterRequest[dto.ChatGetByPhoneNumberIdRequest]) (filter_request.FilterResponse[dto.ChatGetByPhoneNumberIdResponse], bool, error) {
	switch authData.Role {
	case model.UserRoleAdmin:
		// no additional filter needed, admins can see all chats
	case model.UserRoleAgent:
		requestData.SpecificFilter.AgentID = &authData.UserID
	case model.UserRoleSupervisor:
		// get all agents under the supervisor
		agents, err := uc.userRepository.GetBySupervisorID(ctx, authData.UserID)
		if err != nil {
			uc.zsLog.Errorf("[GetChatByPhoneNumberId] error while getting agents by supervisor id: %v", err)
			return filter_request.FilterResponse[dto.ChatGetByPhoneNumberIdResponse]{}, true, err
		}
		var agentIDs []string
		for _, agent := range agents {
			agentIDs = append(agentIDs, agent.DocumentID)
		}
		filter := fmt.Sprintf("in:[%s]", strings.Join(agentIDs, ","))
		requestData.SpecificFilter.AgentID = &filter
	default:
		uc.zsLog.Errorf("[GetChatByPhoneNumberId] unauthorized role: %s", authData.Role)
		return filter_request.FilterResponse[dto.ChatGetByPhoneNumberIdResponse]{}, false, errs.ErrGenericForbidden
	}
	phone, err := uc.waPhoneRepository.GetByPhoneNumberId(ctx, requestData.SpecificFilter.PhoneNumberId)
	if err != nil {
		uc.zsLog.Errorf("[GetChatByPhoneNumberId] error while getting phone by id: %v", err)
		return filter_request.FilterResponse[dto.ChatGetByPhoneNumberIdResponse]{}, true, err
	}
	waba, err := uc.waBusinessAccountRepository.GetByID(ctx, phone.WaBusinessAccountID)
	if err != nil {
		uc.zsLog.Errorf("[GetChatByPhoneNumberId] error while getting wa business account by id: %v", err)
		return filter_request.FilterResponse[dto.ChatGetByPhoneNumberIdResponse]{}, true, err
	}
	if waba.TenantID != authData.TenantID {
		uc.zsLog.Errorf("[GetChatByPhoneNumberId] tenant id mismatch: %s vs %s", waba.TenantID, authData.TenantID)
		return filter_request.FilterResponse[dto.ChatGetByPhoneNumberIdResponse]{}, false, errs.ErrGenericForbidden
	}
	response, err := uc.chatRepository.GetChatByPhoneNumberId(ctx, requestData)
	if err != nil {
		uc.zsLog.Errorf("[GetChatByPhoneNumberId] error while getting chat by phone number id: %v", err)
		return filter_request.FilterResponse[dto.ChatGetByPhoneNumberIdResponse]{}, true, err
	}
	return response, false, nil
}

func (uc *ChatUsecase) GetByID(ctx context.Context, chatID string) (model.Chat, bool, error) {
	var emptyChat model.Chat
	chat, err := uc.chatRepository.GetByID(ctx, chatID)
	if err != nil {
		uc.zsLog.Errorf("[GetByID] error while getting chat by id: %v", err)
		if err == errs.ErrGenericNotFound {
			return emptyChat, false, err
		}
		return emptyChat, true, err
	}
	return chat, false, nil
}

func (uc *ChatUsecase) CreateChat(ctx context.Context, tenantID string, requestData dto.ChatCreateRequest) (model.Chat, bool, error) {
	var emptyChat model.Chat
	phone, err := uc.waPhoneRepository.GetByPhoneNumberId(ctx, requestData.PhoneNumberId)
	if err != nil {
		uc.zsLog.Errorf("[CreateChat] error while getting phone by id: %v", err)
		return emptyChat, true, err
	}
	waba, err := uc.waBusinessAccountRepository.GetByID(ctx, phone.WaBusinessAccountID)
	if err != nil {
		uc.zsLog.Errorf("[CreateChat] error while getting wa business account by id: %v", err)
		return emptyChat, true, err
	}
	if waba.TenantID != tenantID {
		uc.zsLog.Errorf("[CreateChat] tenant id mismatch: %s vs %s", waba.TenantID, tenantID)
		return emptyChat, false, errs.ErrGenericForbidden
	}
	tenant, err := uc.tenantRepository.GetByID(ctx, tenantID)
	if err != nil {
		uc.zsLog.Errorf("[CreateChat] error while getting tenant by id: %v", err)
		return emptyChat, true, err
	}
	if tenant.ChatType != "default" {
		return emptyChat, false, fmt.Errorf("tenant chat type is not default")
	}
	defaultChatID := fmt.Sprintf("%s-%s", requestData.PhoneNumberId, requestData.RecipientId)
	defaultChat, err := uc.chatRepository.GetByID(ctx, defaultChatID)
	if err != nil && err != errs.ErrGenericNotFound {
		uc.zsLog.Errorf("[CreateChat] error while getting chat by phone number id: %v", err)
		return emptyChat, true, err
	}
	if err == nil {
		return defaultChat, false, nil
	}
	chat := model.Chat{
		DocumentID:    fmt.Sprintf("%s-%s", requestData.PhoneNumberId, requestData.RecipientId),
		ChatType:      "individual",
		TenantID:      tenantID,
		PhoneNumberId: requestData.PhoneNumberId,
		RecipientId:   requestData.RecipientId,
		RecipientName: requestData.RecipientName,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	logData := model.MessageSystemData{
		Type:    "chat_created",
		Message: "Chat created",
	}
	logPayload, err := utils.AnyToJsonString(logData)
	if err != nil {
		uc.zsLog.Errorf("[createDefaultChat] error while marshalling log payload: %v", err)
		return model.Chat{}, true, err
	}
	message := model.Message{
		DocumentID:  uuid.New().String(),
		ChatID:      chat.DocumentID,
		MessageType: "system_flag",
		Payload:     logPayload,
		CreatedAt:   time.Now(),
	}

	serverError, err := uc.txManager.DoFirestore(ctx, func(ctx context.Context, txFirestore *firestore.Transaction) (bool, error) {
		_, _, err = uc.chatRepository.Upsert(ctx, txFirestore, chat)
		if err != nil {
			uc.zsLog.Errorf("[CreateChat] error while creating default chat: %v", err)
			return true, err
		}
		_, err = uc.messageRepository.Upsert(ctx, txFirestore, message)
		if err != nil {
			uc.zsLog.Errorf("[CreateChat] error while creating system message for default chat: %v", err)
			return true, err
		}
		err = uc.searchMessageRepository.AddDocumentsSync(ctx, []model.Message{message})
		if err != nil {
			uc.zsLog.Errorf("[CreateChat] error while adding chat messages to search index: %v", err)
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return emptyChat, serverError, err
	}
	return chat, false, nil
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
	if chat.AgentID != nil && *chat.AgentID == requestData.AgentID {
		return false, nil
	}
	if chat.TenantID != tenantID {
		uc.zsLog.Errorf("[AssignAgent] tenant id mismatch: %s vs %s", chat.TenantID, tenantID)
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

	chat.ChatStatus = model.ChatStatusInProgress
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
			Type:    "assign_agent",
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
		CreatedAt:   time.Now(),
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
		err = uc.searchMessageRepository.AddDocumentsSync(ctx, []model.Message{message})
		if err != nil {
			uc.zsLog.Errorf("[CloseTicket] error while adding message to search index: %v", err)
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return serverError, err
	}
	return false, nil
}

func (uc *ChatUsecase) CloseChat(ctx context.Context, tenantID string, requestData dto.ChatCloseRequest) (bool, error) {
	chat, err := uc.chatRepository.GetByID(ctx, requestData.ChatID)
	if err != nil {
		uc.zsLog.Errorf("[CloseChat] error while getting chat by id: %v", err)
		if err == errs.ErrGenericNotFound {
			return false, err
		}
		return true, err
	}
	if chat.TenantID != tenantID {
		uc.zsLog.Errorf("[CloseChat] tenant id mismatch: %s vs %s", chat.TenantID, tenantID)
		return false, errs.ErrGenericForbidden
	}
	if chat.ChatStatus == model.ChatStatusClosed {
		return false, nil
	}
	chat.ChatStatus = model.ChatStatusClosed
	chat.UpdatedAt = time.Now()

	messageID, err := uuid.NewV7()
	if err != nil {
		uc.zsLog.Errorf("[CloseChat] error while generating message id: %v", err)
		return true, err
	}
	logPayload, err := utils.AnyToJsonString(model.MessageSystemData{
		Type:    "close_chat",
		Message: "Chat closed",
	})
	if err != nil {
		uc.zsLog.Errorf("[CloseChat] error while marshalling log payload: %v", err)
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
			uc.zsLog.Errorf("[CloseChat] error while upserting chat: %v", err)
			return true, err
		}
		_, err = uc.messageRepository.Upsert(ctx, txFirestore, message)
		if err != nil {
			uc.zsLog.Errorf("[CloseChat] error while creating system message: %v", err)
			return true, err
		}
		err = uc.searchMessageRepository.AddDocumentsSync(ctx, []model.Message{message})
		if err != nil {
			uc.zsLog.Errorf("[CloseChat] error while adding message to search index: %v", err)
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return serverError, err
	}
	return false, nil
}
