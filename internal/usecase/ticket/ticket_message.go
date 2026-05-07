package ticket_usecase

import (
	"context"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

// SaveTicketMessage saves a chat message and ensures it is attached to an open ticket.
//
// Behavior summary:
// - Reuses an existing ticket when TicketID is provided or when a running ticket exists.
// - Creates a new open ticket when no running ticket is found.
// - Inserts a one-time "chat_created" system message when the chat is newly created.
// - Upserts both ticket and message in one Firestore transaction.
// - Indexes persisted messages into search after the transaction completes.
//
// Returns:
// - serverError=true when an internal/server-side failure happens.
// - error with the underlying cause.
func (uc *TicketUsecase) SaveTicketMessage(ctx context.Context, tenantID string, inputData dto.MessageSaveRequest) (bool, error) {
	var ticketID string
	var ticket model.Ticket
	var err error
	// if ticket ID is provided, get the ticket by ID,
	// if not provided, try to get the running ticket by phone number ID and recipient ID,
	// if not found, create a new ticket
	if inputData.TicketID != nil && *inputData.TicketID != "" {
		ticket, err = uc.ticketRepository.GetByID(ctx, *inputData.TicketID)
		if err != nil {
			uc.zsLog.Errorf("[SaveTicketMessage] Failed to get ticket by ID: %v", err)
			return err != errs.ErrGenericNotFound, err
		}
		ticketID = ticket.DocumentID
	} else {
		ticket, err = uc.ticketRepository.GetRunningTicket(ctx, inputData.PhoneNumberId, inputData.RecipientId)
		switch err {
		case nil:
			ticketID = ticket.DocumentID
		case errs.ErrGenericNotFound:
			newTicketID, err := uuid.NewV7()
			if err != nil {
				uc.zsLog.Errorf("[SaveTicketMessage] Failed to generate ticket ID: %v", err)
				return true, err
			}
			ticketID = newTicketID.String()
		default:
			uc.zsLog.Errorf("[SaveTicketMessage] Failed to get opened ticket: %v", err)
			return true, err
		}
	}
	if ticket.DocumentID == "" {
		// create new ticket if not exist
		ticket = model.Ticket{
			DocumentID:    ticketID,
			TenantID:      tenantID,
			PhoneNumberId: inputData.PhoneNumberId,
			RecipientId:   inputData.RecipientId,
			RecipientName: inputData.RecipientName,
			TicketStatus:  model.TicketStatusOpen,
			CreatedAt:     inputData.CreatedAt,
		}
	}
	// if recipient name is empty, set recipient name to recipient ID to make sure the ticket can be displayed in the UI, and update recipient name when there is a new message with recipient name in the future.
	if ticket.RecipientName == "" {
		ticket.RecipientName = inputData.RecipientId
	}
	if inputData.UserLastMessageAt != nil {
		ticket.UserLastMessageAt = inputData.UserLastMessageAt
	}
	if inputData.LastMessage != "" {
		ticket.LastMessage = inputData.LastMessage
	}
	ticket.UpdatedAt = time.Now()

	messageLogID, err := uuid.NewV7()
	if err != nil {
		uc.zsLog.Errorf("[SaveMessage] Failed to generate message log ID: %v", err)
		return true, err
	}
	// prepare log message if the message is the first message in the chat
	logData := model.MessageSystemData{
		Type:    "chat_created",
		Message: "Chat created",
	}
	logPayload, err := utils.AnyToJsonString(logData)
	if err != nil {
		uc.zsLog.Errorf("[SaveMessage] Failed to convert log payload to JSON: %v", err)
		return true, err
	}
	ticketMessageLog := model.TicketMessage{
		DocumentID:      messageLogID.String(),
		Wamid:           "",
		TicketID:        ticketID,
		MessageType:     "",
		MessageCategory: "system_flag",
		SenderName:      "",
		Payload:         logPayload,
		StorageMediaID:  nil,
		StorageMedia:    nil,
		Status:          "",
		CreatedAt:       inputData.CreatedAt.Add(-1 * time.Second), // set created time before the first message to make sure the log message is always before the first message in the chat
	}
	ticketMessage := model.TicketMessage{
		Wamid:           inputData.Wamid,
		TicketID:        ticketID,
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
		ticketMessage.DocumentID = *inputData.ID
	} else {
		newMessageID, genErr := uuid.NewV7()
		if genErr != nil {
			uc.zsLog.Errorf("[SaveMessage] Failed to generate message ID: %v", genErr)
			return true, genErr
		}
		ticketMessage.DocumentID = newMessageID.String()
		if ticketMessage.CreatedAt.IsZero() {
			ticketMessage.CreatedAt = time.Now()
		}
	}
	var messagesToIndex []model.TicketMessage
	// Firestore write is atomic for ticket/message persistence.
	// Search indexing is intentionally performed after commit to batch index only successfully persisted messages, with a best-effort approach (indexing failure does not block the main flow).
	serverError, err := uc.txManager.DoFirestore(ctx, func(ctx context.Context, txFirestore *firestore.Transaction) (bool, error) {
		_, isNewChat, err := uc.ticketRepository.Upsert(ctx, txFirestore, ticket)
		if err != nil {
			uc.zsLog.Errorf("[SaveMessage] Failed to upsert ticket: %v", err)
			return true, err
		}
		if isNewChat {
			err = uc.ticketMessageRepository.Upsert(ctx, txFirestore, ticketMessageLog)
			if err != nil {
				uc.zsLog.Errorf("[SaveMessage] Failed to upsert ticket message: %v", err)
				return true, err
			}
			// TODO: check if adding system log message to search index is necessary or not
			messagesToIndex = append(messagesToIndex, ticketMessageLog)
		}
		err = uc.ticketMessageRepository.Upsert(ctx, txFirestore, ticketMessage)
		if err != nil {
			uc.zsLog.Errorf("[SaveMessage] Failed to save ticket message: %v", err)
			return true, err
		}
		messagesToIndex = append(messagesToIndex, ticketMessage)
		return false, nil
	})
	// Best-effort indexing: persistence success is still returned even if indexing fails.
	_, err = uc.searchTicketMessageRepository.AddDocuments(ctx, messagesToIndex)
	if err != nil {
		uc.zsLog.Errorf("[SaveMessage] Failed to add ticket message to search index: %v", err)
	}
	return serverError, err
}

// GetTicketMessageByWamid retrieves a message by WhatsApp message ID from the caller's running ticket.
//
// It first resolves the running ticket by phone number and recipient, verifies tenant ownership,
// then looks up the message by WAMID inside that ticket scope.
func (uc *TicketUsecase) GetTicketMessageByWamid(ctx context.Context, tenantID string, phoneNumberId string, recipientId string, wamid string) (model.TicketMessage, bool, error) {
	ticket, err := uc.ticketRepository.GetRunningTicket(ctx, phoneNumberId, recipientId)
	if err != nil {
		uc.zsLog.Errorf("[GetTicketMessageByWamid] Failed to get running ticket: %v", err)
		return model.TicketMessage{}, err != errs.ErrGenericNotFound, err
	}
	if ticket.TenantID != tenantID {
		uc.zsLog.Errorf("[GetTicketMessageByWamid] Tenant ID mismatch: expected %s, got %s", tenantID, ticket.TenantID)
		return model.TicketMessage{}, true, errs.ErrGenericForbidden
	}
	ticketMessage, err := uc.ticketMessageRepository.GetTicketMessageByWamid(ctx, ticket.DocumentID, wamid)
	if err != nil {
		uc.zsLog.Errorf("[GetTicketMessageByWamid] Failed to get ticket message by WAMID: %v", err)
		return model.TicketMessage{}, err != errs.ErrGenericNotFound, err
	}
	return ticketMessage, false, nil
}
