package model

import "time"

type BroadcastRecipient struct {
	DocumentID    string    `json:"__name__" firestore:"-"`     // uuid v7
	BroadcastID   string    `json:"broadcast_id" firestore:"-"` // reference to broadcast document
	RecipientID   string    `json:"recipient_id" firestore:"recipient_id"`
	RecipientName string    `json:"recipient_name" firestore:"recipient_name"`
	RecipientType string    `json:"recipient_type" firestore:"recipient_type"` // individual, group
	ReplyData     *string   `json:"reply_data,omitempty" firestore:"reply_data"`
	Status        string    `json:"status" firestore:"status"` // pending, sent, failed
	CreatedAt     time.Time `json:"created_at" firestore:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" firestore:"updated_at"`
	Errors        *string   `json:"errors,omitempty" firestore:"errors,omitempty"`
}

func (b BroadcastRecipient) TableName() string {
	return "recipients"
}
