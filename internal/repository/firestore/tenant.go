package repository_firestore

import (
	"context"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
)

type TenantRepository struct {
	tenant model.Tenant
	db     *firestore.Client
}

func NewTenantRepository(db *firestore.Client) *TenantRepository {
	return &TenantRepository{db: db}
}
func (r *TenantRepository) GetByPhoneNumberID(ctx context.Context, phoneNumberID string) (model.Tenant, error) {
	doc, err := r.db.Collection(r.tenant.TableName()).Where("phone_number_id", "==", phoneNumberID).Limit(1).Documents(ctx).Next()
	if err != nil {
		return model.Tenant{}, err
	}
	// Unmarshal the document into a Tenant model
	var tenant model.Tenant
	docData := doc.Data()
	docData[firestore.DocumentID] = doc.Ref.ID
	err = utils.MapToStruct(docData, &tenant)
	if err != nil {
		return model.Tenant{}, err
	}
	return tenant, nil
}
