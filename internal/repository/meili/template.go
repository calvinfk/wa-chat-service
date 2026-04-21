package repository_meili

import (
	"context"
	"log"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/filter_request"

	"github.com/meilisearch/meilisearch-go"
)

type MeiliTemplateRepository struct {
	template model.Template
	db       meilisearch.ServiceManager
}

func NewMeiliTemplateRepository(db meilisearch.ServiceManager) *MeiliTemplateRepository {
	var template model.Template
	var filterableAttributes []any
	for _, field := range template.AllowedFilterFields() {
		filterableAttributes = append(filterableAttributes, field)
	}
	index := db.Index(template.TableName())
	_, err := index.UpdateFilterableAttributes(&filterableAttributes)
	if err != nil {
		panic(err)
	}
	sortableAttributes := template.AllowedSortFields()
	_, err = index.UpdateSortableAttributes(&sortableAttributes)
	if err != nil {
		panic(err)
	}
	settings, err := index.GetFilterableAttributes()
	if err != nil {
		log.Fatalf("failed to get settings: %v", err)
	}
	log.Printf("filterable attributes: %+v", settings)
	return &MeiliTemplateRepository{
		db: db,
	}
}

func (r *MeiliTemplateRepository) AddDocuments(ctx context.Context, documents []model.Template) error {
	primaryKey := r.template.PKName()
	_, err := r.db.Index(r.template.TableName()).AddDocuments(documents, &meilisearch.DocumentOptions{
		PrimaryKey: &primaryKey,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *MeiliTemplateRepository) DeleteDocuments(ctx context.Context, documentIDs []string) error {
	primaryKey := r.template.PKName()
	_, err := r.db.Index(r.template.TableName()).DeleteDocuments(documentIDs, &meilisearch.DocumentOptions{
		PrimaryKey: &primaryKey,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *MeiliTemplateRepository) GetFiltered(ctx context.Context, filterRequest filter_request.FilterRequest[dto.TemplateGetByTenantID]) ([]model.Template, int64, filter_request.Paginate, error) {
	var templates []model.Template
	filters, sort, paginate, err := filter_request.InitializeFilter(filterRequest, r.template.AllowedFilterFields(), r.template.AllowedSortFields())
	if err != nil {
		return templates, 0, paginate, err
	}
	filters = append(filters, filter_request.Filter{
		Field:    "tenant_id",
		Operator: filter_request.OpEq,
		Value:    filterRequest.SpecificFilter.TenantID,
	})
	searchRequest := filter_request.ApplyFilterMeili(filters, sort, paginate)
	searched, err := r.db.Index(r.template.TableName()).Search(filterRequest.SpecificFilter.Search, searchRequest)
	if err != nil {
		return templates, 0, paginate, err
	}
	for _, hit := range searched.Hits {
		var template model.Template
		err := hit.DecodeInto(&template)
		if err != nil {
			return templates, 0, paginate, err
		}
		if template.TenantID == filterRequest.SpecificFilter.TenantID {
			templates = append(templates, template)
		}
	}
	return templates, searched.TotalHits, paginate, nil
}
