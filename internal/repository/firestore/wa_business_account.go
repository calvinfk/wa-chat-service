package repository_firestore

import (
	"context"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		if status.Code(err) == codes.NotFound {
			return model.WaBusinessAccount{}, errs.ErrGenericNotFound
		}
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

func (r *WhatsappBusinessAccountRepository) GetByTenantID(ctx context.Context, tenantID string) ([]model.WaBusinessAccount, error) {
	var accounts []model.WaBusinessAccount
	docs, err := r.db.Collection(r.whatsappBusinessAccount.TableName()).Where("tenant_id", "==", tenantID).Documents(ctx).GetAll()
	if err != nil {
		return accounts, err
	}
	for _, doc := range docs {
		var account model.WaBusinessAccount
		docData := doc.Data()
		docData["id"] = doc.Ref.ID
		err = utils.MapToStruct(docData, &account)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}
