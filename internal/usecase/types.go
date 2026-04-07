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
	}

	Message interface {
		SendMessage(ctx context.Context, inputData dto.MessageSendRequest) (model.Message, bool, error)
		GetTemplateList(ctx context.Context, inputData dto.TemplateListRequest) ([]any, bool, error)
		GetMessagesByChatID(ctx context.Context, requestData filter_request.FilterRequest[dto.MessageGetByChatIDRequest]) (filter_request.FilterResponse[dto.MessageGetByChatIDResponse], bool, error)
	}

	StorageMedia interface {
		UploadMedia(ctx context.Context, inputData dto.StorageMediaUploadRequest) (dto.StorageMediaUploadResponse, bool, error)
		// the caller is responsible to close the reader after use
		GetMedia(ctx context.Context, inputData dto.StorageMediaGetRequest) (dto.StorageMediaGetMediaResponse, bool, error)
		DeleteMedia(ctx context.Context, inputData dto.StorageMediaDeleteRequest) (bool, error)
		SaveMediaID(ctx context.Context, inputData dto.StorageMediaSaveMediaIDRequest) (dto.StorageMediaSaveMediaIDResponse, bool, error)
		StoreMediaFromURL(ctx context.Context, mediaURL string) (model.StorageMedia, bool, error)
	}

	Chat interface {
		GetChatByPhoneNumberID(ctx context.Context, requestData filter_request.FilterRequest[dto.ChatGetByPhoneNumberIDRequest]) (filter_request.FilterResponse[dto.ChatGetByPhoneNumberIDResponse], bool, error)
	}
)
