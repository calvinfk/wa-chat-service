package repository_meili

import (
	"context"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/utils"

	"github.com/meilisearch/meilisearch-go"
	"go.uber.org/zap"
)

type MeiliTemplateRepository struct {
	template model.Template
	db       meilisearch.ServiceManager
	zsLog    *zap.SugaredLogger
}

func NewMeiliTemplateRepository(db meilisearch.ServiceManager, zsLog *zap.SugaredLogger) *MeiliTemplateRepository {
	var template model.Template
	// Ensure the index exists and has the correct settings
	index := db.Index(template.TableName())
	_, err := index.FetchInfo()
	if err != nil {
		if meiliErr, ok := err.(*meilisearch.Error); ok && meiliErr.MeilisearchApiError.Code == "index_not_found" {
			zsLog.Infof("[NewMeiliTemplateRepository] Index %s not found, creating...", template.TableName())
			taskInfo, err := db.CreateIndex(&meilisearch.IndexConfig{
				Uid:        template.TableName(),
				PrimaryKey: template.PKName(),
			})
			if err != nil {
				zsLog.Fatalf("[NewMeiliTemplateRepository] failed to create index: %v", err)
			}
			// Wait for the index creation task to complete
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for {
				task, err := db.GetTask(taskInfo.TaskUID)
				if err != nil {
					zsLog.Fatalf("[NewMeiliTemplateRepository] failed to get task status: %v", err)
				}
				if task.Status == "succeeded" {
					break
				} else if task.Status == "failed" {
					zsLog.Fatalf("[NewMeiliTemplateRepository] index creation task failed: %v", task.Error)
				}
				select {
				case <-ctx.Done():
					zsLog.Fatalf("[NewMeiliTemplateRepository] timed out waiting for index creation task to complete")
				case <-time.After(500 * time.Millisecond):
				}
			}
		} else {
			zsLog.Fatalf("[NewMeiliTemplateRepository] failed to fetch index info: %v", err)
		}
	}
	// Update filterable attributes only when different
	desiredFilterable := template.AllowedFilterFields()
	currentFilterable, err := db.Index(template.TableName()).GetFilterableAttributes()
	if err != nil {
		zsLog.Fatalf("[NewMeiliTemplateRepository] failed to get filterable attributes: %v", err)
	}
	currentFilterableSlice := []string{}
	if currentFilterable != nil {
		currentFilterableSlice, err = utils.AnySliceToStringSlice(*currentFilterable)
		if err != nil {
			zsLog.Fatalf("[NewMeiliTemplateRepository] failed to convert filterable attributes: %v", err)
		}
	}
	if currentFilterable == nil || !utils.SameStringSet(currentFilterableSlice, desiredFilterable) {
		filterableAttributes := make([]any, 0, len(desiredFilterable))
		for _, field := range desiredFilterable {
			filterableAttributes = append(filterableAttributes, field)
		}
		if _, err := db.Index(template.TableName()).UpdateFilterableAttributes(&filterableAttributes); err != nil {
			zsLog.Fatalf("[NewMeiliTemplateRepository] failed to update filterable attributes: %v", err)
		}
	}

	// Update sortable attributes only when different
	desiredSortable := template.AllowedSortFields()
	currentSortable, err := db.Index(template.TableName()).GetSortableAttributes()
	if err != nil {
		zsLog.Fatalf("[NewMeiliTemplateRepository] failed to get sortable attributes: %v", err)
	}
	currentSortableSlice := []string{}
	if currentSortable != nil {
		currentSortableSlice, err = utils.AnySliceToStringSlice(*currentSortable)
		if err != nil {
			zsLog.Fatalf("[NewMeiliTemplateRepository] failed to convert sortable attributes: %v", err)
		}
	}
	if currentSortable == nil || !utils.SameStringSet(currentSortableSlice, desiredSortable) {
		sortableAttributes := desiredSortable
		if _, err := db.Index(template.TableName()).UpdateSortableAttributes(&sortableAttributes); err != nil {
			zsLog.Fatalf("[NewMeiliTemplateRepository] failed to update sortable attributes: %v", err)
		}
	}

	return &MeiliTemplateRepository{
		db:    db,
		zsLog: zsLog,
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

func (r *MeiliTemplateRepository) GetFiltered(ctx context.Context, filterRequest filter_request.FilterRequest[dto.TemplateFilterRequest]) ([]model.Template, int64, filter_request.Paginate, error) {
	var templates []model.Template
	filters, sort, paginate, err := filter_request.InitializeFilter(filterRequest, r.template.AllowedFilterFields(), r.template.AllowedSortFields())
	if err != nil {
		return templates, 0, paginate, err
	}
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
		templates = append(templates, template)
	}
	return templates, searched.TotalHits, paginate, nil
}
