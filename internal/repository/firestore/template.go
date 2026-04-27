package repository_firestore

import (
	"context"
	"wa_chat_service/internal/model"
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

// func (r *TemplateRepository) GetFilteredByWhatsappBusinessAccountID(ctx context.Context, whatsappBusinessAccountID string, inputData filter_request.FilterRequest[dto.TemplateFilterRequest]) (filter_request.FilterResponse[dto.TemplateResponse], error) {
// 	var response filter_request.FilterResponse[dto.TemplateResponse]
// 	filters, sort, paginate, err := filter_request.InitializeFilter(inputData, r.template.AllowedFilterFields(), r.template.AllowedSortFields())
// 	if err != nil {
// 		return response, err
// 	}
// 	collection := r.db.Collection(r.WaBusinessAccount.TableName()).Doc(whatsappBusinessAccountID).Collection(r.template.TableName())
// 	query := collection.Query
// 	docs, totalData, err := filter_request.ApplyFilterFirestore(ctx, query, filters, sort, paginate)
// 	if err != nil {
// 		return response, err
// 	}
// 	var result []dto.TemplateResponse
// 	for _, doc := range docs {
// 		var template model.Template
// 		docData := doc.Data()
// 		docData["id"] = doc.Ref.ID
// 		docData["whatsapp_business_account_id"] = whatsappBusinessAccountID
// 		err := utils.MapToStruct(docData, &template)
// 		if err != nil {
// 			return response, err
// 		}
// 		result = append(result, dto.TemplateResponse{}.FromModel(template))
// 	}
// 	response = filter_request.NewFilterResponse(result, paginate, totalData)
// 	return response, nil
// }

func (r *TemplateRepository) GetAll(ctx context.Context, whatsappBusinessAccountID string) ([]model.Template, error) {
	var templates []model.Template
	collection := r.db.Collection(r.WaBusinessAccount.TableName()).Doc(whatsappBusinessAccountID).Collection(r.template.TableName())
	query := collection.Query
	docs, err := query.Documents(ctx).GetAll()
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
	docRef := r.db.
		Collection(r.WaBusinessAccount.TableName()).
		Doc(whatsappBusinessAccountID).
		Collection(r.template.TableName()).
		Doc(documentID)
	doc, err := docRef.Get(ctx)
	if err != nil {
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
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		doc := r.db.
			Collection(r.WaBusinessAccount.TableName()).
			Doc(inputData.WaBusinessAccountID).
			Collection(r.template.TableName()).
			Doc(inputData.DocumentID)
		_, getErr := tx.Get(doc)
		if getErr != nil {
			if status.Code(getErr) != codes.NotFound {
				return getErr
			}

			setErr := tx.Set(doc, inputData)
			if setErr != nil {
				return setErr
			}
			return nil
		}
		updateErr := tx.Update(doc, []firestore.Update{
			{Path: "name", Value: inputData.Name},
			{Path: "category", Value: inputData.Category},
			{Path: "is_primary_device_delivery_only", Value: inputData.IsPrimaryDeviceDeliveryOnly},
			{Path: "language", Value: inputData.Language},
			{Path: "message_send_ttl_seconds", Value: inputData.MessageSendTTLSeconds},
			{Path: "parameter_format", Value: inputData.ParameterFormat},
			{Path: "status", Value: inputData.Status},
			{Path: "components", Value: inputData.Components},
			{Path: "updated_at", Value: inputData.UpdatedAt},
		})
		if updateErr != nil {
			return updateErr
		}

		return nil
	}

	var err error
	if tx == nil {
		err = r.db.RunTransaction(ctx, execDB)
	} else {
		err = execDB(ctx, tx)
	}
	return inputData, err
}

func (r *TemplateRepository) DeleteByID(ctx context.Context, tx *firestore.Transaction, whatsappBusinessAccountID string, templateID string) error {
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		doc := r.db.Collection(r.WaBusinessAccount.TableName()).
			Doc(whatsappBusinessAccountID).
			Collection(r.template.TableName()).
			Doc(templateID)
		_, err := tx.Get(doc)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return nil
			}
			return err
		}
		return tx.Delete(doc)
	}
	var err error
	if tx == nil {
		err = r.db.RunTransaction(ctx, execDB)
	} else {
		err = execDB(ctx, tx)
	}
	return err
}
