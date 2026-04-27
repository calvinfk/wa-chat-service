package model

import "time"

type WaPhone struct {
	DocumentID          string    `json:"id" firestore:"-"`                                          // generated uuid
	WaBusinessAccountID string    `json:"wa_business_account_id" firestore:"wa_business_account_id"` // foreign key to WhatsApp Business Account
	PhoneNumberId       string    `json:"phone_number_id" firestore:"phone_number_id"`
	PhoneNumber         string    `json:"phone_number" firestore:"phone_number"`
	DisplayName         string    `json:"display_name" firestore:"display_name"`
	CreatedAt           time.Time `json:"created_at" firestore:"created_at"`
}

func (WaPhone) TableName() string {
	return "wa_phones"
}
