package repository_firestore

import (
	"context"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/formatter"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TemplateRepository struct {
	template model.Template
	db       *firestore.Client
}

func NewTemplateRepository(db *firestore.Client) *TemplateRepository {
	return &TemplateRepository{
		db: db,
	}
}

func (r *TemplateRepository) GetFilteredByPhoneNumberID(ctx context.Context, tenantID string, inputData filter_request.FilterRequest[dto.TemplateGetByPhoneNumberID]) (filter_request.FilterResponse[dto.TemplateGetByPhoneNumberIDResponse], error) {
	var response filter_request.FilterResponse[dto.TemplateGetByPhoneNumberIDResponse]
	filters, sort, paginate, err := filter_request.InitializeFilter(inputData, r.template.AllowedFilterFields(), r.template.AllowedSortFields())
	if err != nil {
		return response, err
	}
	collection := r.db.Collection("tenants").Doc(tenantID).Collection(r.template.TableName())
	query := collection.Query
	docs, totalData, err := filter_request.ApplyFilterFirestore(ctx, query, filters, paginate, sort)
	if err != nil {
		return response, err
	}
	var result []dto.TemplateGetByPhoneNumberIDResponse
	for _, doc := range docs {
		var template model.Template
		docData := doc.Data()
		docData[firestore.DocumentID] = doc.Ref.ID
		err := formatter.MapToStruct(docData, &template)
		if err != nil {
			return response, err
		}
		result = append(result, dto.TemplateGetByPhoneNumberIDResponse{}.FromModel(template))
	}
	response = filter_request.NewFilterResponse(result, paginate, totalData)
	return response, nil
}

func (r *TemplateRepository) GetByTenantID(ctx context.Context, tenantID string) ([]model.Template, error) {
	var templates []model.Template
	collection := r.db.Collection("tenants").Doc(tenantID).Collection(r.template.TableName())
	query := collection.Query
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return templates, err
	}
	for _, doc := range docs {
		var template model.Template
		docData := doc.Data()
		docData[firestore.DocumentID] = doc.Ref.ID
		err := formatter.MapToStruct(docData, &template)
		if err != nil {
			return templates, err
		}
		templates = append(templates, template)
	}
	return templates, nil
}

func (r *TemplateRepository) Upsert(ctx context.Context, tx *firestore.Transaction, tenantID string, inputData model.Template) (model.Template, error) {
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		doc := r.db.
			Collection("tenants").
			Doc(tenantID).
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

func (r *TemplateRepository) DeleteByID(ctx context.Context, tx *firestore.Transaction, tenantID string, templateID string) error {
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		doc := r.db.Collection("tenants").
			Doc(tenantID).
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

func (r *TemplateRepository) DeleteByName(ctx context.Context, tx *firestore.Transaction, tenantID string, name string) error {
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		collection := r.db.Collection("tenants").Doc(tenantID).Collection(r.template.TableName())
		query := collection.Where("name", "==", name)
		docs, err := query.Documents(ctx).GetAll()
		if err != nil {
			return err
		}
		for _, doc := range docs {
			err := tx.Delete(doc.Ref)
			if err != nil {
				return err
			}
		}
		return nil
	}
	var err error
	if tx == nil {
		err = r.db.RunTransaction(ctx, execDB)
	} else {
		err = execDB(ctx, tx)
	}
	return err
}
