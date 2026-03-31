package model

type Message struct {
	DocumentID      string `firestore:"-"`                // id from whatsapp
	ChatID          string `firestore:"-"`                // reference to chat document
	MessageType     string `firestore:"message_type"`     // text, image, video, etc
	MessageCategory string `firestore:"message_category"` // marketing, authentication, utility, service
	SenderName      string `firestore:"sender_name"`      // sender name for individual chat, group name for group chat
	Content         string `firestore:"message"`          // message content for text message, caption for media message
	Status          string `firestore:"status"`           // -, sent, delivered, read
	CreatedAt       int64  `firestore:"created_at"`
	UpdatedAt       int64  `firestore:"updated_at"`
}
