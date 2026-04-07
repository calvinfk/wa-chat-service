package repository_firestore

import (
	"context"
	"log"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/service"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/formatter"

	"cloud.google.com/go/firestore"
)

type MessageRepository struct {
	message model.Message
	// messageLog           model.MessageLog
	db                   *firestore.Client
	googleStorageService service.GoogleStorage
}

func NewMessageRepository(db *firestore.Client, googleStorageService service.GoogleStorage) *MessageRepository {
	return &MessageRepository{db: db, googleStorageService: googleStorageService}
}

func (r *MessageRepository) Upsert(ctx context.Context, tx *firestore.Transaction, message model.Message) (model.Message, error) {
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		docRef := r.db.
			Collection("chats").Doc(message.ChatID).
			Collection("messages").Doc(message.DocumentID)
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

func (r *MessageRepository) GetMessageByChatID(ctx context.Context, requestData filter_request.FilterRequest[dto.MessageGetByChatIDRequest]) (filter_request.FilterResponse[dto.MessageGetByChatIDResponse], error) {
	var response filter_request.FilterResponse[dto.MessageGetByChatIDResponse]
	filters, sort, paginate, err := filter_request.InitializeFilter(requestData, r.message.AllowedFilterFields(), r.message.AllowedSortFields())
	if err != nil {
		return response, err
	}
	collection := r.db.Collection("chats").Doc(requestData.SpecificFilter.ChatID).Collection(r.message.TableName())
	query := collection.Query
	docs, totalData, err := filter_request.ApplyFilterFirestore(ctx, query, filters, paginate, sort)
	if err != nil {
		return response, err
	}
	var result []dto.MessageGetByChatIDResponse
	for _, doc := range docs {
		var message model.Message
		docData := doc.Data()
		docData[firestore.DocumentID] = doc.Ref.ID
		docData["chat_id"] = doc.Ref.Parent.Parent.ID
		err := formatter.MapToStruct(docData, &message)
		if err != nil {
			return response, err
		}
		// get storage media if exist
		if message.StorageMediaID != nil {
			var storageMedia model.StorageMedia
			storageMediaDoc, err := r.db.Collection(storageMedia.TableName()).Doc(*message.StorageMediaID).Get(ctx)
			if err == nil && storageMediaDoc.Exists() {
				storageMediaData := storageMediaDoc.Data()
				storageMediaData[firestore.DocumentID] = storageMediaDoc.Ref.ID
				err := formatter.MapToStruct(storageMediaData, &storageMedia)
				if err != nil {
					log.Println("[ERROR][internal/repository/firestore/message.go][GetMessageByChatID] Failed to map storage media data:", err)
				} else {
					message.StorageMedia = &storageMedia
				}
			} else {
				log.Println("[INFO][internal/repository/firestore/message.go][GetMessageByChatID] No storage media found for ID:", *message.StorageMediaID, "err: ", err) // log if no storage media found
			}
		}
		// sign storage media url
		var sto *dto.StorageMediaUploadResponse
		if message.StorageMedia != nil {
			var storageMediaUploadResponse dto.StorageMediaUploadResponse
			signedUrl, err := r.googleStorageService.GenerateV4GetObjectSignedURL(*message.StorageMedia.URL, 0)
			if err != nil {
				log.Println("[ERROR][internal/repository/firestore/message.go][GetMessageByChatID] Failed to generate signed URL for storage media:", err)
				signedUrl = ""
			}
			storageMediaUploadResponse.FromModel(*message.StorageMedia, signedUrl)
			sto = &storageMediaUploadResponse
		}
		var data dto.MessageGetByChatIDResponse
		data.FromModel(message, sto)
		result = append(result, data)
	}
	response = filter_request.NewFilterResponse(result, paginate, totalData)
	return response, nil
}
