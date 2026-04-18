package model

import "time"

type Template struct {
	DocumentID                  string    `json:"id" firestore:"-"` // id from whatsapp
	TenantID                    string    `json:"tenant_id" firestore:"-"`
	Name                        string    `json:"name" firestore:"name"`
	Category                    string    `json:"category" firestore:"category"` // marketing, utility, authentication
	IsPrimaryDeviceDeliveryOnly bool      `json:"is_primary_device_delivery_only" firestore:"is_primary_device_delivery_only"`
	Language                    string    `json:"language" firestore:"language"`
	MessageSendTTLSeconds       int       `json:"message_send_ttl_seconds" firestore:"message_send_ttl_seconds"`
	ParameterFormat             *string   `json:"parameter_format" firestore:"parameter_format"`
	Status                      string    `json:"status" firestore:"status"`         // approved, rejected, etc
	Components                  string    `json:"components" firestore:"components"` // json string of components array from whatsapp
	CreatedAt                   time.Time `json:"created_at" firestore:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at" firestore:"updated_at"`
}

func (Template) TableName() string {
	return "templates"
}

func (Template) PrimaryKey() string {
	return "id"
}

func (Template) AllowedFilterFields() []string {
	return []string{"tenant_id", "name", "category", "status"}
}
func (Template) AllowedSortFields() []string {
	return []string{"created_at"}
}
