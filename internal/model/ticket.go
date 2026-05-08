package model

import "time"

type Ticket struct {
	DocumentID         string       `json:"id" firestore:"-"` // {recipient_id}-{phone_number_id}, or uuid v7 for group and ticket
	PhoneNumberId      string       `json:"phone_number_id" firestore:"phone_number_id"`
	RecipientId        string       `json:"recipient_id" firestore:"recipient_id"`
	RecipientName      string       `json:"recipient_name" firestore:"recipient_name"`
	TenantID           string       `json:"tenant_id" firestore:"tenant_id"`
	AgentID            *string      `json:"agent_id,omitempty" firestore:"agent_id,omitempty"`
	LastMessage        string       `json:"last_message" firestore:"last_message"`
	UserLastMessageAt  *time.Time   `json:"user_last_message_at" firestore:"user_last_message_at"`   // to calculate csw
	AgentLastMessageAt *time.Time   `json:"agent_last_message_at" firestore:"agent_last_message_at"` // to calculate sla
	TicketStatus       TicketStatus `json:"ticket_status" firestore:"ticket_status"`
	CreatedAt          time.Time    `json:"created_at" firestore:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at" firestore:"updated_at"`
}

type TicketStatus string

const (
	TicketStatusOpen       TicketStatus = "open"
	TicketStatusInProgress TicketStatus = "in_progress"
	TicketStatusClosed     TicketStatus = "closed"
)

func (t *Ticket) TableName() string {
	return "tickets"
}

func (t *Ticket) AllowedFilterFields() []string {
	return []string{"phone_number_id", "ticket_status", "agent_id"}
}

func (t *Ticket) AllowedSortFields() []string {
	return []string{"created_at"}
}
