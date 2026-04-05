package model

type Message struct {
	DocumentID      string  `json:"__name__" firestore:"-"`                        // id from whatsapp
	ChatID          string  `json:"chat_id" firestore:"-"`                         // reference to chat document
	MessageType     string  `json:"message_type" firestore:"message_type"`         // text, image, video, etc
	MessageCategory string  `json:"message_category" firestore:"message_category"` // marketing, authentication, utility, service
	SenderName      string  `json:"sender_name" firestore:"sender_name"`           // sender name for individual chat, group name for group chat
	Payload         string  `json:"payload" firestore:"payload"`                   // raw payload from whatsapp, can be used for debugging or future processing
	MediaURL        *string `json:"media_url,omitempty" firestore:"media_url"`     // URL of the associated media file, if any
	Status          string  `json:"status" firestore:"status"`                     // -, sent, delivered, read
	CreatedAt       int64   `json:"created_at" firestore:"created_at"`
	UpdatedAt       int64   `json:"updated_at" firestore:"updated_at"`
}

func (m Message) TableName() string {
	return "messages"
}

func (m Message) AllowedFilterFields() []string {
	return []string{"message_type", "message_category", "sender_name", "status"}
}

func (m Message) AllowedSortFields() []string {
	return []string{"created_at", "updated_at"}
}
