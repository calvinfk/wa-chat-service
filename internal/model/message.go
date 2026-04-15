package model

import "time"

type Message struct {
	DocumentID      string        `json:"__name__" firestore:"-"`                                  // id from whatsapp
	ChatID          string        `json:"chat_id" firestore:"-"`                                   // reference to chat document
	MessageType     string        `json:"message_type" firestore:"message_type"`                   // text, image, video, etc
	MessageCategory string        `json:"message_category" firestore:"message_category"`           // marketing, authentication, utility, service
	SenderName      string        `json:"sender_name" firestore:"sender_name"`                     // sender name for individual chat, group name for group chat
	Payload         string        `json:"payload" firestore:"payload"`                             // raw payload from whatsapp, can be used for debugging or future processing
	StorageMediaID  *string       `json:"storage_media_id,omitempty" firestore:"storage_media_id"` // reference to media document if message has media
	StorageMedia    *StorageMedia `json:"storage_media,omitempty" firestore:"-"`                   // media document if message has media
	Status          string        `json:"status" firestore:"status"`                               // -, sent, delivered, read
	CreatedAt       time.Time     `json:"created_at" firestore:"created_at"`
	SentAt          *time.Time    `json:"sent_at,omitempty" firestore:"sent_at,omitempty"`
	DeliveredAt     *time.Time    `json:"delivered_at,omitempty" firestore:"delivered_at,omitempty"`
	ReadAt          *time.Time    `json:"read_at,omitempty" firestore:"read_at,omitempty"`
	Error           *string       `json:"error,omitempty" firestore:"error,omitempty"` // error message if failed to send
}

func (m Message) TableName() string {
	return "messages"
}

func (m Message) AllowedFilterFields() []string {
	return []string{"chat_id"}
}

func (m Message) AllowedSortFields() []string {
	return []string{"created_at"}
}
