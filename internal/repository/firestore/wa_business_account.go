package repository_firestore

import (
	"context"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
)

type WhatsappBusinessAccountRepository struct {
	whatsappBusinessAccount model.WaBusinessAccount
	db                      *firestore.Client
}

func NewWhatsappBusinessAccountRepository(db *firestore.Client) *WhatsappBusinessAccountRepository {
	return &WhatsappBusinessAccountRepository{db: db}
}

func (r *WhatsappBusinessAccountRepository) GetByID(ctx context.Context, id string) (model.WaBusinessAccount, error) {
	doc, err := r.db.Collection(r.whatsappBusinessAccount.TableName()).Doc(id).Get(ctx)
	if err != nil {
		return model.WaBusinessAccount{}, err
	}
	var account model.WaBusinessAccount
	docData := doc.Data()
	docData["id"] = doc.Ref.ID
	err = utils.MapToStruct(docData, &account)
	if err != nil {
		return model.WaBusinessAccount{}, err
	}
	return account, nil
}
