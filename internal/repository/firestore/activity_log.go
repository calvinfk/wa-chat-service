package repository_firestore

import (
	"context"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/formatter"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

type ActivityLogRepository struct {
	model           model.ActivityLog
	firestoreClient *firestore.Client
}

func NewActivityLogRepository(firestoreClient *firestore.Client) *ActivityLogRepository {
	return &ActivityLogRepository{
		firestoreClient: firestoreClient,
	}
}

func (r *ActivityLogRepository) Insert(ctx context.Context, tx *firestore.Transaction, data model.ActivityLog) (model.ActivityLog, error) {
	var err error
	data.ID, err = uuid.NewV7()
	if err != nil {
		return r.model, err
	}
	data.CreatedAt = time.Now()
	dataMap, err := formatter.StructToMap(data, true)
	if err != nil {
		return r.model, err
	}
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		docRef := r.firestoreClient.Collection(r.model.TableName()).Doc(data.ID.String())
		return tx.Create(docRef, dataMap)
	}
	if tx == nil {
		err = r.firestoreClient.RunTransaction(ctx, execDB)
	} else {
		err = execDB(ctx, tx)
	}
	if err != nil {
		return r.model, err
	}
	return data, nil
}

func (r *ActivityLogRepository) GetFiltered(ctx context.Context, filter filter_request.FilterRequest[dto.ActivityLogFilterRequest]) (filter_request.FilterResponse[dto.ActivityLogResponse], error) {
	var response filter_request.FilterResponse[dto.ActivityLogResponse]
	var activityLogs []model.ActivityLog
	filters, sort, paginate, err := filter_request.InitializeFilter(filter, r.model.AllowedFilterFields(), r.model.AllowedSortFields())
	if err != nil {
		return response, err
	}
	query := r.firestoreClient.Collection(r.model.TableName())
	totalData, err := filter_request.ApplyFilterFirestore(ctx, query, &activityLogs, filters, paginate, sort)
	var result []dto.ActivityLogResponse
	for _, activityLog := range activityLogs {
		var data dto.ActivityLogResponse
		data.FromModel(activityLog)
		result = append(result, data)
	}
	response = filter_request.NewFilterResponse(result, paginate, totalData)
	return response, nil
}
