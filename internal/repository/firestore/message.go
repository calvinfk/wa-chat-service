package repository_firestore

import (
	"context"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type MessageRepository struct {
	chat    model.Chat
	message model.Message
	db      *firestore.Client
}

func NewMessageRepository(db *firestore.Client) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Upsert(ctx context.Context, tx *firestore.Transaction, message model.Message) (model.Message, error) {
	docRef := r.db.
		Collection(r.chat.TableName()).Doc(message.ChatID).
		Collection(r.message.TableName()).Doc(message.DocumentID)
	var err error
	if tx != nil {
		err = tx.Set(docRef, message)
	} else {
		_, err = docRef.Set(ctx, message)
	}
	return message, err
}
func (r *MessageRepository) GetMessageByWamid(ctx context.Context, chatID string, wamid string) (model.Message, error) {
	var message model.Message
	doc, err := r.db.
		Collection(r.chat.TableName()).Doc(chatID).
		Collection(r.message.TableName()).
		Where("wamid", "==", wamid).Limit(1).Documents(ctx).Next()
	if err != nil {
		if err == iterator.Done {
			return message, errs.ErrGenericNotFound
		}
		return message, err
	}
	docData := doc.Data()
	docData["id"] = doc.Ref.ID
	docData["chat_id"] = doc.Ref.Parent.Parent.ID
	err = utils.MapToStruct(docData, &message)

	return message, err
}
