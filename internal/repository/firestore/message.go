package repository_firestore

import (
	"context"
	"wa_chat_service/internal/model"

	"cloud.google.com/go/firestore"
)

type MessageRepository struct {
	message    model.Message
	messageLog model.MessageLog
	db         *firestore.Client
}

func NewMessageRepository(db *firestore.Client) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Upsert(ctx context.Context, tx *firestore.Transaction, message model.Message) (model.Message, error) {
	var err error
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		docRef := r.db.
			Collection("chats").Doc(message.ChatID).
			Collection("messages").Doc(message.DocumentID)
		err := tx.Set(docRef, message)
		if err != nil {
			return err
		}
		return nil
	}
	if tx == nil {
		err = r.db.RunTransaction(ctx, execDB)
	} else {
		err = execDB(ctx, tx)
	}
	return message, err
}

func (r *MessageRepository) InsertLog(ctx context.Context, tx *firestore.Transaction, messageLog model.MessageLog) (model.MessageLog, error) {
	var err error
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		_, err := r.db.
			Collection("chats").Doc(messageLog.ChatID).
			Collection("messages_log").Doc(messageLog.DocumentID.String()).
			Set(ctx, messageLog)
		if err != nil {
			return err
		}
		return nil
	}
	if tx == nil {
		err = r.db.RunTransaction(ctx, execDB)
	} else {
		err = execDB(ctx, tx)
	}
	return messageLog, err
}
