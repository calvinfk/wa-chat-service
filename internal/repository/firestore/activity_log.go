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
	logID, err := uuid.NewV7()
	if err != nil {
		return r.model, err
	}
	data.ID = logID.String()
	data.CreatedAt = time.Now()
	dataMap, err := formatter.StructToMap(data, true)
	if err != nil {
		return r.model, err
	}
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		docRef := r.firestoreClient.Collection(r.model.TableName()).Doc(data.ID)
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
	// var activityLogs []model.ActivityLog
	filters, sort, paginate, err := filter_request.InitializeFilter(filter, r.model.AllowedFilterFields(), r.model.AllowedSortFields())
	if err != nil {
		return response, err
	}
	collection := r.firestoreClient.Collection(r.model.TableName())
	query := collection.Query
	docs, totalData, err := filter_request.ApplyFilterFirestore(ctx, query, filters, paginate, sort)
	if err != nil {
		return response, err
	}
	var result []dto.ActivityLogResponse
	for _, doc := range docs {
		var data model.ActivityLog
		err := doc.DataTo(&data)
		if err != nil {
			return response, err
		}
		var responseData dto.ActivityLogResponse
		responseData.FromModel(data)
		result = append(result, responseData)
	}
	response = filter_request.NewFilterResponse(result, paginate, totalData)
	return response, nil
}
