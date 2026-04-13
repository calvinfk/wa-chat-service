package usecase

import (
	"context"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/meta/whatsapp_business"
)

type (
	ActivityLog interface {
		// Creates a new activity log entry in the system based on the provided data.
		Insert(ctx context.Context, inputData dto.ActivityLogCreateRequest) (model.ActivityLog, bool, error)
	}

	Message interface {
		SendMessage(ctx context.Context, whatsappClient *whatsapp_business.Client, tenantID string, inputData dto.MessageSendRequest) (model.Message, bool, error)
		GetMessagesByChatID(ctx context.Context, requestData filter_request.FilterRequest[dto.MessageGetByChatIDRequest]) (filter_request.FilterResponse[dto.MessageGetByChatIDResponse], bool, error)
	}

	Template interface {
		CreateTemplate(ctx context.Context, inputData dto.TemplateCreateRequest) (any, bool, error)
		GetFilteredByPhoneNumberID(ctx context.Context, inputData filter_request.FilterRequest[dto.TemplateGetByPhoneNumberID]) (filter_request.FilterResponse[dto.TemplateGetByPhoneNumberIDResponse], bool, error)
		SyncTemplate(ctx context.Context, inputData dto.TemplateSyncRequest) (bool, error)
		DeleteTemplate(ctx context.Context, inputData dto.TemplateDeleteRequest) (bool, error)
		UpdateTemplate(ctx context.Context, inputData dto.TemplateUpdateRequest) (bool, error)
	}

	StorageMedia interface {
		UploadMedia(ctx context.Context, inputData dto.StorageMediaUploadRequest) (dto.StorageMediaResponse, bool, error)
		// the caller is responsible to close the reader after use
		GetMedia(ctx context.Context, inputData dto.StorageMediaGetRequest) (dto.StorageMediaGetMediaResponse, bool, error)
		DeleteMedia(ctx context.Context, inputData dto.StorageMediaDeleteRequest) (bool, error)
		SaveMediaID(ctx context.Context, inputData dto.StorageMediaSaveMediaIDRequest) (dto.StorageMediaSaveMediaIDResponse, bool, error)
		UploadResumableMedia(ctx context.Context, inputData dto.StorageMediaResumableUploadRequest) (dto.StorageMediaResumableUploadResponse, bool, error)
		UploadMediaMeta(ctx context.Context, inputData dto.StorageMediaUploadMetaRequest) (dto.StorageMediaUploadMetaResponse, bool, error)
		GetFiltered(ctx context.Context, inputData filter_request.FilterRequest[dto.StorageMediaGetListRequest]) (filter_request.FilterResponse[dto.StorageMediaResponse], bool, error)
	}

	Chat interface {
		GetChatByPhoneNumberID(ctx context.Context, requestData filter_request.FilterRequest[dto.ChatGetByPhoneNumberIDRequest]) (filter_request.FilterResponse[dto.ChatGetByPhoneNumberIDResponse], bool, error)
	}

	Tenant interface {
		GetWhatsappClient(ctx context.Context, phoneNumberID string) (*whatsapp_business.Client, string, error)
		CreateContact(ctx context.Context, inputData dto.ContactCreateRequest) (bool, error)
		GetContactsFiltered(ctx context.Context, inputData filter_request.FilterRequest[dto.ContactGetFilteredRequest]) (filter_request.FilterResponse[dto.ContactResponse], bool, error)
		UpdateContact(ctx context.Context, inputData dto.ContactUpdateRequest) (bool, error)
	}

	Broadcast interface {
		ScheduleBroadcast(ctx context.Context, inputData dto.BroadcastScheduleRequest) (bool, error)
		SendBroadcast(ctx context.Context, broadcastID string) (bool, error)
	}

	Auth interface {
		Login(ctx context.Context, req dto.AuthLoginRequest) (string, bool, error)
	}
)
