package model

import (
	"time"
)

type Chat struct {
	DocumentID        string     `json:"id" firestore:"-"` // {recipient_id}-{phone_number_id}, or uuid v7 for group and ticket
	PhoneNumberId     string     `json:"phone_number_id" firestore:"phone_number_id"`
	RecipientId       string     `json:"recipient_id" firestore:"recipient_id"`
	RecipientName     string     `json:"recipient_name" firestore:"recipient_name"`
	TenantID          string     `json:"tenant_id" firestore:"tenant_id"`
	AgentID           *string    `json:"agent_id,omitempty" firestore:"agent_id,omitempty"`
	ChatType          string     `json:"chat_type" firestore:"chat_type"`     // individual, group
	ChatStatus        ChatStatus `json:"chat_status" firestore:"chat_status"` // open, in_progress, closed
	LastMessage       string     `json:"last_message" firestore:"last_message"`
	UserLastMessageAt *time.Time `json:"user_last_message_at" firestore:"user_last_message_at"` // to calculate csw
	CreatedAt         time.Time  `json:"created_at" firestore:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" firestore:"updated_at"`
}

type ChatStatus string

const (
	ChatStatusOpen       ChatStatus = "open"
	ChatStatusInProgress ChatStatus = "in_progress"
	ChatStatusClosed     ChatStatus = "closed"
)

func (c *Chat) AllowedFilterFields() []string {
	return []string{"phone_number_id", "chat_status", "chat_type", "agent_id"}
}

func (c *Chat) AllowedSortFields() []string {
	return []string{"created_at"}
}

func (c *Chat) TableName() string {
	return "chats"
}
