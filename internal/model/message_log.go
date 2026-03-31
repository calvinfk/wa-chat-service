package model

import "github.com/google/uuid"

type MessageLog struct {
	DocumentID uuid.UUID `firestore:"-"`
	ChatID     string    `firestore:"-"`          // reference to chat document id
	MessageID  string    `firestore:"message_id"` // id from whatsapp
	Type       string    `firestore:"type"`       // message type {}
	Content    string    `firestore:"content"`    // message content
	CreatedAt  int64     `firestore:"created_at"`
}
