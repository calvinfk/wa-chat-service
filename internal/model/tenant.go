package model

import (
	"time"
)

type Tenant struct {
	DocumentID    string    `json:"id" firestore:"-"`
	WabaID        string    `json:"waba_id" firestore:"waba_id"`
	AccessToken   string    `json:"access_token" firestore:"access_token"` // encrypted
	PhoneNumberID string    `json:"phone_number_id" firestore:"phone_number_id"`
	Name          string    `json:"name" firestore:"name"`
	CreatedAt     time.Time `json:"created_at" firestore:"created_at"`
}

func (t *Tenant) TableName() string {
	return "tenants"
}
