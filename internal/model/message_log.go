package model

import "github.com/google/uuid"

type MessageLog struct {
	DocumentID uuid.UUID `json:"__name__" firestore:"-"`
	ChatID     string    `json:"chat_id" firestore:"-"`             // reference to chat document id
	MessageID  string    `json:"message_id" firestore:"message_id"` // id from whatsapp
	Type       string    `json:"type" firestore:"type"`             // message type {}
	Content    string    `json:"content" firestore:"content"`       // message content
	CreatedAt  int64     `json:"created_at" firestore:"created_at"`
}
