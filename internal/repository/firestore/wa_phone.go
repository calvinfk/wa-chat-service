package repository_firestore

import (
	"context"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type WaPhoneRepository struct {
	waPhone model.WaPhone
	db      *firestore.Client
}

func NewWaPhoneRepository(db *firestore.Client) *WaPhoneRepository {
	return &WaPhoneRepository{db: db}
}

func (r *WaPhoneRepository) GetByPhoneNumberId(ctx context.Context, phoneNumberId string) (model.WaPhone, error) {
	doc, err := r.db.Collection(r.waPhone.TableName()).Where("phone_number_id", "==", phoneNumberId).Limit(1).Documents(ctx).Next()
	if err != nil {
		if err == iterator.Done {
			return model.WaPhone{}, errs.ErrGenericNotFound
		}
		return model.WaPhone{}, err
	}
	var phone model.WaPhone
	docData := doc.Data()
	docData["id"] = doc.Ref.ID
	err = utils.MapToStruct(docData, &phone)
	if err != nil {
		return model.WaPhone{}, err
	}
	return phone, nil
}

func (r *WaPhoneRepository) GetByWaBusinessAccountID(ctx context.Context, waBusinessAccountID string) ([]model.WaPhone, error) {
	var phones []model.WaPhone
	docs, err := r.db.Collection(r.waPhone.TableName()).Where("wa_business_account_id", "==", waBusinessAccountID).Documents(ctx).GetAll()
	if err != nil {
		return phones, err
	}
	for _, doc := range docs {
		var phone model.WaPhone
		docData := doc.Data()
		docData["id"] = doc.Ref.ID
		err = utils.MapToStruct(docData, &phone)
		if err != nil {
			return nil, err
		}
		phones = append(phones, phone)
	}
	return phones, nil
}
