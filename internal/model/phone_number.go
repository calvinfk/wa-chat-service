package model

import "github.com/google/uuid"

type PhoneNumber struct {
	DocumentID    uuid.UUID `firestore:"-"`
	WabaID        string    `firestore:"waba_id"`
	AccessToken   string    `firestore:"access_token"` // encrypted
	PhoneNumberID string    `firestore:"phone_number_id"`
	Name          string    `firestore:"name"`
	CreatedAt     int64     `firestore:"created_at"`
}
