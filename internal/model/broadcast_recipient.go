package model

import "time"

type BroadcastRecipient struct {
	DocumentID    string    `json:"id" firestore:"-"`           // uuid v7
	BroadcastID   string    `json:"broadcast_id" firestore:"-"` // reference to broadcast document
	WamID         string    `json:"wam_id" firestore:"wam_id"`
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

func (b BroadcastRecipient) AllowedFilterFields() []string {
	return []string{"broadcast_id"}
}

func (b BroadcastRecipient) AllowedSortFields() []string {
	return []string{"created_at"}
}
