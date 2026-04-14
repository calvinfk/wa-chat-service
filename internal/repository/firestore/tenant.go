package repository_firestore

import (
	"context"
	"fmt"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
)

type TenantRepository struct {
	tenant        model.Tenant
	contact       model.Contact
	templateField model.TemplateField
	db            *firestore.Client
}

func NewTenantRepository(db *firestore.Client) *TenantRepository {
	return &TenantRepository{db: db}
}

func (r *TenantRepository) GetByID(ctx context.Context, tenantID string) (model.Tenant, error) {
	doc, err := r.db.Collection(r.tenant.TableName()).Doc(tenantID).Get(ctx)
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

func (r *TenantRepository) InsertContact(ctx context.Context, contact model.Contact) error {
	_, err := r.db.
		Collection(r.tenant.TableName()).Doc(contact.TenantID).
		Collection(r.contact.TableName()).Doc(contact.DocumentID).
		Set(ctx, contact)
	return err
}

func (r *TenantRepository) GetContactsFiltered(ctx context.Context, tenantID string, filterRequest filter_request.FilterRequest[dto.ContactGetFilteredRequest]) (filter_request.FilterResponse[dto.ContactResponse], error) {
	filters, sort, paginate, err := filter_request.InitializeFilter(filterRequest, r.contact.AllowedFilterFields(), r.contact.AllowedSortFields())
	if err != nil {
		return filter_request.FilterResponse[dto.ContactResponse]{}, err
	}
	query := r.db.Collection(r.tenant.TableName()).Doc(tenantID).Collection(r.contact.TableName()).Query
	docs, totalData, err := filter_request.ApplyFilterFirestore(ctx, query, filters, sort, paginate)
	if err != nil {
		return filter_request.FilterResponse[dto.ContactResponse]{}, err
	}
	var result []dto.ContactResponse
	for _, doc := range docs {
		var contact model.Contact
		docData := doc.Data()
		docData[firestore.DocumentID] = doc.Ref.ID
		docData["tenant_id"] = tenantID
		err = utils.MapToStruct(docData, &contact)
		if err != nil {
			return filter_request.FilterResponse[dto.ContactResponse]{}, err
		}
		result = append(result, dto.ContactResponse{}.FromModel(contact))
	}
	response := filter_request.NewFilterResponse(result, paginate, totalData)
	return response, nil
}

func (r *TenantRepository) GetContactByPhoneNumbers(ctx context.Context, tenantID string, phoneNumbers []string) (map[string]map[string]string, error) {
	docs, err := r.db.Collection(r.tenant.TableName()).Doc(tenantID).Collection(r.contact.TableName()).Where("phone_number", "in", phoneNumbers).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	// Unmarshal the documents into Contact models
	contacts := make(map[string]map[string]string)
	for _, doc := range docs {
		docData := doc.Data()
		docData[firestore.DocumentID] = doc.Ref.ID
		docData["tenant_id"] = tenantID
		docDataCopy := make(map[string]string)
		for key, value := range docData {
			docDataCopy[key] = fmt.Sprintf("%v", value)
		}
		contacts[docDataCopy["phone_number"]] = docDataCopy
	}
	return contacts, nil
}

func (r *TenantRepository) GetContactByID(ctx context.Context, tenantID string, contactID string) (model.Contact, error) {
	docRef, err := r.db.Collection(r.tenant.TableName()).Doc(tenantID).Collection(r.contact.TableName()).Doc(contactID).Get(ctx)
	if err != nil {
		return model.Contact{}, err
	}
	var contact model.Contact
	docData := docRef.Data()
	docData[firestore.DocumentID] = docRef.Ref.ID
	docData["tenant_id"] = tenantID
	err = utils.MapToStruct(docData, &contact)
	if err != nil {
		return model.Contact{}, err
	}
	return contact, nil
}

func (r *TenantRepository) UpdateContact(ctx context.Context, contact model.Contact) error {
	_, err := r.db.
		Collection(r.tenant.TableName()).Doc(contact.TenantID).
		Collection(r.contact.TableName()).Doc(contact.DocumentID).
		Set(ctx, contact)
	return err
}

func (r *TenantRepository) GetTemplateFields(ctx context.Context, tenantID string) (map[string]model.TemplateField, error) {
	docs, err := r.db.Collection(r.tenant.TableName()).Doc(tenantID).Collection(r.templateField.TableName()).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	templateFields := make(map[string]model.TemplateField)
	for _, doc := range docs {
		var templateField model.TemplateField
		docData := doc.Data()
		docData[firestore.DocumentID] = doc.Ref.ID
		docData["tenant_id"] = tenantID
		err = utils.MapToStruct(docData, &templateField)
		if err != nil {
			return nil, err
		}
		templateFields[templateField.Key] = templateField
	}
	return templateFields, nil
}
