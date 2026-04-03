package repository_firestore

import (
	"context"
	"wa_chat_service/internal/model"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChatRepository struct {
	chat model.Chat
	db   *firestore.Client
}

func NewChatRepository(db *firestore.Client) *ChatRepository {
	return &ChatRepository{
		db: db,
	}
}

func (r *ChatRepository) Insert(ctx context.Context, tx *firestore.Transaction, chat model.Chat) (model.Chat, error) {
	var err error
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		doc := r.db.Collection("chats").Doc(chat.DocumentID)
		err := tx.Update(doc, []firestore.Update{
			{Path: "last_message", Value: chat.LastMessage},
			{Path: "updated_at", Value: chat.UpdatedAt},
		})
		if err == nil {
			return nil
		}
		if status.Code(err) != codes.NotFound {
			return err
		}
		err = tx.Set(doc, chat)
		return err
	}
	if tx == nil {
		err = r.db.RunTransaction(ctx, execDB)
	} else {
		err = execDB(ctx, tx)
	}
	return chat, err
}
