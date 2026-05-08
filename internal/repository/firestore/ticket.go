package repository_firestore

import (
	"context"
	"time"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TicketRepository struct {
	ticket model.Ticket
	db     *firestore.Client
}

func NewTicketRepository(db *firestore.Client) *TicketRepository {
	return &TicketRepository{
		db: db,
	}
}

func (r *TicketRepository) Upsert(ctx context.Context, tx *firestore.Transaction, ticket model.Ticket) (model.Ticket, bool, error) {
	created := false
	docRef := r.db.Collection(r.ticket.TableName()).Doc(ticket.DocumentID)
	_, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return r.ticket, created, err
		}
		created = true
	}
	updates := []firestore.Update{
		{Path: "agent_id", Value: ticket.AgentID},
		{Path: "last_message", Value: ticket.LastMessage},
		{Path: "updated_at", Value: ticket.UpdatedAt},
	}
	if ticket.TicketStatus != "" {
		updates = append(updates, firestore.Update{Path: "ticket_status", Value: ticket.TicketStatus})
	}
	if ticket.UserLastMessageAt != nil {
		updates = append(updates, firestore.Update{Path: "user_last_message_at", Value: ticket.UserLastMessageAt})
	}
	if ticket.AgentLastMessageAt != nil {
		updates = append(updates, firestore.Update{Path: "agent_last_message_at", Value: ticket.AgentLastMessageAt})
	}
	if tx != nil {
		if created {
			err = tx.Set(docRef, ticket)
		} else {
			err = tx.Update(docRef, updates)
		}
	} else {
		if created {
			_, err = docRef.Set(ctx, ticket)
		} else {
			_, err = docRef.Update(ctx, updates)
		}
	}
	return ticket, created, err
}

func (r *TicketRepository) GetTicketDataAnalytics(ctx context.Context, phoneNumberIds []string, startTime time.Time, endTime time.Time) ([]model.Ticket, error) {
	var tickets []model.Ticket
	docs, err := r.db.Collection(r.ticket.TableName()).
		Where("phone_number_id", "in", phoneNumberIds).
		Where("created_at", ">=", startTime).
		Where("created_at", "<=", endTime).
		OrderBy("created_at", firestore.Desc).
		Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	for _, doc := range docs {
		var ticket model.Ticket
		docData := doc.Data()
		docData["id"] = doc.Ref.ID
		err = utils.MapToStruct(docData, &ticket)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, ticket)
	}
	return tickets, nil
}

func (r *TicketRepository) GetRunningTicket(ctx context.Context, phoneNumberId string, recipientId string) (model.Ticket, error) {
	doc, err := r.db.Collection(r.ticket.TableName()).
		Where("phone_number_id", "==", phoneNumberId).
		Where("recipient_id", "==", recipientId).
		Where("ticket_status", "in", []model.TicketStatus{model.TicketStatusOpen, model.TicketStatusInProgress}).
		OrderBy("created_at", firestore.Desc).
		Limit(1).Documents(ctx).Next()
	if err != nil {
		if err == iterator.Done {
			return r.ticket, errs.ErrGenericNotFound
		}
		return r.ticket, err
	}
	var ticket model.Ticket
	docData := doc.Data()
	docData["id"] = doc.Ref.ID
	err = utils.MapToStruct(docData, &ticket)
	if err != nil {
		return r.ticket, err
	}
	return ticket, nil
}

func (r *TicketRepository) GetByID(ctx context.Context, ticketID string) (model.Ticket, error) {
	doc, err := r.db.Collection(r.ticket.TableName()).Doc(ticketID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return r.ticket, errs.ErrGenericNotFound
		}
		return r.ticket, err
	}
	var ticket model.Ticket
	docData := doc.Data()
	docData["id"] = doc.Ref.ID
	err = utils.MapToStruct(docData, &ticket)
	if err != nil {
		return r.ticket, err
	}
	return ticket, nil
}

func (r *TicketRepository) UpdateLastMessage(ctx context.Context, tx *firestore.Transaction, ticketID string, lastMessage string) error {
	docRef := r.db.Collection(r.ticket.TableName()).Doc(ticketID)
	updates := []firestore.Update{
		{Path: "last_message", Value: lastMessage},
		{Path: "updated_at", Value: firestore.ServerTimestamp},
	}
	if tx != nil {
		return tx.Update(docRef, updates)
	}
	_, err := docRef.Update(ctx, updates)
	return err
}

func (r *TicketRepository) GetTicketsNeedAttention(ctx context.Context, respondTime time.Duration) ([]model.Ticket, error) {
	var tickets []model.Ticket
	iter := r.db.Collection(r.ticket.TableName()).
		Where("ticket_status", "in", []model.TicketStatus{model.TicketStatusOpen, model.TicketStatusInProgress}).
		Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var ticket model.Ticket
		docData := doc.Data()
		docData["id"] = doc.Ref.ID
		err = utils.MapToStruct(docData, &ticket)
		if err != nil {
			return nil, err
		}
		currentTime := time.Now()
		if ticket.UserLastMessageAt == nil || currentTime.Sub(*ticket.UserLastMessageAt) < respondTime {
			continue // skip tickets that don't have user messages or haven't exceeded max response time
		}
		// check if agent has responded after the last user message
		if ticket.AgentLastMessageAt != nil && ticket.AgentLastMessageAt.After(*ticket.UserLastMessageAt) {
			continue // skip tickets that have agent responses after the last user message
		}

		tickets = append(tickets, ticket)
	}
	return tickets, nil
}
