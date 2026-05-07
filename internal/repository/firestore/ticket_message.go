package repository_firestore

import (
	"context"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type TicketMessageRepository struct {
	ticket        model.Ticket
	ticketMessage model.TicketMessage
	db            *firestore.Client
}

func NewTicketMessageRepository(db *firestore.Client) *TicketMessageRepository {
	return &TicketMessageRepository{
		db: db,
	}
}

func (r *TicketMessageRepository) Upsert(ctx context.Context, tx *firestore.Transaction, ticketMessage model.TicketMessage) (model.TicketMessage, error) {
	docRef := r.db.
		Collection(r.ticket.TableName()).Doc(ticketMessage.TicketID).
		Collection(r.ticketMessage.TableName()).Doc(ticketMessage.DocumentID)
	if tx != nil {
		err := tx.Set(docRef, ticketMessage)
		if err != nil {
			return ticketMessage, err
		}
	} else {
		_, err := docRef.Set(ctx, ticketMessage)
		if err != nil {
			return ticketMessage, err
		}
	}
	return ticketMessage, nil
}

func (r *TicketMessageRepository) GetTicketMessageByWamid(ctx context.Context, ticketID string, wamid string) (model.TicketMessage, error) {
	var ticketMessage model.TicketMessage
	doc, err := r.db.
		Collection(r.ticket.TableName()).Doc(ticketID).
		Collection(r.ticketMessage.TableName()).
		Where("wamid", "==", wamid).
		Limit(1).
		Documents(ctx).Next()
	if err != nil {
		if err == iterator.Done {
			return ticketMessage, errs.ErrGenericNotFound
		}
		return ticketMessage, err
	}
	docData := doc.Data()
	docData["id"] = doc.Ref.ID
	err = utils.MapToStruct(docData, &ticketMessage)
	return ticketMessage, err
}
