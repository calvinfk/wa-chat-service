package ticket_usecase

import (
	"context"
	"fmt"
	"slices"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TicketUsecase struct {
	ticketRepository              repository.Ticket
	ticketMessageRepository       repository.TicketMessage
	searchTicketMessageRepository repository.SearchTicketMessage
	waBusinessAccountRepository   repository.WaBusinessAccount
	waPhoneRepository             repository.WaPhone
	userRepository                repository.User
	googleTaskService             service.GoogleTask
	txManager                     *utils.TxManager
	zsLog                         *zap.SugaredLogger
}

// NewTicketUsecase wires ticket-related repositories and utilities into a TicketUsecase.
func NewTicketUsecase(
	ticketRepository repository.Ticket,
	ticketMessageRepository repository.TicketMessage,
	searchTicketMessageRepository repository.SearchTicketMessage,
	waBusinessAccountRepository repository.WaBusinessAccount,
	waPhoneRepository repository.WaPhone,
	userRepository repository.User,
	googleTaskService service.GoogleTask,
	txManager *utils.TxManager,
	zsLog *zap.SugaredLogger,
) *TicketUsecase {
	return &TicketUsecase{
		ticketRepository:              ticketRepository,
		ticketMessageRepository:       ticketMessageRepository,
		searchTicketMessageRepository: searchTicketMessageRepository,
		waBusinessAccountRepository:   waBusinessAccountRepository,
		waPhoneRepository:             waPhoneRepository,
		userRepository:                userRepository,
		googleTaskService:             googleTaskService,
		txManager:                     txManager,
		zsLog:                         zsLog,
	}
}

// GetTicketAnalytics aggregates ticket counts and resolution-time metrics for a tenant.
//
// If requestData.PhoneNumberIds is empty, all phone numbers under the tenant's WABA accounts
// are resolved and used as the analytics scope.
//
// Returns:
// - dto.TicketGetAnalyticsResponse with total/open/in-progress/closed counters and resolution stats.
// - serverError=true when the operation fails because of internal/repository issues.
// - error describing the failure.
func (uc *TicketUsecase) GetTicketAnalytics(ctx context.Context, tenantID string, requestData dto.TicketGetAnalyticsRequest) (dto.TicketGetAnalyticsResponse, bool, error) {
	var emptyResponse dto.TicketGetAnalyticsResponse
	// If phone number IDs are not provided, resolve all phone numbers under the tenant's WABA accounts.
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

	ticketEntries, err := uc.ticketRepository.GetTicketDataAnalytics(ctx, *requestData.PhoneNumberIds, requestData.StartTime, requestData.EndTime)
	if err != nil {
		uc.zsLog.Errorf("[GetAnalytics] error while fetching ticket data for analytics: %v", err)
		return emptyResponse, true, err
	}
	var closedTicket []model.Ticket
	var openedTicketCount int
	var inProgressTicketCount int
	for _, ticket := range ticketEntries {
		switch ticket.TicketStatus {
		case model.TicketStatusClosed:
			closedTicket = append(closedTicket, ticket)
		case model.TicketStatusOpen:
			openedTicketCount++
		case model.TicketStatusInProgress:
			inProgressTicketCount++
		}
	}
	// Resolution-time metrics are computed from closed tickets only.
	// sort by ticketClosed time in ascending order
	slices.SortFunc(closedTicket, func(a, b model.Ticket) int {
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
		for _, ticket := range closedTicket {
			totalResolutionTime += ticket.UpdatedAt.Unix() - ticket.CreatedAt.Unix()
		}
		averageResolutionTimeMinutes = int((totalResolutionTime / int64(len(closedTicket))) / int64(time.Minute.Seconds()))
	}
	response := dto.TicketGetAnalyticsResponse{
		OpenedCount:               openedTicketCount,
		InProgressCount:           inProgressTicketCount,
		ClosedCount:               len(closedTicket),
		TotalCount:                len(ticketEntries),
		LongestResolutionMinutes:  longestResolutionTimeMinutes,
		ShortestResolutionMinutes: shortestResolutionTimeMinutes,
		MedianResolutionMinutes:   medianResolutionTimeMinutes,
		AverageResolutionMinutes:  averageResolutionTimeMinutes,
	}
	return response, false, nil
}

// CloseTicket sets a ticket status to closed and appends a system-flag message.
//
// The status update and system message creation are executed in one Firestore transaction,
// then the message is indexed for search within the same unit of work.
func (uc *TicketUsecase) CloseTicket(ctx context.Context, tenantID string, requestData dto.TicketCloseRequest) (bool, error) {
	ticket, err := uc.ticketRepository.GetByID(ctx, requestData.TicketID)
	if err != nil {
		uc.zsLog.Errorf("[CloseTicket] error while getting ticket by id: %v", err)
		if err == errs.ErrGenericNotFound {
			return false, err
		}
		return true, err
	}
	if ticket.TenantID != tenantID {
		uc.zsLog.Errorf("[CloseTicket] tenant id mismatch: %s vs %s", ticket.TenantID, tenantID)
		return false, errs.ErrGenericForbidden
	}

	// if ticket is already closed, do nothing and return success, otherwise update status to closed.
	if ticket.TicketStatus == model.TicketStatusClosed {
		return false, nil
	}
	ticket.TicketStatus = model.TicketStatusClosed
	ticket.UpdatedAt = time.Now()

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
	message := model.TicketMessage{
		DocumentID:  messageID.String(),
		TicketID:    ticket.DocumentID,
		MessageType: "system_flag",
		Payload:     logPayload,
	}
	// Keep status transition and message creation atomic.
	serverError, err := uc.txManager.DoFirestore(ctx, func(ctx context.Context, txFirestore *firestore.Transaction) (bool, error) {
		_, _, err = uc.ticketRepository.Upsert(ctx, txFirestore, ticket)
		if err != nil {
			uc.zsLog.Errorf("[CloseTicket] error while upserting ticket: %v", err)
			return true, err
		}
		err = uc.ticketMessageRepository.Upsert(ctx, txFirestore, message)
		if err != nil {
			uc.zsLog.Errorf("[CloseTicket] error while creating system message: %v", err)
			return true, err
		}
		err = uc.searchTicketMessageRepository.AddDocumentsSync(ctx, []model.TicketMessage{message})
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

// AssignAgent assigns or reassigns a ticket to an agent and records a system-flag audit message.
//
// Validation includes ticket ownership, ticket state, agent tenant ownership, and agent role.
// Ticket update and system-message insert are performed atomically in Firestore.
func (uc *TicketUsecase) AssignAgent(ctx context.Context, tenantID string, requestData dto.TicketAssignAgentRequest) (bool, error) {
	ticket, err := uc.ticketRepository.GetByID(ctx, requestData.TicketID)
	if err != nil {
		uc.zsLog.Errorf("[AssignAgent] error while getting ticket by id: %v", err)
		if err == errs.ErrGenericNotFound {
			return false, err
		}
		return true, err
	}
	if ticket.TicketStatus == model.TicketStatusClosed {
		return false, fmt.Errorf("cannot assign agent to closed ticket")
	}
	if ticket.AgentID != nil && *ticket.AgentID == requestData.AgentID {
		return false, nil
	}
	if ticket.TenantID != tenantID {
		uc.zsLog.Errorf("[AssignAgent] tenant id mismatch: %s vs %s", ticket.TenantID, tenantID)
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

	if ticket.AgentID != nil {
		prevAgent, err := uc.userRepository.GetByID(ctx, *ticket.AgentID)
		if err != nil {
			uc.zsLog.Errorf("[AssignAgent] error while getting previous agent by id: %v", err)
			return true, err
		}
		prevAgentName = prevAgent.Name
	}

	ticket.TicketStatus = model.TicketStatusInProgress
	ticket.AgentID = &requestData.AgentID
	ticket.UpdatedAt = time.Now()

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
			Message: fmt.Sprintf("Agent %s assigned to ticket", agent.Name),
		})
	}
	if err != nil {
		uc.zsLog.Errorf("[AssignAgent] error while marshalling log payload: %v", err)
		return true, err
	}
	message := model.TicketMessage{
		DocumentID:  messageID.String(),
		TicketID:    ticket.DocumentID,
		MessageType: "system_flag",
		Payload:     logPayload,
		CreatedAt:   time.Now(),
	}
	// Keep assignment change and audit message consistent in a single transaction.
	serverError, err := uc.txManager.DoFirestore(ctx, func(ctx context.Context, txFirestore *firestore.Transaction) (bool, error) {
		_, _, err = uc.ticketRepository.Upsert(ctx, txFirestore, ticket)
		if err != nil {
			uc.zsLog.Errorf("[AssignAgent] error while upserting ticket: %v", err)
			return true, err
		}
		err = uc.ticketMessageRepository.Upsert(ctx, txFirestore, message)
		if err != nil {
			uc.zsLog.Errorf("[AssignAgent] error while creating system message: %v", err)
			return true, err
		}
		err = uc.searchTicketMessageRepository.AddDocumentsSync(ctx, []model.TicketMessage{message})
		if err != nil {
			uc.zsLog.Errorf("[AssignAgent] error while adding message to search index: %v", err)
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return serverError, err
	}
	return false, nil
}

func (uc *TicketUsecase) RemindSLA(ctx context.Context) (bool, error) {
	// Create new google task for a new reminder
	respondTime := 5 * time.Minute
	err := uc.googleTaskService.CreateReminderSLATask(time.Now().Add(respondTime))
	if err != nil {
		uc.zsLog.Errorf("[RemindSLA] error while creating Google Task for next reminder: %v", err)
		return true, err
	}
	tickets, err := uc.ticketRepository.GetTicketsNeedAttention(ctx, respondTime)
	if err != nil {
		uc.zsLog.Errorf("[RemindSLA] error while fetching tickets needing supervisor attention: %v", err)
		return true, err
	}
	if len(tickets) == 0 {
		uc.zsLog.Infof("[RemindSLA] no tickets need supervisor attention at this time")
		return false, nil
	}
	// For simplicity, we just log the reminder here. In a real implementation, this could be an email or in-app notification to the supervisor.
	for _, ticket := range tickets {
		lastAgentMessageAt := ticket.CreatedAt
		if ticket.AgentLastMessageAt != nil {
			lastAgentMessageAt = *ticket.AgentLastMessageAt
		}
		// If the ticket has an assigned agent, remind the supervisor to follow up with the agent.
		// If not, remind the admin to assign the ticket to an agent.
		if ticket.AgentID != nil {
			// Remind the supervisor to follow up with the assigned agent
			agent, err := uc.userRepository.GetByID(ctx, *ticket.AgentID)
			if err != nil {
				uc.zsLog.Errorf("[RemindSLA] error while fetching agent info for ticket %s: %v", ticket.DocumentID, err)
				continue
			}
			if agent.SupervisorID == nil {
				uc.zsLog.Warnf("[RemindSLA] Agent %s assigned to ticket %s has no supervisor assigned. Please assign a supervisor to the agent.", agent.Name, ticket.DocumentID)
				continue
			}
			supervisor, err := uc.userRepository.GetByID(ctx, *agent.SupervisorID)
			if err != nil {
				uc.zsLog.Errorf("[RemindSLA] error while fetching supervisor info for agent %s: %v", agent.DocumentID, err)
				continue
			}
			uc.zsLog.Warnf("[RemindSLA] Supervisor %s, please follow up on ticket %s assigned to agent %s that has not received a response for %v since last agent message at %v.", supervisor.Name, ticket.DocumentID, agent.Name, time.Since(lastAgentMessageAt), lastAgentMessageAt)
		} else {
			// Remind the admin to assign the ticket to an agent
			uc.zsLog.Warnf("[RemindSLA] Admin, please assign ticket %s to an agent as it has not been assigned for %v since creation at %v.", ticket.DocumentID, time.Since(ticket.CreatedAt), ticket.CreatedAt)
		}
	}
	return false, nil
}
