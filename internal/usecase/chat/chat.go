package chat_usecase

import (
	"context"
	"fmt"
	"slices"
	"strings"
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

func (uc *ChatUsecase) CloseTicket(ctx context.Context, tenantID string, requestData dto.ChatCloseTicketRequest) (bool, error) {
	chat, err := uc.chatRepository.GetByID(ctx, requestData.ChatID)
	if err != nil {
		uc.zsLog.Errorf("[CloseTicket] error while getting chat by id: %v", err)
		if err == errs.ErrGenericNotFound {
			return false, err
		}
		return true, err
	}
	if chat.TenantID != tenantID {
		uc.zsLog.Errorf("[CloseTicket] tenant id mismatch: %s vs %s", chat.TenantID, tenantID)
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
	defaultChatID := fmt.Sprintf("%s-%s", requestData.PhoneNumberId, requestData.RecipientId)
	defaultChat, err := uc.chatRepository.GetByID(ctx, defaultChatID)
	if err != nil && err != errs.ErrGenericNotFound {
		uc.zsLog.Errorf("[CreateChat] error while getting chat by phone number id: %v", err)
		return emptyChat, true, err
	}
	defaultChatExists := err == nil
	var messagesToIndex []model.Message

	if tenant.ChatType != "ticket" {
		if defaultChatExists {
			return defaultChat, false, nil
		}

		defaultChat, defaultChatMessage, err := uc.createDefaultChat(tenantID, requestData)
		if err != nil {
			uc.zsLog.Errorf("[CreateChat] error while creating default chat: %v", err)
			return emptyChat, true, err
		}
		defaultChatMessage.CreatedAt = defaultChat.CreatedAt.Add(time.Second)

		serverError, err := uc.txManager.DoFirestore(ctx, func(ctx context.Context, txFirestore *firestore.Transaction) (bool, error) {
			_, _, err = uc.chatRepository.Upsert(ctx, txFirestore, defaultChat)
			if err != nil {
				uc.zsLog.Errorf("[CreateChat] error while creating default chat: %v", err)
				return true, err
			}
			_, err = uc.messageRepository.Upsert(ctx, txFirestore, defaultChatMessage)
			if err != nil {
				uc.zsLog.Errorf("[CreateChat] error while creating system message for default chat: %v", err)
				return true, err
			}
			return false, nil
		})
		if err != nil {
			return emptyChat, serverError, err
		}
		messagesToIndex = append(messagesToIndex, defaultChatMessage)
		if len(messagesToIndex) > 0 {
			err = uc.searchMessageRepository.AddDocumentsSync(ctx, messagesToIndex)
			if err != nil {
				uc.zsLog.Errorf("[CreateChat] error while adding default chat message to search index: %v", err)
				return emptyChat, true, err
			}
		}
		return defaultChat, false, nil
	}

	if !defaultChatExists {
		defaultChat, defaultChatMessage, err := uc.createDefaultChat(tenantID, requestData)
		if err != nil {
			uc.zsLog.Errorf("[CreateChat] error while creating default chat for ticket: %v", err)
			return emptyChat, true, err
		}
		defaultChatMessage.CreatedAt = defaultChat.CreatedAt.Add(time.Second)

		serverError, err := uc.txManager.DoFirestore(ctx, func(ctx context.Context, txFirestore *firestore.Transaction) (bool, error) {
			_, _, err = uc.chatRepository.Upsert(ctx, txFirestore, defaultChat)
			if err != nil {
				uc.zsLog.Errorf("[CreateChat] error while creating default chat for ticket: %v", err)
				return true, err
			}
			_, err = uc.messageRepository.Upsert(ctx, txFirestore, defaultChatMessage)
			if err != nil {
				uc.zsLog.Errorf("[CreateChat] error while creating system message for default chat: %v", err)
				return true, err
			}
			return false, nil
		})
		if err != nil {
			return emptyChat, serverError, err
		}
		messagesToIndex = append(messagesToIndex, defaultChatMessage)
	}

	runningTicketChat, err := uc.chatRepository.GetRunningTicketChat(ctx, requestData.PhoneNumberId, requestData.RecipientId)
	if err != nil && err != errs.ErrGenericNotFound {
		uc.zsLog.Errorf("[CreateChat] error while getting running tickets by phone number id: %v", err)
		return emptyChat, true, err
	} else if err == nil {
		return runningTicketChat, false, nil
	}

	newChatID, err := uuid.NewV7()
	if err != nil {
		uc.zsLog.Errorf("[CreateChat] error while generating chat id: %v", err)
		return emptyChat, true, err
	}
	newChat := model.Chat{
		DocumentID:    newChatID.String(),
		ChatType:      "ticket",
		TenantID:      tenantID,
		PhoneNumberId: requestData.PhoneNumberId,
		RecipientId:   requestData.RecipientId,
		RecipientName: requestData.RecipientName,
		ChatStatus:    model.ChatStatusOpen,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	newChatLogData := model.MessageSystemData{
		Type:    "ticket_created",
		Message: fmt.Sprintf("A new ticket was created with ID %s", newChat.DocumentID),
	}
	logPayload, err := utils.AnyToJsonString(newChatLogData)
	if err != nil {
		uc.zsLog.Errorf("[CreateChat] error while marshalling log payload: %v", err)
		return emptyChat, true, err
	}
	messageLogID, err := uuid.NewV7()
	if err != nil {
		uc.zsLog.Errorf("[CreateChat] error while generating message log id: %v", err)
		return emptyChat, true, err
	}
	messageLog := model.Message{
		DocumentID:  messageLogID.String(),
		ChatID:      newChat.DocumentID,
		MessageType: "system_flag",
		Payload:     logPayload,
		CreatedAt:   newChat.CreatedAt.Add(time.Second),
	}
	serverError, err := uc.txManager.DoFirestore(ctx, func(ctx context.Context, txFirestore *firestore.Transaction) (bool, error) {
		_, _, err = uc.chatRepository.Upsert(ctx, txFirestore, newChat)
		if err != nil {
			uc.zsLog.Errorf("[CreateChat] error while creating ticket chat: %v", err)
			return true, err
		}
		_, err = uc.messageRepository.Upsert(ctx, txFirestore, messageLog)
		if err != nil {
			uc.zsLog.Errorf("[CreateChat] error while creating system message for ticket chat: %v", err)
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return emptyChat, serverError, err
	}
	messagesToIndex = append(messagesToIndex, messageLog)
	if len(messagesToIndex) > 0 {
		err = uc.searchMessageRepository.AddDocumentsSync(ctx, messagesToIndex)
		if err != nil {
			uc.zsLog.Errorf("[CreateChat] error while adding chat messages to search index: %v", err)
			return emptyChat, true, err
		}
	}
	return newChat, false, nil
}

// createDefaultChat creates a default chat with id in format of phoneNumberId-recipientId
func (uc *ChatUsecase) createDefaultChat(tenantID string, requestData dto.ChatCreateRequest) (model.Chat, model.Message, error) {
	chat := model.Chat{
		DocumentID:    fmt.Sprintf("%s-%s", requestData.PhoneNumberId, requestData.RecipientId),
		ChatType:      "individual",
		TenantID:      tenantID,
		PhoneNumberId: requestData.PhoneNumberId,
		RecipientId:   requestData.RecipientId,
		RecipientName: requestData.RecipientName,
		ChatStatus:    model.ChatStatusOpen,
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
		return model.Chat{}, model.Message{}, err
	}
	message := model.Message{
		DocumentID:  uuid.New().String(),
		ChatID:      chat.DocumentID,
		MessageType: "system_flag",
		Payload:     logPayload,
		CreatedAt:   time.Now(),
	}
	return chat, message, nil
}

func (uc *ChatUsecase) GetTicketAnalytics(ctx context.Context, tenantID string, requestData dto.ChatGetTicketAnalyticsRequest) (dto.ChatGetTicketAnalyticsResponse, bool, error) {
	var emptyResponse dto.ChatGetTicketAnalyticsResponse
	if requestData.PhoneNumberIds == nil {
		waba, err := uc.waBusinessAccountRepository.GetByTenantID(ctx, tenantID)
		if err != nil {
			uc.zsLog.Errorf("[GetAnalytics] error while getting WhatsApp Business Accounts by tenant ID: %v", err)
			return emptyResponse, true, err
		}
		var phoneNumbers []string
		for _, account := range waba {
			phones, err := uc.waPhoneRepository.GetByWaBusinessAccountID(ctx, account.DocumentID)
			if err != nil {
				uc.zsLog.Errorf("[GetAnalytics] error while getting phone numbers by WhatsApp Business Account ID: %v", err)
				return emptyResponse, true, err
			}
			for _, phone := range phones {
				phoneNumbers = append(phoneNumbers, phone.PhoneNumberId)
			}
		}
		requestData.PhoneNumberIds = &phoneNumbers
	}

	chatEntries, err := uc.chatRepository.GetChatTicketDataAnalytics(ctx, *requestData.PhoneNumberIds, requestData.StartTime, requestData.EndTime)
	if err != nil {
		uc.zsLog.Errorf("[GetAnalytics] error while fetching chat data for analytics: %v", err)
		return emptyResponse, true, err
	}
	var closedTicket []model.Chat
	var openedTicketCount int
	var inProgressTicketCount int
	for _, chat := range chatEntries {
		switch chat.ChatStatus {
		case model.ChatStatusClosed:
			closedTicket = append(closedTicket, chat)
		case model.ChatStatusOpen:
			openedTicketCount++
		case model.ChatStatusInProgress:
			inProgressTicketCount++
		}
	}
	// sort by ticketClosed time in ascending order
	slices.SortFunc(closedTicket, func(a, b model.Chat) int {
		closeTimeA := a.UpdatedAt.Unix() - a.CreatedAt.Unix()
		closeTimeB := b.UpdatedAt.Unix() - b.CreatedAt.Unix()
		return int(closeTimeA - closeTimeB)
	})
	var longestResolutionTimeMinutes, shortestResolutionTimeMinutes, medianResolutionTimeMinutes, averageResolutionTimeMinutes int
	if len(closedTicket) > 0 {
		// calculate longest and shortest resolution time in minutes
		lastClosedTicket := closedTicket[len(closedTicket)-1]
		longestResolutionTimeMinutes = int((lastClosedTicket.UpdatedAt.Unix() - lastClosedTicket.CreatedAt.Unix()) / int64(time.Minute.Seconds()))

		// calculate shortest resolution time in minutes
		firstClosedTicket := closedTicket[0]
		shortestResolutionTimeMinutes = int((firstClosedTicket.UpdatedAt.Unix() - firstClosedTicket.CreatedAt.Unix()) / int64(time.Minute.Seconds()))

		// calculate median resolution time in minutes
		// if the number of closed tickets is odd, take the middle one, if even take the average of the two middle ones
		if len(closedTicket)%2 == 0 {
			middleIndex1 := len(closedTicket)/2 - 1
			middleIndex2 := len(closedTicket) / 2
			medianResolutionTimeMinutes = int(((closedTicket[middleIndex1].UpdatedAt.Unix() - closedTicket[middleIndex1].CreatedAt.Unix()) + (closedTicket[middleIndex2].UpdatedAt.Unix() - closedTicket[middleIndex2].CreatedAt.Unix())) / 2 / int64(time.Minute.Seconds()))
		} else {
			medianIndex := len(closedTicket) / 2
			medianResolutionTimeMinutes = int((closedTicket[medianIndex].UpdatedAt.Unix() - closedTicket[medianIndex].CreatedAt.Unix()) / int64(time.Minute.Seconds()))
		}
		// calculate average resolution time in minutes
		var totalResolutionTime int64
		for _, chat := range closedTicket {
			totalResolutionTime += chat.UpdatedAt.Unix() - chat.CreatedAt.Unix()
		}
		averageResolutionTimeMinutes = int((totalResolutionTime / int64(len(closedTicket))) / int64(time.Minute.Seconds()))
	}
	response := dto.ChatGetTicketAnalyticsResponse{
		OpenedCount:               openedTicketCount,
		InProgressCount:           inProgressTicketCount,
		ClosedCount:               len(closedTicket),
		TotalCount:                len(chatEntries),
		LongestResolutionMinutes:  longestResolutionTimeMinutes,
		ShortestResolutionMinutes: shortestResolutionTimeMinutes,
		MedianResolutionMinutes:   medianResolutionTimeMinutes,
		AverageResolutionMinutes:  averageResolutionTimeMinutes,
	}
	return response, false, nil
}
