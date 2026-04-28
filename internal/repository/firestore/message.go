package repository_firestore

import (
	"context"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/service"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type MessageRepository struct {
	chat                 model.Chat
	message              model.Message
	db                   *firestore.Client
	googleStorageService service.GoogleStorage
}

func NewMessageRepository(db *firestore.Client, googleStorageService service.GoogleStorage) *MessageRepository {
	return &MessageRepository{db: db, googleStorageService: googleStorageService}
}

func (r *MessageRepository) Upsert(ctx context.Context, tx *firestore.Transaction, message model.Message) (model.Message, error) {
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		docRef := r.db.
			Collection(r.chat.TableName()).Doc(message.ChatID).
			Collection(r.message.TableName()).Doc(message.DocumentID)
		txErr := tx.Set(docRef, message)
		if txErr != nil {
			return txErr
		}
		return nil
	}

	var err error
	if tx == nil {
		err = r.db.RunTransaction(ctx, execDB)
	} else {
		err = execDB(ctx, tx)
	}
	return message, err
}

func (r *MessageRepository) GetByChatID(ctx context.Context, requestData filter_request.FilterRequest[dto.MessageGetByChatIDRequest]) (filter_request.FilterResponse[dto.MessageResponse], error) {
	var response filter_request.FilterResponse[dto.MessageResponse]
	filters, sort, paginate, err := filter_request.InitializeFilter(requestData, r.message.AllowedFilterFields(), r.message.AllowedSortFields())
	if err != nil {
		return response, err
	}
	collection := r.db.Collection(r.chat.TableName()).Doc(requestData.SpecificFilter.ChatID).Collection(r.message.TableName())
	query := collection.Query
	docs, totalData, err := filter_request.ApplyFilterFirestore(ctx, query, filters, sort, paginate)
	if err != nil {
		return response, err
	}
	var result []dto.MessageResponse
	for _, doc := range docs {
		var message model.Message
		docData := doc.Data()
		docData["id"] = doc.Ref.ID
		docData["chat_id"] = doc.Ref.Parent.Parent.ID
		err := utils.MapToStruct(docData, &message)
		if err != nil {
			return response, err
		}
		// get storage media if exist
		if message.StorageMediaID != nil {
			var storageMedia model.StorageMedia
			storageMediaDoc, err := r.db.Collection(storageMedia.TableName()).Doc(*message.StorageMediaID).Get(ctx)
			if err != nil || !storageMediaDoc.Exists() {
			} else {
				storageMediaData := storageMediaDoc.Data()
				storageMediaData["id"] = storageMediaDoc.Ref.ID
				err := utils.MapToStruct(storageMediaData, &storageMedia)
				if err != nil {
					return response, err
				}
			}
		}
		// sign storage media url
		var sto *dto.StorageMediaResponse
		if message.StorageMedia != nil {
			var accessURL *string
			accessURL = message.StorageMedia.URL
			if message.StorageMedia.IsURLFromStorage {
				signedURL, err := r.googleStorageService.GenerateV4GetObjectSignedURL(*message.StorageMedia.URL, 0)
				if err != nil {
					return response, err
				}
				accessURL = &signedURL
			}
			storageMediaResponse := dto.StorageMediaResponse{}.FromModel(*message.StorageMedia, accessURL)
			sto = &storageMediaResponse
		}
		result = append(result, dto.MessageResponse{}.FromModel(message, sto))
	}
	response = filter_request.NewFilterResponse(result, paginate, totalData)
	return response, nil
}

func (r *MessageRepository) GetByWamid(ctx context.Context, chatID string, wamid string) (model.Message, error) {
	var message model.Message
	doc, err := r.db.
		Collection(r.chat.TableName()).Doc(chatID).
		Collection(r.message.TableName()).
		Where("wamid", "==", wamid).Limit(1).Documents(ctx).Next()
	if err != nil {
		if err == iterator.Done {
			return message, errs.ErrGenericNotFound
		}
		return message, err
	}
	docData := doc.Data()
	docData["id"] = doc.Ref.ID
	docData["chat_id"] = doc.Ref.Parent.Parent.ID
	err = utils.MapToStruct(docData, &message)
	return message, err
}
