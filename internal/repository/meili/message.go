package repository_meili

import (
	"context"
	"fmt"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/utils"

	"github.com/meilisearch/meilisearch-go"
	"go.uber.org/zap"
)

type MeiliMessageRepository struct {
	message model.Message
	db      meilisearch.ServiceManager
}

func NewMeiliMessageRepository(db meilisearch.ServiceManager, zsLog *zap.SugaredLogger) *MeiliMessageRepository {
	var message model.Message
	index := db.Index(message.TableName())
	_, err := index.FetchInfo()
	if err != nil {
		if meiliErr, ok := err.(*meilisearch.Error); ok && meiliErr.MeilisearchApiError.Code == "index_not_found" {
			zsLog.Infof("[NewMeiliMessageRepository] Index %s not found, creating...", message.TableName())
			taskInfo, err := db.CreateIndex(&meilisearch.IndexConfig{
				Uid:        message.TableName(),
				PrimaryKey: message.PKName(),
			})
			if err != nil {
				zsLog.Fatalf("[NewMeiliMessageRepository] failed to create index: %v", err)
			}
			// Wait for the index creation task to complete
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for {
				task, err := db.GetTask(taskInfo.TaskUID)
				if err != nil {
					zsLog.Fatalf("[NewMeiliMessageRepository] failed to get task status: %v", err)
				}
				if task.Status == "succeeded" {
					break
				}
				if task.Status == "failed" {
					zsLog.Fatalf("[NewMeiliMessageRepository] index creation task failed: %v", task.Error)
				}
				select {
				case <-ctx.Done():
					zsLog.Fatalf("[NewMeiliMessageRepository] timed out waiting for index creation task to complete")
				case <-time.After(500 * time.Millisecond):
				}
			}
		}
		zsLog.Fatalf("[NewMeiliMessageRepository] failed to fetch index info: %v", err)
	}

	// Update filterable attributes only when different
	// Assume the fillterable uses a string slice
	desiredFilterable := message.AllowedFilterFields()
	currentFilterable, err := db.Index(message.TableName()).GetFilterableAttributes()
	if err != nil {
		zsLog.Fatalf("[NewMeiliMessageRepository] failed to get filterable attributes: %v", err)
	}
	currentFilterableSlice := []string{}
	if currentFilterable != nil {
		currentFilterableSlice, err = utils.AnySliceToStringSlice(*currentFilterable)
		if err != nil {
			zsLog.Fatalf("[NewMeiliMessageRepository] failed to convert filterable attributes: %v", err)
		}
	}
	if currentFilterable == nil || !utils.SameStringSet(currentFilterableSlice, desiredFilterable) {
		filterableAttributes := make([]any, 0, len(desiredFilterable))
		for _, field := range desiredFilterable {
			filterableAttributes = append(filterableAttributes, field)
		}
		if _, err := db.Index(message.TableName()).UpdateFilterableAttributes(&filterableAttributes); err != nil {
			zsLog.Fatalf("[NewMeiliMessageRepository] failed to update filterable attributes: %v", err)
		}
	}

	// Update sortable attributes only when different
	desiredSortable := message.AllowedSortFields()
	currentSortable, err := db.Index(message.TableName()).GetSortableAttributes()
	if err != nil {
		zsLog.Fatalf("[NewMeiliMessageRepository] failed to get sortable attributes: %v", err)
	}
	currentSortableSlice := []string{}
	if currentSortable != nil {
		currentSortableSlice, err = utils.AnySliceToStringSlice(*currentSortable)
		if err != nil {
			zsLog.Fatalf("[NewMeiliMessageRepository] failed to convert sortable attributes: %v", err)
		}
	}
	if currentSortable == nil || !utils.SameStringSet(currentSortableSlice, desiredSortable) {
		sortableAttributes := desiredSortable
		if _, err := db.Index(message.TableName()).UpdateSortableAttributes(&sortableAttributes); err != nil {
			zsLog.Fatalf("[NewMeiliMessageRepository] failed to update sortable attributes: %v", err)
		}
	}

	return &MeiliMessageRepository{
		message: message,
		db:      db,
	}
}

func (r *MeiliMessageRepository) AddDocuments(ctx context.Context, messages []model.Message) (*meilisearch.TaskInfo, error) {
	primaryKey := r.message.PKName()
	taskInfo, err := r.db.Index("messages").AddDocuments(messages, &meilisearch.DocumentOptions{
		PrimaryKey: &primaryKey,
	})
	return taskInfo, err
}

func (r *MeiliMessageRepository) AddDocumentsSync(ctx context.Context, messages []model.Message) error {
	taskInfo, err := r.AddDocuments(ctx, messages)
	if err != nil {
		return err
	}
	// Wait for the indexing task to complete
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			task, err := r.db.GetTask(taskInfo.TaskUID)
			if err != nil {
				return err
			}
			if task.Status != meilisearch.TaskStatusProcessing {
				if task.Status == meilisearch.TaskStatusFailed {
					return fmt.Errorf("search index task failed: %v", task.Error)
				}
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
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
