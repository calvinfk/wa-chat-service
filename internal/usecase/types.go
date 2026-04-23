package usecase

import (
	"context"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/meta/whatsapp_business"
)

type (
	// Message defines chat message operations including send, read, and persistence flows.
	Message interface {
		// SendMessage sends a WhatsApp message and persists its state.
		// Returns the saved message model, a server-error flag (true if error is from server), and an error.
		SendMessage(ctx context.Context, whatsappClient *whatsapp_business.Client, tenantID string, inputData dto.MessageSendRequest) (model.Message, bool, error)
		// GetMessagesByChatID fetches paginated messages for a chat.
		// Returns a filtered response payload, a server-error flag (true if error is from server), and an error.
		GetMessagesByChatID(ctx context.Context, requestData filter_request.FilterRequest[dto.MessageGetByChatIDRequest]) (filter_request.FilterResponse[dto.MessageResponse], bool, error)
		// SaveMessage stores an inbound or outbound message without sending it.
		// Returns a server-error flag (true if error is from server) and an error.
		SaveMessage(ctx context.Context, inputData dto.MessageSaveRequest) (bool, error)
	}

	// Template defines template lifecycle operations such as create, sync, update, delete, and filtered retrieval.
	Template interface {
		// CreateTemplate creates a WhatsApp template and stores metadata in local storage.
		// Returns the provider response payload, a server-error flag (true if error is from server), and an error.
		CreateTemplate(ctx context.Context, inputData dto.TemplateCreateRequest) (any, bool, error)
		// GetFilteredByTenantID retrieves templates using filter, sort, and pagination rules.
		// Returns a filtered template response, a server-error flag (true if error is from server), and an error.
		GetFilteredByTenantID(ctx context.Context, inputData filter_request.FilterRequest[dto.TemplateGetByTenantID]) (filter_request.FilterResponse[dto.TemplateResponse], bool, error)
		// SyncTemplate synchronizes template data between provider and local storage.
		// Returns a server-error flag (true if error is from server) and an error.
		SyncTemplate(ctx context.Context, inputData dto.TemplateSyncRequest) (bool, error)
		// DeleteTemplate removes a template from provider/local storage according to business rules.
		// Returns a server-error flag (true if error is from server) and an error.
		DeleteTemplate(ctx context.Context, inputData dto.TemplateDeleteRequest) (bool, error)
		// UpdateTemplate updates template data and reconciles persisted state.
		// Returns a server-error flag (true if error is from server) and an error.
		UpdateTemplate(ctx context.Context, inputData dto.TemplateUpdateRequest) (bool, error)
	}

	// StorageMedia defines media upload, retrieval, deletion, and filtered listing operations.
	StorageMedia interface {
		// UploadMedia uploads media content and records media metadata.
		// Returns media response data, a server-error flag (true if error is from server), and an error.
		UploadMedia(ctx context.Context, inputData dto.StorageMediaUploadRequest) (dto.StorageMediaResponse, bool, error)
		// the caller is responsible to close the reader after use
		// GetMedia retrieves a media stream and its metadata for downstream use.
		// Returns media retrieval data, a server-error flag (true if error is from server), and an error.
		GetMedia(ctx context.Context, inputData dto.StorageMediaGetRequest, rangeHeader string) (dto.StorageMediaGetMediaResponse, bool, error)
		// DeleteMedia deletes stored media and related references.
		// Returns a server-error flag (true if error is from server) and an error.
		DeleteMedia(ctx context.Context, inputData dto.StorageMediaDeleteRequest) (bool, error)
		// SaveMediaID persists a provider media identifier for later access.
		// Returns save result data, a server-error flag (true if error is from server), and an error.
		SaveMediaID(ctx context.Context, inputData dto.StorageMediaSaveMediaIDRequest) (dto.StorageMediaSaveMediaIDResponse, bool, error)
		// GetFiltered returns media records with filtering, sorting, and pagination applied.
		// Returns a filtered media response, a server-error flag (true if error is from server), and an error.
		GetFiltered(ctx context.Context, inputData filter_request.FilterRequest[dto.StorageMediaGetListRequest]) (filter_request.FilterResponse[dto.StorageMediaResponse], bool, error)
		// ParsePublicURL parses a public URL and extract the encrypted media token
		// Returns the extracted file path if parsing is successful, or an error if the URL is invalid.
		ParsePublicURL(url string) (string, error)
		// ParseMediaToken parses a media string which can be a URL or an encrypted media token, and returns the file path, a boolean indicating if it's a URL, and an error if the parsing fails.
		ParseMediaToken(mediaToken string) (string, bool, error)
		// GenerateEncryptedLink generates an encrypted link for a given media ID.
		// Returns the encrypted link, a server-error flag (true if error is from server), and an error.
		GenerateEncryptedLink(ctx context.Context, inputData dto.StorageMediaEncryptLinkRequest) (string, bool, error)
	}

	// Chat defines chat querying operations for conversation-level views.
	Chat interface {
		// GetChatByPhoneNumberID retrieves chat sessions for a phone number with filter options.
		// Returns a filtered chat response, a server-error flag (true if error is from server), and an error.
		GetChatByPhoneNumberID(ctx context.Context, requestData filter_request.FilterRequest[dto.ChatGetByPhoneNumberIDRequest]) (filter_request.FilterResponse[dto.ChatGetByPhoneNumberIDResponse], bool, error)
	}

	// Tenant defines tenant-contact operations and tenant-specific WhatsApp client resolution.
	Tenant interface {
		// GetWhatsappClientByPhone resolves tenant identity and builds a WhatsApp client for outbound calls.
		// Returns the WhatsApp client, resolved tenant ID, and an error when resolution fails.
		GetWhatsappClientByPhone(ctx context.Context, phoneNumberID string) (*whatsapp_business.Client, string, error)
		// GetWhatsappClientByPhone resolves tenant identity and builds a WhatsApp client for outbound calls.
		// Returns the WhatsApp client, resolved tenant ID, and an error when resolution fails.
		GetWhatsappClientByTenant(ctx context.Context, tenantID string) (*whatsapp_business.Client, string, error)
		// CreateContact creates a tenant contact record.
		// Returns a server-error flag (true if error is from server) and an error.
		CreateContact(ctx context.Context, inputData dto.ContactCreateRequest) (bool, error)
		// GetContactsFiltered fetches tenant contacts using filter and pagination options.
		// Returns a filtered contact response, a server-error flag (true if error is from server), and an error.
		GetContactsFiltered(ctx context.Context, inputData filter_request.FilterRequest[dto.ContactGetFilteredRequest]) (filter_request.FilterResponse[dto.ContactResponse], bool, error)
		// UpdateContact updates existing tenant contact data.
		// Returns a server-error flag (true if error is from server) and an error.
		UpdateContact(ctx context.Context, inputData dto.ContactUpdateRequest) (bool, error)
		// DeleteContact deletes a tenant contact record.
		// Returns a server-error flag (true if error is from server) and an error.
		DeleteContact(ctx context.Context, inputData dto.ContactDeleteRequest) (bool, error)
	}

	// Broadcast defines broadcast scheduling, execution, update, cancelation, and read operations.
	Broadcast interface {
		// ScheduleBroadcast creates or updates a broadcast schedule for asynchronous delivery.
		// Returns a server-error flag (true if error is from server) and an error.
		ScheduleBroadcast(ctx context.Context, inputData dto.BroadcastScheduleRequest) (bool, error)
		// SendBroadcast executes broadcast delivery for a specific broadcast job.
		// Returns a server-error flag (true if error is from server) and an error.
		SendBroadcast(ctx context.Context, broadcastID string) (bool, error)
		// UpsertBroadcast creates or updates broadcast metadata and recipients.
		// Returns the resulting broadcast response, a server-error flag (true if error is from server), and an error.
		UpsertBroadcast(ctx context.Context, inputData dto.BroadcastUpsertRequest) (dto.BroadcastResponse, bool, error)
		// CancelBroadcast marks a scheduled or running broadcast as canceled.
		// Returns a server-error flag (true if error is from server) and an error.
		CancelBroadcast(ctx context.Context, inputData dto.BroadcastCancelRequest) (bool, error)
		// GetFilteredBroadcast returns broadcasts using filter/sort/pagination criteria.
		// Returns a filtered broadcast response, a server-error flag (true if error is from server), and an error.
		GetFilteredBroadcast(ctx context.Context, inputData filter_request.FilterRequest[dto.BroadcastGetFilteredRequest]) (filter_request.FilterResponse[dto.BroadcastResponse], bool, error)
		// GetFilteredBroadcastRecipients returns recipient-level rows for a broadcast.
		// Returns a filtered recipient response, a server-error flag (true if error is from server), and an error.
		GetFilteredBroadcastRecipients(ctx context.Context, inputData filter_request.FilterRequest[dto.BroadcastGetRecipientsFilteredRequest]) (filter_request.FilterResponse[dto.BroadcastRecipientResponse], bool, error)
	}

	// Auth defines authentication operations for issuing access tokens.
	Auth interface {
		// Login validates credentials and issues an encrypted access token.
		// Returns the token string, a server-error flag (true if error is from server), and an error.
		Login(ctx context.Context, req dto.AuthLoginRequest) (string, bool, error)
	}
)
