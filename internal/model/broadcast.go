package model

import "time"

type (
	Broadcast struct {
		DocumentID      string    `json:"id" firestore:"-"` // uuid v7
		Name            string    `json:"name" firestore:"name"`
		TemplateID      string    `json:"template_id" firestore:"template_id"`
		TenantID        string    `json:"tenant_id" firestore:"tenant_id"`
		RecipientIDs    []string  `json:"recipient_ids" firestore:"recipient_ids"`
		PhoneNumberID   string    `json:"phone_number_id" firestore:"phone_number_id"`
		ParameterFormat *string   `json:"parameter_format" firestore:"parameter_format"`
		Payload         string    `json:"payload" firestore:"payload"` // raw json string of template
		Status          string    `json:"status" firestore:"status"`   // scheduled, sent, cancelled
		SendAt          time.Time `json:"send_at" firestore:"send_at"`
		CreatedBy       string    `json:"created_by" firestore:"created_by"`
		CreatedAt       time.Time `json:"created_at" firestore:"created_at"`
		UpdatedAt       time.Time `json:"updated_at" firestore:"updated_at"`
	}
)

func (b Broadcast) TableName() string {
	return "broadcasts"
}

func (b Broadcast) AllowedFilterFields() []string {
	return []string{"phone_number_id", "status", "tenant_id"}
}

func (b Broadcast) AllowedSortFields() []string {
	return []string{"created_at"}
}
