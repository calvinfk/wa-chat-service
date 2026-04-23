package repository

import (
	"context"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/filter_request"

	"cloud.google.com/go/firestore"
)

type (
	// Chat defines persistence operations for chat data.
	Chat interface {
		// Upsert inserts or updates a chat entry.
		Upsert(ctx context.Context, tx *firestore.Transaction, data model.Chat) (model.Chat, error)
		// GetChatByPhoneNumberID gets chat entries filtered by phone number ID.
		GetChatByPhoneNumberID(ctx context.Context, requestData filter_request.FilterRequest[dto.ChatGetByPhoneNumberIDRequest]) (filter_request.FilterResponse[dto.ChatGetByPhoneNumberIDResponse], error)
	}

	// Message defines persistence operations for message data.
	Message interface {
		// Upsert inserts or updates a message entry.
		Upsert(ctx context.Context, tx *firestore.Transaction, data model.Message) (model.Message, error)
		// GetMessageByChatID gets message entries filtered by chat ID.
		GetMessageByChatID(ctx context.Context, requestData filter_request.FilterRequest[dto.MessageGetByChatIDRequest]) (filter_request.FilterResponse[dto.MessageResponse], error)
	}

	// StorageMedia defines persistence operations for media storage metadata.
	StorageMedia interface {
		// Insert inserts a media entry.
		Insert(ctx context.Context, tx *firestore.Transaction, data model.StorageMedia) (model.StorageMedia, error)
		// GetByDocumentID gets a media entry by document ID.
		GetByDocumentID(ctx context.Context, documentID string) (model.StorageMedia, error)
		// GetByDocumentIDs gets media entries by multiple document IDs.
		GetByDocumentIDs(ctx context.Context, documentIDs []string) (map[string]model.StorageMedia, error)
		// GetByURL gets a media entry by URL.
		GetByURL(ctx context.Context, url string) (model.StorageMedia, error)
		// GetByMediaID gets a media entry by media ID.
		GetByMediaID(ctx context.Context, mediaID string) (model.StorageMedia, error)
		// Delete deletes a media entry by document ID.
		Delete(ctx context.Context, tx *firestore.Transaction, documentID string) error
		// Update updates an existing media entry.
		Update(ctx context.Context, tx *firestore.Transaction, data model.StorageMedia) error
		// GetFiltered gets media entries by filter criteria with pagination metadata.
		GetFiltered(ctx context.Context, inputData filter_request.FilterRequest[dto.StorageMediaGetListRequest]) ([]model.StorageMedia, filter_request.Paginate, int64, error)
	}

	// Tenant defines persistence operations for tenant and contact data.
	Tenant interface {
		// GetByID gets a tenant by tenant ID.
		GetByID(ctx context.Context, tenantID string) (model.Tenant, error)
		// GetByPhoneNumberID gets a tenant by phone number ID.
		GetByPhoneNumberID(ctx context.Context, phoneNumberID string) (model.Tenant, error)
		// InsertContact inserts a new contact for a tenant.
		InsertContact(ctx context.Context, contact model.Contact) error
		// GetContactsFiltered gets contacts by filter criteria.
		GetContactsFiltered(ctx context.Context, tenantID string, filterRequest filter_request.FilterRequest[dto.ContactGetFilteredRequest]) (filter_request.FilterResponse[dto.ContactResponse], error)
		// GetContactByPhoneNumbers gets contacts mapped by provided phone numbers.
		GetContactByPhoneNumbers(ctx context.Context, tenantID string, phoneNumbers []string) (map[string]map[string]string, error)
		// GetContactByID gets a contact by contact ID.
		GetContactByID(ctx context.Context, tenantID string, contactID string) (model.Contact, error)
		// UpdateContact updates an existing contact.
		UpdateContact(ctx context.Context, contact model.Contact) error
		// GetTemplateFields gets template fields for a tenant.
		GetTemplateFields(ctx context.Context, tenantID string) (map[string]model.TemplateField, error)
		// DeleteContact deletes a contact by contact ID.
		DeleteContact(ctx context.Context, tenantID string, contactID string) error
	}

	// Template defines persistence operations for template data.
	Template interface {
		// GetFilteredByTenantID gets templates by tenant ID and filter criteria.
		GetFilteredByTenantID(ctx context.Context, tenantID string, requestData filter_request.FilterRequest[dto.TemplateGetByTenantID]) (filter_request.FilterResponse[dto.TemplateResponse], error)
		// GetAll gets all templates for a tenant.
		GetAll(ctx context.Context, tenantID string) ([]model.Template, error)
		// GetByID gets a template by document ID.
		GetByID(ctx context.Context, tenantID string, documentID string) (model.Template, error)
		// Upsert inserts or updates a template entry.
		Upsert(ctx context.Context, tx *firestore.Transaction, data model.Template) (model.Template, error)
		// DeleteByID deletes a template by document ID.
		DeleteByID(ctx context.Context, tx *firestore.Transaction, tenantID string, documentID string) error
		// DeleteByName deletes templates by name.
		DeleteByName(ctx context.Context, tx *firestore.Transaction, tenantID string, name string) error
	}

	// Broadcast defines persistence operations for broadcast data and recipients.
	Broadcast interface {
		// Insert inserts a new broadcast entry.
		Insert(ctx context.Context, tx *firestore.Transaction, broadcast model.Broadcast) error
		// GetByID gets a broadcast by broadcast ID.
		GetByID(ctx context.Context, broadcastID string) (model.Broadcast, error)
		// Update updates an existing broadcast entry.
		Update(ctx context.Context, tx *firestore.Transaction, broadcast model.Broadcast) error
		// Delete deletes a broadcast by broadcast ID.
		Delete(ctx context.Context, tx *firestore.Transaction, broadcastID string) error
		// InsertRecipient inserts a broadcast recipient entry.
		InsertRecipient(ctx context.Context, tx *firestore.Transaction, broadcastRecipient model.BroadcastRecipient) error
		// GetRecipientsByBroadcastID gets recipients by broadcast ID.
		GetRecipientsByBroadcastID(ctx context.Context, broadcastID string) ([]model.BroadcastRecipient, error)
		// UpdateRecipientStatus updates recipient status.
		UpdateRecipientStatus(ctx context.Context, tx *firestore.Transaction, data model.BroadcastRecipient) error
		// GetFiltered gets broadcast entries by filter criteria.
		GetFiltered(ctx context.Context, inputData filter_request.FilterRequest[dto.BroadcastGetFilteredRequest]) (filter_request.FilterResponse[dto.BroadcastResponse], error)
		// GetRecipientsFiltered gets broadcast recipients by filter criteria.
		GetRecipientsFiltered(ctx context.Context, inputData filter_request.FilterRequest[dto.BroadcastGetRecipientsFilteredRequest]) (filter_request.FilterResponse[dto.BroadcastRecipientResponse], error)
	}

	// SearchTemplate defines indexing and query operations for template search.
	SearchTemplate interface {
		// AddDocuments adds template documents to the search index.
		AddDocuments(ctx context.Context, document []model.Template) error
		// DeleteDocuments deletes template documents from the search index.
		DeleteDocuments(ctx context.Context, documentIDs []string) error
		// GetFiltered gets indexed templates by filter criteria with pagination metadata.
		GetFiltered(ctx context.Context, filterRequest filter_request.FilterRequest[dto.TemplateGetByTenantID]) ([]model.Template, int64, filter_request.Paginate, error)
	}
	// SearchMessage defines indexing and query operations for message search.
	SearchMessage interface {
		// AddDocuments adds message documents to the search index.
		AddDocuments(ctx context.Context, document []model.Message) error
		// GetFiltered gets indexed messages by filter criteria with pagination metadata.
		GetFiltered(ctx context.Context, filter filter_request.FilterRequest[dto.MessageGetByChatIDRequest]) ([]model.Message, int64, filter_request.Paginate, error)
	}
)
