package repository_meili

import (
	"context"
	"log"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/utils"

	"github.com/meilisearch/meilisearch-go"
)

type MeiliTemplateRepository struct {
	template model.Template
	db       meilisearch.ServiceManager
}

func NewMeiliTemplateRepository(db meilisearch.ServiceManager) *MeiliTemplateRepository {
	var template model.Template

	// Update filterable attributes only when different
	desiredFilterable := template.AllowedFilterFields()
	currentFilterable, err := db.Index(template.TableName()).GetFilterableAttributes()
	if err != nil {
		log.Fatalf("failed to get filterable attributes: %v", err)
	}
	currentFilterableSlice := []string{}
	if currentFilterable != nil {
		currentFilterableSlice, err = utils.AnySliceToStringSlice(*currentFilterable)
		if err != nil {
			log.Fatalf("failed to convert filterable attributes: %v", err)
		}
	}
	if currentFilterable == nil || !utils.SameStringSet(currentFilterableSlice, desiredFilterable) {
		filterableAttributes := make([]any, 0, len(desiredFilterable))
		for _, field := range desiredFilterable {
			filterableAttributes = append(filterableAttributes, field)
		}
		if _, err := db.Index(template.TableName()).UpdateFilterableAttributes(&filterableAttributes); err != nil {
			log.Fatalf("failed to update filterable attributes: %v", err)
		}
	}

	// Update sortable attributes only when different
	desiredSortable := template.AllowedSortFields()
	currentSortable, err := db.Index(template.TableName()).GetSortableAttributes()
	if err != nil {
		log.Fatalf("failed to get sortable attributes: %v", err)
	}
	currentSortableSlice := []string{}
	if currentSortable != nil {
		currentSortableSlice, err = utils.AnySliceToStringSlice(*currentSortable)
		if err != nil {
			log.Fatalf("failed to convert sortable attributes: %v", err)
		}
	}
	if currentSortable == nil || !utils.SameStringSet(currentSortableSlice, desiredSortable) {
		sortableAttributes := desiredSortable
		if _, err := db.Index(template.TableName()).UpdateSortableAttributes(&sortableAttributes); err != nil {
			log.Fatalf("failed to update sortable attributes: %v", err)
		}
	}

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
	searchRequest, err := filter_request.ApplyFilterMeili(filters, sort, paginate)
	if err != nil {
		return templates, 0, paginate, err
	}
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

// updated by assistant
