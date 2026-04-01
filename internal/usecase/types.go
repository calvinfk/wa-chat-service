package usecase

import (
	"context"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
)

type (
	ActivityLog interface {
		// Creates a new activity log entry in the system based on the provided data.
		Insert(ctx context.Context, inputData dto.ActivityLogCreateRequest) (model.ActivityLog, bool, error)
	}

	Message interface {
		SendMessage(ctx context.Context, inputData dto.MessageSendRequest) (model.Message, bool, error)
	}
)
