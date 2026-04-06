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
		// Inserts or updates a chat entry.
		Upsert(ctx context.Context, tx *firestore.Transaction, data model.Chat) (model.Chat, error)
		GetChatByPhoneNumberID(ctx context.Context, requestData filter_request.FilterRequest[dto.ChatGetByPhoneNumberIDRequest]) (filter_request.FilterResponse[dto.ChatGetByPhoneNumberIDResponse], error)
	}

	Message interface {
		// Inserts or updates a message entry.
		Upsert(ctx context.Context, tx *firestore.Transaction, data model.Message) (model.Message, error)
		// Insert Log Message entry.
		InsertLog(ctx context.Context, tx *firestore.Transaction, data model.MessageLog) (model.MessageLog, error)
		GetMessageByChatID(ctx context.Context, requestData filter_request.FilterRequest[dto.MessageGetByChatIDRequest]) (filter_request.FilterResponse[dto.MessageGetByChatIDResponse], error)
	}

	StorageMedia interface {
		// Inserts a media entry.
		Insert(ctx context.Context, tx *firestore.Transaction, data model.StorageMedia) (model.StorageMedia, error)
		// Gets media entry by document ID.
		GetByDocumentID(ctx context.Context, documentID string) (model.StorageMedia, error)
		// Gets media entry by URL.
		GetByURL(ctx context.Context, url string) (model.StorageMedia, error)
		GetByAccessURL(ctx context.Context, accessURL string) (model.StorageMedia, error)
		// Gets media entry by media ID.
		GetByMediaID(ctx context.Context, mediaID string) (model.StorageMedia, error)
		// Deletes media entry by document ID.
		Delete(ctx context.Context, tx *firestore.Transaction, documentID string) error
		// Updates media entry by document ID.
		UpdateAccessURL(ctx context.Context, tx *firestore.Transaction, documentID string, accessURL string) error
	}

	PhoneNumber interface {
		GetByPhoneNumberID(ctx context.Context, phoneNumberID string) (model.PhoneNumber, error)
	}
)
