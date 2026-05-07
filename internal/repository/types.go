package repository

import (
	"context"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/filter_request"

	"cloud.google.com/go/firestore"
	"github.com/meilisearch/meilisearch-go"
)

type (
	// Chat defines persistence operations for chat data.
	Chat interface {
		// Upsert inserts or updates a chat entry.
		// returns the upserted chat, a boolean indicating whether it was created (true) or updated (false), and an error if any.
		Upsert(ctx context.Context, tx *firestore.Transaction, data model.Chat) (model.Chat, bool, error)
		// GetChatByPhoneNumberId gets chat entries filtered by phone number ID.
		GetChatByPhoneNumberId(ctx context.Context, requestData filter_request.FilterRequest[dto.ChatGetByPhoneNumberIdRequest]) (filter_request.FilterResponse[dto.ChatGetByPhoneNumberIdResponse], error)
		// GetByID gets a chat entry by chat ID.
		GetByID(ctx context.Context, chatID string) (model.Chat, error)
		// Update chat last message info by chat ID.
		UpdateLastMessage(ctx context.Context, tx *firestore.Transaction, chatID string, lastMessage string) error
	}

	// Message defines persistence operations for message data.
	Message interface {
		// Upsert inserts or updates a message entry. also updates the search index accordingly.
		Upsert(ctx context.Context, tx *firestore.Transaction, data model.Message) (model.Message, error)
		// GetMessageByWamid gets a message entry by WAMID and chat ID.
		GetMessageByWamid(ctx context.Context, chatID string, wamid string) (model.Message, error)
	}

	// StorageMedia defines persistence operations for media storage metadata.
	StorageMedia interface {
		// Upsert inserts or updates a media entry.
		Upsert(ctx context.Context, tx *firestore.Transaction, data model.StorageMedia) (model.StorageMedia, error)
		// GetByID gets a media entry by ID.
		GetByID(ctx context.Context, ID string) (model.StorageMedia, error)
		// GetByIDs gets media entries by multiple  IDs.
		GetByIDs(ctx context.Context, IDs []string) (map[string]model.StorageMedia, error)
		// Delete deletes a media entry by document ID.
		Delete(ctx context.Context, tx *firestore.Transaction, documentID string) error
		// GetFilteredByTenantID gets media entries by filter criteria with pagination metadata.
		GetFilteredByTenantID(ctx context.Context, tenantID string, inputData filter_request.FilterRequest[dto.StorageMediaGetListRequest]) ([]model.StorageMedia, filter_request.Paginate, int64, error)
	}

	// Tenant defines persistence operations for tenant and contact data.
	Tenant interface {
		// GetByID gets a tenant by tenant ID.
		GetByID(ctx context.Context, tenantID string) (model.Tenant, error)
		// UpsertContact inserts or updates a contact for a tenant.
		UpsertContact(ctx context.Context, tx *firestore.Transaction, contact model.Contact) error
		// GetContactsFiltered gets contacts by filter criteria.
		GetContactsFiltered(ctx context.Context, tenantID string, filterRequest filter_request.FilterRequest[dto.ContactGetFilteredRequest]) (filter_request.FilterResponse[dto.ContactResponse], error)
		// GetContactByPhoneNumbers gets contacts mapped by provided phone numbers.
		GetContactByPhoneNumbers(ctx context.Context, tenantID string, phoneNumbers []string) (map[string]map[string]string, error)
		// GetContactByID gets a contact by contact ID.
		GetContactByID(ctx context.Context, tenantID string, contactID string) (model.Contact, error)
		// GetTemplateFields gets template fields for a tenant.
		GetTemplateFields(ctx context.Context, tenantID string) (map[string]model.TemplateField, error)
		// DeleteContact deletes a contact by contact ID.
		DeleteContact(ctx context.Context, tx *firestore.Transaction, tenantID string, contactID string) error
	}

	// Template defines persistence operations for template data.
	Template interface {
		// GetAll gets all templates for a phone number ID.
		GetAll(ctx context.Context, whatsappBusinessAccountID string) ([]model.Template, error)
		// GetByID gets a template by document ID.
		GetByID(ctx context.Context, whatsappBusinessAccountID string, documentID string) (model.Template, error)
		// Upsert inserts or updates a template entry.
		Upsert(ctx context.Context, tx *firestore.Transaction, data model.Template) (model.Template, error)
		// DeleteByID deletes a template by document ID.
		DeleteByID(ctx context.Context, tx *firestore.Transaction, whatsappBusinessAccountID string, documentID string) error
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
		GetFilteredByTenantID(ctx context.Context, tenantID string, inputData filter_request.FilterRequest[dto.BroadcastGetFilteredRequest]) (filter_request.FilterResponse[dto.BroadcastResponse], error)
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
		GetFiltered(ctx context.Context, filterRequest filter_request.FilterRequest[dto.TemplateFilterRequest]) ([]model.Template, int64, filter_request.Paginate, error)
	}
	// SearchMessage defines indexing and query operations for message search.
	SearchMessage interface {
		// AddDocuments adds message documents to the search index.
		AddDocuments(ctx context.Context, document []model.Message) (*meilisearch.TaskInfo, error)
		// AddDocumentsSync adds message documents to the search index and waits for the operation to complete.
		AddDocumentsSync(ctx context.Context, document []model.Message) error
		// GetFiltered gets indexed messages by filter criteria with pagination metadata.
		GetFiltered(ctx context.Context, filter filter_request.FilterRequest[dto.MessageGetByChatIDRequest]) ([]model.Message, int64, filter_request.Paginate, error)
	}

	User interface {
		// GetByEmail gets a user by email.
		GetByEmail(ctx context.Context, email string) (model.User, error)
		// GetByID gets a user by ID.
		GetByID(ctx context.Context, id string) (model.User, error)
		// GetByTenantIDFiltered gets users by tenant ID with filtering options.
		GetByTenantIDFiltered(ctx context.Context, tenantID string, filter filter_request.FilterRequest[dto.UserListRequest]) (filter_request.FilterResponse[dto.UserResponse], error)
		// Upsert inserts or updates a user entry.
		Upsert(ctx context.Context, tx *firestore.Transaction, user model.User) (model.User, error)
		// GetBySupervisorID gets users by supervisor ID.
		GetBySupervisorID(ctx context.Context, supervisorID string) ([]model.User, error)
	}

	WaBusinessAccount interface {
		// GetByID gets a WhatsApp Business Account by ID
		GetByID(ctx context.Context, id string) (model.WaBusinessAccount, error)
		// GetByTenantID gets WhatsApp Business Accounts by tenant ID
		GetByTenantID(ctx context.Context, tenantID string) ([]model.WaBusinessAccount, error)
	}
	WaPhone interface {
		// GetByPhoneNumberId gets a WhatsApp Business Account phone data by phone number ID from meta.
		GetByPhoneNumberId(ctx context.Context, phoneNumberId string) (model.WaPhone, error)
		// GetByWaBusinessAccountID gets WhatsApp Business Account phone data by WhatsApp Business Account ID from meta.
		GetByWaBusinessAccountID(ctx context.Context, waBusinessAccountID string) ([]model.WaPhone, error)
	}

	Ticket interface {
		// Upsert inserts or updates a ticket entry.
		// returns the upserted ticket, a boolean indicating whether it was created (true) or updated (false), and an error if any.
		Upsert(ctx context.Context, tx *firestore.Transaction, data model.Ticket) (model.Ticket, bool, error)
		// GetRunningTicket gets a running ticket entry (open or in-progress) filtered by phone number Id and recipient Id.
		GetRunningTicket(ctx context.Context, phoneNumberId string, recipientId string) (model.Ticket, error)
		// GetByID gets a ticket entry by ticket ID.
		GetByID(ctx context.Context, ticketID string) (model.Ticket, error)
		// GetTicketDataAnalytics gets ticket entries for ticket data analytics filtered by phone number IDs and created at range.
		GetTicketDataAnalytics(ctx context.Context, phoneNumberIds []string, startTime time.Time, endTime time.Time) ([]model.Ticket, error)
		// Update last message info of a ticket by ticket ID.
		UpdateLastMessage(ctx context.Context, tx *firestore.Transaction, ticketID string, lastMessage string) error
	}

	TicketMessage interface {
		// Upsert inserts or updates a ticket message entry. also updates the search index accordingly.
		Upsert(ctx context.Context, tx *firestore.Transaction, data model.TicketMessage) error
		// GetTicketMessageByWamid gets a ticket message entry by WAMID and ticket ID.
		GetTicketMessageByWamid(ctx context.Context, ticketID string, wamid string) (model.TicketMessage, error)
	}

	SearchTicketMessage interface {
		// AddDocuments adds ticket message documents to the search index.
		AddDocuments(ctx context.Context, document []model.TicketMessage) (*meilisearch.TaskInfo, error)
		// AddDocumentsSync adds ticket message documents to the search index and waits for the operation to complete.
		AddDocumentsSync(ctx context.Context, document []model.TicketMessage) error
		// GetFiltered gets indexed ticket messages by filter criteria with pagination metadata.
		GetFiltered(ctx context.Context, filter filter_request.FilterRequest[dto.TicketMessageGetByTicketIDRequest]) ([]model.TicketMessage, int64, filter_request.Paginate, error)
	}
)
