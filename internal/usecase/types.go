package usecase

import (
	"context"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/filter_request"

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
		GetMessagesByChatID(ctx context.Context, requestData filter_request.FilterRequest[dto.MessageGetByChatIDRequest]) (filter_request.FilterResponse[dto.MessageGetByChatIDResponse], bool, error)
	}

	StorageMedia interface {
		UploadMedia(ctx context.Context, inputData dto.StorageMediaUploadRequest) (dto.StorageMediaUploadResponse, bool, error)
		GetMedia(ctx context.Context, inputData dto.StorageMediaGetRequest) (*storage.Reader, *storage.ObjectAttrs, bool, error)
		DeleteMedia(ctx context.Context, inputData dto.StorageMediaDeleteRequest) (bool, error)
		UploadMediaUsingMediaID(ctx context.Context, inputData dto.StorageMediaUploadUsingMediaIDRequest) (string, bool, error)
		StoreMediaFromURL(ctx context.Context, fileURL string) (model.StorageMedia, bool, error)
	}

	Chat interface {
		GetChatByPhoneNumberID(ctx context.Context, requestData filter_request.FilterRequest[dto.ChatGetByPhoneNumberIDRequest]) (filter_request.FilterResponse[dto.ChatGetByPhoneNumberIDResponse], bool, error)
	}
)
