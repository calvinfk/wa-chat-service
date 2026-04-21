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

type MeiliMessageRepository struct {
	message model.Message
	db      meilisearch.ServiceManager
}

func NewMeiliMessageRepository(db meilisearch.ServiceManager) *MeiliMessageRepository {
	var message model.Message

	// Update filterable attributes only when different
	// Assume the fillterable uses a string slice
	desiredFilterable := message.AllowedFilterFields()
	currentFilterable, err := db.Index(message.TableName()).GetFilterableAttributes()
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
		if _, err := db.Index(message.TableName()).UpdateFilterableAttributes(&filterableAttributes); err != nil {
			log.Fatalf("failed to update filterable attributes: %v", err)
		}
	}

	// Update sortable attributes only when different
	desiredSortable := message.AllowedSortFields()
	currentSortable, err := db.Index(message.TableName()).GetSortableAttributes()
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
		if _, err := db.Index(message.TableName()).UpdateSortableAttributes(&sortableAttributes); err != nil {
			log.Fatalf("failed to update sortable attributes: %v", err)
		}
	}

	return &MeiliMessageRepository{
		message: message,
		db:      db,
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
