package model

import (
	"time"
)

type Tenant struct {
	DocumentID string         `json:"id" firestore:"-"`
	Name       string         `json:"name" firestore:"name"`
	ChatType   TenantChatType `json:"chat_type" firestore:"chat_type"` // "chat" or "ticket"
	CreatedAt  time.Time      `json:"created_at" firestore:"created_at"`
}

type TenantChatType string

const (
	TenantChatTypeChat   TenantChatType = "chat"
	TenantChatTypeTicket TenantChatType = "ticket"
)

func (t *Tenant) TableName() string {
	return "tenants"
}
