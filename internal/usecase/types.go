package usecase

import (
	"context"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"

	"cloud.google.com/go/storage"
)

type (
	ActivityLog interface {
		// Creates a new activity log entry in the system based on the provided data.
		Insert(ctx context.Context, inputData dto.ActivityLogCreateRequest) (model.ActivityLog, bool, error)
	}

	Message interface {
		SendMessage(ctx context.Context, inputData dto.MessageSendRequest) (model.Message, bool, error)
		GetTemplateList(ctx context.Context, inputData dto.TemplateListRequest) ([]any, bool, error)
	}

	StorageMedia interface {
		UploadMedia(ctx context.Context, inputData dto.StorageMediaUploadRequest) (dto.StorageMediaUploadResponse, bool, error)
		GetMedia(ctx context.Context, inputData dto.StorageMediaGetRequest) (*storage.Reader, *storage.ObjectAttrs, bool, error)
	}
)
