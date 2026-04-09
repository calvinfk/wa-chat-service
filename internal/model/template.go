package model

import "time"

type Template struct {
	DocumentID                  string    `json:"__name__" firestore:"-"` // id from whatsapp
	Name                        string    `json:"name" firestore:"name"`
	Category                    string    `json:"category" firestore:"category"` // marketing, utility, authentication
	IsPrimaryDeviceDeliveryOnly bool      `json:"is_primary_device_delivery_only" firestore:"is_primary_device_delivery_only"`
	Language                    string    `json:"language" firestore:"language"`
	MessageSendTTLSeconds       int       `json:"message_send_ttl_seconds" firestore:"message_send_ttl_seconds"`
	ParameterFormat             *string   `json:"parameter_format" firestore:"parameter_format"`
	Status                      string    `json:"status" firestore:"status"`         // approved, rejected, etc
	Components                  string    `json:"components" firestore:"components"` // json string of components array from whatsapp
	CreatedAt                   time.Time `json:"created_at" firestore:"created_at"`
}

func (t Template) TableName() string {
	return "templates"
}

func (Template) AllowedFilterFields() []string {
	return []string{"name", "category", "language", "status"}
}
func (Template) AllowedSortFields() []string {
	return []string{"created_at", "updated_at"}
}
