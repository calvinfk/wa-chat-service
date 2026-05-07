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

type MeiliTicketMessageRepository struct {
	ticketMessage model.TicketMessage
	db            meilisearch.ServiceManager
	zsLog         *zap.SugaredLogger
}

func NewMeiliTicketMessageRepository(db meilisearch.ServiceManager, zsLog *zap.SugaredLogger) *MeiliTicketMessageRepository {
	var ticketMessage model.TicketMessage
	// Ensure the index exists and has the correct settings
	index := db.Index(ticketMessage.TableName())
	_, err := index.FetchInfo()
	if err != nil {
		if meiliErr, ok := err.(*meilisearch.Error); ok && meiliErr.MeilisearchApiError.Code == "index_not_found" {
			zsLog.Infof("[NewMeiliTicketMessageRepository] Index %s not found, creating...", ticketMessage.TableName())
			taskInfo, err := db.CreateIndex(&meilisearch.IndexConfig{
				Uid:        ticketMessage.TableName(),
				PrimaryKey: ticketMessage.PKName(),
			})
			if err != nil {
				zsLog.Fatalf("[NewMeiliTicketMessageRepository] failed to create index: %v", err)
			}
			// Wait for the index creation task to complete
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for {
				task, err := db.GetTask(taskInfo.TaskUID)
				if err != nil {
					zsLog.Fatalf("[NewMeiliTicketMessageRepository] failed to get task status: %v", err)
				}
				if task.Status == "succeeded" {
					break
				}
				if task.Status == "failed" {
					zsLog.Fatalf("[NewMeiliTicketMessageRepository] index creation task failed: %v", task.Error)
				}
				select {
				case <-ctx.Done():
					zsLog.Fatalf("[NewMeiliTicketMessageRepository] timed out waiting for index creation task to complete")
				case <-time.After(500 * time.Millisecond):
				}
			}
		} else {
			zsLog.Fatalf("[NewMeiliTicketMessageRepository] failed to fetch index info: %v", err)
		}
	}

	// Update filterable attributes only when different
	// Assume the fillterable uses a string slice
	desiredFilterable := ticketMessage.AllowedFilterFields()
	currentFilterable, err := db.Index(ticketMessage.TableName()).GetFilterableAttributes()
	if err != nil {
		zsLog.Fatalf("[NewMeiliTicketMessageRepository] failed to get filterable attributes: %v", err)
	}
	currentFilterableSlice := []string{}
	if currentFilterable != nil {
		currentFilterableSlice, err = utils.AnySliceToStringSlice(*currentFilterable)
		if err != nil {
			zsLog.Fatalf("[NewMeiliTicketMessageRepository] failed to convert filterable attributes: %v", err)
		}
	}
	if currentFilterable == nil || !utils.SameStringSet(currentFilterableSlice, desiredFilterable) {
		filterableAttributes := make([]any, 0, len(desiredFilterable))
		for _, field := range desiredFilterable {
			filterableAttributes = append(filterableAttributes, field)
		}
		if _, err := db.Index(ticketMessage.TableName()).UpdateFilterableAttributes(&filterableAttributes); err != nil {
			zsLog.Fatalf("[NewMeiliTicketMessageRepository] failed to update filterable attributes: %v", err)
		}
	}

	// Update sortable attributes only when different
	desiredSortable := ticketMessage.AllowedSortFields()
	currentSortable, err := db.Index(ticketMessage.TableName()).GetSortableAttributes()
	if err != nil {
		zsLog.Fatalf("[NewMeiliTicketMessageRepository] failed to get sortable attributes: %v", err)
	}
	currentSortableSlice := []string{}
	if currentSortable != nil {
		currentSortableSlice, err = utils.AnySliceToStringSlice(*currentSortable)
		if err != nil {
			zsLog.Fatalf("[NewMeiliTicketMessageRepository] failed to convert sortable attributes: %v", err)
		}
	}
	if currentSortable == nil || !utils.SameStringSet(currentSortableSlice, desiredSortable) {
		sortableAttributes := desiredSortable
		if _, err := db.Index(ticketMessage.TableName()).UpdateSortableAttributes(&sortableAttributes); err != nil {
			zsLog.Fatalf("[NewMeiliTicketMessageRepository] failed to update sortable attributes: %v", err)
		}
	}

	return &MeiliTicketMessageRepository{
		ticketMessage: ticketMessage,
		db:            db,
	}
}

func (r *MeiliTicketMessageRepository) AddDocuments(ctx context.Context, messages []model.TicketMessage) (*meilisearch.TaskInfo, error) {
	primaryKey := r.ticketMessage.PKName()
	taskInfo, err := r.db.Index(r.ticketMessage.TableName()).AddDocuments(messages, &meilisearch.DocumentOptions{
		PrimaryKey: &primaryKey,
	})
	return taskInfo, err
}

func (r *MeiliTicketMessageRepository) AddDocumentsSync(ctx context.Context, messages []model.TicketMessage) error {
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

func (r *MeiliTicketMessageRepository) GetFiltered(ctx context.Context, filterRequest filter_request.FilterRequest[dto.TicketMessageGetByTicketIDRequest]) ([]model.TicketMessage, int64, filter_request.Paginate, error) {
	var messages []model.TicketMessage
	filters, sort, paginate, err := filter_request.InitializeFilter(filterRequest, r.ticketMessage.AllowedFilterFields(), r.ticketMessage.AllowedSortFields())
	if err != nil {
		return messages, 0, paginate, err
	}
	filters = append(filters, filter_request.Filter{
		Field:    "ticket_id",
		Operator: filter_request.OpEq,
		Value:    filterRequest.SpecificFilter.TicketID,
	})
	searchRequest, err := filter_request.ApplyFilterMeili(filters, sort, paginate)
	if err != nil {
		return messages, 0, paginate, err
	}
	searched, err := r.db.Index(r.ticketMessage.TableName()).Search(filterRequest.SpecificFilter.Search, searchRequest)
	if err != nil {
		return messages, 0, paginate, err
	}
	for _, hit := range searched.Hits {
		var message model.TicketMessage
		err := hit.DecodeInto(&message)
		if err != nil {
			return messages, 0, paginate, err
		}
		messages = append(messages, message)
	}
	return messages, searched.TotalHits, paginate, err
}
