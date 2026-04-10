package model

import "time"

type Contact struct {
	DocumentID  string    `json:"__name__" firestore:"-"`  // uuid v7
	TenantID    string    `json:"tenant_id" firestore:"-"` // reference to tenant id from tenant collection
	PhoneNumber string    `json:"phone_number" firestore:"phone_number"`
	Name        string    `json:"name" firestore:"name"`
	CreatedAt   time.Time `json:"created_at" firestore:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" firestore:"updated_at"`
}

func (c Contact) TableName() string {
	return "contacts"
}

func (c Contact) AllowedFilterFields() []string {
	return []string{"phone_number", "name"}
}

func (c Contact) AllowedSortFields() []string {
	return []string{"created_at", "updated_at"}
}
