package repository

import (
	"context"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/filter_request"

	"cloud.google.com/go/firestore"
)

type (
	ActivityLog interface {
		// Inserts an activity log entry.
		Insert(ctx context.Context, tx *firestore.Transaction, data model.ActivityLog) (model.ActivityLog, error)
		// Gets activity log entries by filter.
		GetFiltered(ctx context.Context, filter filter_request.FilterRequest[dto.ActivityLogFilterRequest]) (filter_request.FilterResponse[dto.ActivityLogResponse], error)
	}
)
