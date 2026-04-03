package model

type Message struct {
	DocumentID      string `firestore:"-"`                // id from whatsapp
	ChatID          string `firestore:"-"`                // reference to chat document
	MessageType     string `firestore:"message_type"`     // text, image, video, etc
	MessageCategory string `firestore:"message_category"` // marketing, authentication, utility, service
	SenderName      string `firestore:"sender_name"`      // sender name for individual chat, group name for group chat
	Payload         string `firestore:"payload"`          // raw payload from whatsapp, can be used for debugging or future processing
	Content         string `firestore:"content"`          // extracted content from payload, can be used for searching or displaying in UI
	Status          string `firestore:"status"`           // -, sent, delivered, read
	CreatedAt       int64  `firestore:"created_at"`
	UpdatedAt       int64  `firestore:"updated_at"`
}
