package model

import "time"

type TicketMessage struct {
	DocumentID      string        `json:"id" firestore:"-"`                                        // uuid v7
	TicketID        string        `json:"ticket_id" firestore:"-"`                                 // reference to ticket document
	Wamid           string        `json:"wamid" firestore:"wamid"`                                 // id from whatsapp, empty for system generated messages
	MessageType     string        `json:"message_type" firestore:"message_type"`                   // text, image, video, etc
	MessageCategory string        `json:"message_category" firestore:"message_category"`           // marketing, authentication, utility, service, (and system_flag for system generated messages)
	SenderName      string        `json:"sender_name" firestore:"sender_name"`                     // sender name for individual ticket
	Payload         string        `json:"payload" firestore:"payload"`                             // raw payload from whatsapp or system, can be used for debugging or future processing.
	StorageMediaID  *string       `json:"storage_media_id,omitempty" firestore:"storage_media_id"` // reference to media document if message has media
	StorageMedia    *StorageMedia `json:"storage_media,omitempty" firestore:"-"`                   // media document if message has media
	Status          string        `json:"status" firestore:"status"`                               // -, sent, delivered, read
	CreatedAt       time.Time     `json:"created_at" firestore:"created_at"`
	SentAt          *time.Time    `json:"sent_at,omitempty" firestore:"sent_at,omitempty"`
	DeliveredAt     *time.Time    `json:"delivered_at,omitempty" firestore:"delivered_at,omitempty"`
	ReadAt          *time.Time    `json:"read_at,omitempty" firestore:"read_at,omitempty"`
	Error           *string       `json:"error,omitempty" firestore:"error,omitempty"` // error message if failed to send
}

func (m TicketMessage) TableName() string {
	return "ticket_messages"
}

func (m TicketMessage) PKName() string {
	return "id"
}

func (m TicketMessage) AllowedFilterFields() []string {
	return []string{"ticket_id", "message_type", "message_category", "status"}
}

func (m TicketMessage) AllowedSortFields() []string {
	return []string{"created_at"}
}

type TicketMessageSystemData struct {
	Type    string `json:"type"`    // system_flag for system generated messages, e.g. "chat_closed", "agent_assigned", etc
	Message string `json:"message"` // message content for system generated messages, can be used for display
}
