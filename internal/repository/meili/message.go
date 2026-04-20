package repository_meili

import (
	"context"
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

func (r *MeiliMessageRepository) GetFiltered(ctx context.Context, filter filter_request.FilterRequest[dto.MessageGetByChatIDRequest]) ([]model.Message, int64, error) {
	var messages []model.Message
	searched, err := r.db.Index(r.message.TableName()).Search(filter.SpecificFilter.Search, &meilisearch.SearchRequest{
		Limit:  int64(filter.PageSize),
		Offset: int64((filter.Page - 1) * filter.PageSize),
		Filter: "chat_id = " + filter.SpecificFilter.ChatID,
		Sort:   []string{"created_at:desc"},
	})
	if err != nil {
		return messages, 0, err
	}
	for _, hit := range searched.Hits {
		var message model.Message
		err := hit.DecodeInto(&message)
		if err != nil {
			return messages, 0, err
		}
		messages = append(messages, message)
	}
	return messages, searched.TotalHits, err
}
