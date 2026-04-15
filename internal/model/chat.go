package model

import (
	"strings"
	"time"
)

type Chat struct {
	DocumentID    string    `json:"__name__" firestore:"-"` // {recipient_id}-{phone_number_id}
	PhoneNumberID string    `json:"phone_number_id" firestore:"phone_number_id"`
	RecipientID   string    `json:"recipient_id" firestore:"recipient_id"`
	ChatType      string    `json:"chat_type" firestore:"chat_type"` // individual or group
	DisplayName   string    `json:"display_name" firestore:"display_name"`
	LastMessage   string    `json:"last_message" firestore:"last_message"`
	CreatedAt     time.Time `json:"created_at" firestore:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" firestore:"updated_at"`
}

func (c *Chat) GetRecipientID() string {
	i := strings.LastIndex(c.DocumentID, "-")
	if i != -1 {
		return c.DocumentID[:i]
	}
	return ""
}

func (c *Chat) GetPhoneNumberID() string {
	i := strings.LastIndex(c.DocumentID, "-")
	if i != -1 {
		return c.DocumentID[i+1:]
	}
	return ""
}

func (c *Chat) AllowedFilterFields() []string {
	return []string{"phone_number_id"}
}

func (c *Chat) AllowedSortFields() []string {
	return []string{"created_at"}
}

func (c *Chat) TableName() string {
	return "chats"
}
