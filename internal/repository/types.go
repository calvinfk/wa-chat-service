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

	Chat interface {
		// Inserts a chat entry.
		Insert(ctx context.Context, tx *firestore.Transaction, data model.Chat) (model.Chat, error)
	}

	Message interface {
		// Inserts or updates a message entry.
		Upsert(ctx context.Context, tx *firestore.Transaction, data model.Message) (model.Message, error)
		// Insert Log Message entry.
		InsertLog(ctx context.Context, tx *firestore.Transaction, data model.MessageLog) (model.MessageLog, error)
	}

	StorageMedia interface {
		// Inserts a media entry.
		Insert(ctx context.Context, tx *firestore.Transaction, data model.StorageMedia) (model.StorageMedia, error)
		// Gets media entry by document ID.
		GetByDocumentID(ctx context.Context, documentID string) (model.StorageMedia, error)
	}
)
