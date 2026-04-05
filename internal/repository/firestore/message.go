package repository_firestore

import (
	"context"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/formatter"

	"cloud.google.com/go/firestore"
)

type MessageRepository struct {
	message    model.Message
	messageLog model.MessageLog
	db         *firestore.Client
}

func NewMessageRepository(db *firestore.Client) *MessageRepository {
	return &MessageRepository{db: db}
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

func (r *MessageRepository) InsertLog(ctx context.Context, tx *firestore.Transaction, messageLog model.MessageLog) (model.MessageLog, error) {
	var err error
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		_, err := r.db.
			Collection("chats").Doc(messageLog.ChatID).
			Collection("messages_log").Doc(messageLog.DocumentID.String()).
			Set(ctx, messageLog)
		if err != nil {
			return err
		}
		return nil
	}
	if tx == nil {
		err = r.db.RunTransaction(ctx, execDB)
	} else {
		err = execDB(ctx, tx)
	}
	return messageLog, err
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
		var data dto.MessageGetByChatIDResponse
		data.FromModel(message)
		result = append(result, data)
	}
	response = filter_request.NewFilterResponse(result, paginate, totalData)
	return response, nil
}
