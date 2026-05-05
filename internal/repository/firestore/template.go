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

type TemplateRepository struct {
	WaBusinessAccount model.WaBusinessAccount
	template          model.Template
	db                *firestore.Client
}

func NewTemplateRepository(db *firestore.Client) *TemplateRepository {
	return &TemplateRepository{
		db: db,
	}
}

func (r *TemplateRepository) GetAll(ctx context.Context, whatsappBusinessAccountID string) ([]model.Template, error) {
	var templates []model.Template
	docs, err := r.db.
		Collection(r.WaBusinessAccount.TableName()).Doc(whatsappBusinessAccountID).
		Collection(r.template.TableName()).Documents(ctx).GetAll()
	if err != nil {
		return templates, err
	}
	for _, doc := range docs {
		var template model.Template
		docData := doc.Data()
		docData["id"] = doc.Ref.ID
		docData["whatsapp_business_account_id"] = whatsappBusinessAccountID
		err := utils.MapToStruct(docData, &template)
		if err != nil {
			return templates, err
		}
		templates = append(templates, template)
	}
	return templates, nil
}

func (r *TemplateRepository) GetByID(ctx context.Context, whatsappBusinessAccountID string, documentID string) (model.Template, error) {
	var template model.Template
	doc, err := r.db.
		Collection(r.WaBusinessAccount.TableName()).Doc(whatsappBusinessAccountID).
		Collection(r.template.TableName()).Doc(documentID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return template, errs.ErrGenericNotFound
		}
		return template, err
	}
	docData := doc.Data()
	docData["id"] = doc.Ref.ID
	docData["whatsapp_business_account_id"] = whatsappBusinessAccountID
	err = utils.MapToStruct(docData, &template)
	if err != nil {
		return template, err
	}
	return template, nil
}

func (r *TemplateRepository) Upsert(ctx context.Context, tx *firestore.Transaction, inputData model.Template) (model.Template, error) {
	created := false
	updates := []firestore.Update{
		{Path: "name", Value: inputData.Name},
		{Path: "category", Value: inputData.Category},
		{Path: "is_primary_device_delivery_only", Value: inputData.IsPrimaryDeviceDeliveryOnly},
		{Path: "language", Value: inputData.Language},
		{Path: "message_send_ttl_seconds", Value: inputData.MessageSendTTLSeconds},
		{Path: "parameter_format", Value: inputData.ParameterFormat},
		{Path: "status", Value: inputData.Status},
		{Path: "components", Value: inputData.Components},
		{Path: "updated_at", Value: inputData.UpdatedAt},
	}
	docRef := r.db.
		Collection(r.WaBusinessAccount.TableName()).Doc(inputData.WaBusinessAccountID).
		Collection(r.template.TableName()).Doc(inputData.DocumentID)
	_, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return r.template, err
		}
		created = true
	}
	if tx == nil {
		if created {
			_, err = docRef.Set(ctx, inputData)
			if err != nil {
				return r.template, err
			}
		} else {
			_, err = docRef.Update(ctx, updates)
			if err != nil {
				return r.template, err
			}
		}
	} else {
		if created {
			err = tx.Set(docRef, inputData)
		} else {
			err = tx.Update(docRef, updates)
		}
	}
	return inputData, err
}

func (r *TemplateRepository) DeleteByID(ctx context.Context, tx *firestore.Transaction, whatsappBusinessAccountID string, templateID string) error {
	docRef := r.db.
		Collection(r.WaBusinessAccount.TableName()).Doc(whatsappBusinessAccountID).
		Collection(r.template.TableName()).Doc(templateID)
	_, err := tx.Get(docRef)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return err
		}
		return errs.ErrGenericNotFound
	}

	if tx != nil {
		err = tx.Delete(docRef)
	} else {
		_, err = docRef.Delete(ctx)
	}
	return err
}
