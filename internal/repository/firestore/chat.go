package repository_firestore

import (
	"context"
	"wa_chat_service/internal/model"

	"cloud.google.com/go/firestore"
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
		docRef := r.db.
			Collection("chats").Doc(chat.DocumentID)
		return tx.Set(docRef, chat)
	}
	if tx == nil {
		err = r.db.RunTransaction(ctx, execDB)
	} else {
		err = execDB(ctx, tx)
	}
	return chat, err
}
