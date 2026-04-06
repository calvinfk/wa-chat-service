package model

import "github.com/google/uuid"

type PhoneNumber struct {
	DocumentID    uuid.UUID `json:"__name__" firestore:"-"`
	WabaID        string    `json:"waba_id" firestore:"waba_id"`
	AccessToken   string    `json:"access_token" firestore:"access_token"` // encrypted
	PhoneNumberID string    `json:"phone_number_id" firestore:"phone_number_id"`
	Name          string    `json:"name" firestore:"name"`
	CreatedAt     int64     `json:"created_at" firestore:"created_at"`
}
