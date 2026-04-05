package repository_firestore

import (
	"context"
	"wa_chat_service/internal/model"

	"cloud.google.com/go/firestore"
)

type PhoneNumberRepository struct {
	db *firestore.Client
}

func NewPhoneNumberRepository(db *firestore.Client) *PhoneNumberRepository {
	return &PhoneNumberRepository{db: db}
}
func (r *PhoneNumberRepository) GetByPhoneNumberID(ctx context.Context, phoneNumberID string) (model.PhoneNumber, error) {
	doc, err := r.db.Collection("phone_numbers").Where("phone_number_id", "==", phoneNumberID).Limit(1).Documents(ctx).Next()
	if err != nil {
		return model.PhoneNumber{}, err
	}
	// Unmarshal the document into a PhoneNumber model
	var phoneNumber model.PhoneNumber
	if err := doc.DataTo(&phoneNumber); err != nil {
		return model.PhoneNumber{}, err
	}
	return phoneNumber, nil
}
