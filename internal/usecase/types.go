package usecase

import (
	"context"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/filter_request"
)

type (
	ActivityLog interface {
		// Creates a new activity log entry in the system based on the provided data.
		Insert(ctx context.Context, inputData dto.ActivityLogCreateRequest) (model.ActivityLog, bool, error)
		// Retrieves action log from the system based on the provided filter criteria
		GetFiltered(ctx context.Context, user model.User, filter filter_request.FilterRequest[dto.ActivityLogFilterRequest]) (filter_request.FilterResponse[dto.ActivityLogResponse], bool, error)
	}
)
