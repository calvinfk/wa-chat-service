package repository_meili

import (
	"context"
	"log"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/filter_request"

	"github.com/meilisearch/meilisearch-go"
)

type MeiliMessageRepository struct {
	message model.Message
	db      meilisearch.ServiceManager
}

func NewMeiliMessageRepository(db meilisearch.ServiceManager) *MeiliMessageRepository {
	var message model.Message
	var filterableAttributes []any
	for _, field := range message.AllowedFilterFields() {
		filterableAttributes = append(filterableAttributes, field)
	}
	// db.Index(message.TableName()).GetFilterableAttributes()
	_, err := db.Index(message.TableName()).UpdateFilterableAttributes(&filterableAttributes)
	if err != nil {
		log.Fatalf("failed to update filterable attributes: %v", err)
	}
	sortableAttributes := message.AllowedSortFields()
	_, err = db.Index(message.TableName()).UpdateSortableAttributes(&sortableAttributes)
	if err != nil {
		log.Fatalf("failed to update sortable attributes: %v", err)
	}
	return &MeiliMessageRepository{
		db: db,
	}
}

func (r *MeiliMessageRepository) AddDocuments(ctx context.Context, messages []model.Message) error {
	primaryKey := r.message.PKName()
	_, err := r.db.Index("messages").AddDocuments(messages, &meilisearch.DocumentOptions{
		PrimaryKey: &primaryKey,
	})
	return err
}

func (r *MeiliMessageRepository) GetFiltered(ctx context.Context, filterRequest filter_request.FilterRequest[dto.MessageGetByChatIDRequest]) ([]model.Message, int64, filter_request.Paginate, error) {
	var messages []model.Message
	filters, sort, paginate, err := filter_request.InitializeFilter(filterRequest, r.message.AllowedFilterFields(), r.message.AllowedSortFields())
	if err != nil {
		return messages, 0, paginate, err
	}
	filters = append(filters, filter_request.Filter{
		Field:    "chat_id",
		Operator: filter_request.OpEq,
		Value:    filterRequest.SpecificFilter.ChatID,
	})
	searchRequest, err := filter_request.ApplyFilterMeili(filters, sort, paginate)
	if err != nil {
		return messages, 0, paginate, err
	}
	searched, err := r.db.Index(r.message.TableName()).Search(filterRequest.SpecificFilter.Search, searchRequest)
	if err != nil {
		return messages, 0, paginate, err
	}
	for _, hit := range searched.Hits {
		var message model.Message
		err := hit.DecodeInto(&message)
		if err != nil {
			return messages, 0, paginate, err
		}
		messages = append(messages, message)
	}
	return messages, searched.TotalHits, paginate, err
}
