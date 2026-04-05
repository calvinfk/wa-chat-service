package repository_firestore

import (
	"context"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/formatter"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChatRepository struct {
	chat model.Chat
	db   *firestore.Client
}

func NewChatRepository(db *firestore.Client) *ChatRepository {
	return &ChatRepository{
		db: db,
	}
}

func (r *ChatRepository) Insert(ctx context.Context, tx *firestore.Transaction, chat model.Chat) (model.Chat, error) {
	var err error
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		doc := r.db.Collection("chats").Doc(chat.DocumentID)
		err := tx.Update(doc, []firestore.Update{
			{Path: "last_message", Value: chat.LastMessage},
			{Path: "updated_at", Value: chat.UpdatedAt},
		})
		if err == nil {
			return nil
		}
		if status.Code(err) != codes.NotFound {
			return err
		}
		err = tx.Set(doc, chat)
		return err
	}
	if tx == nil {
		err = r.db.RunTransaction(ctx, execDB)
	} else {
		err = execDB(ctx, tx)
	}
	return chat, err
}

func (r *ChatRepository) GetChatByPhoneNumberID(ctx context.Context, filter filter_request.FilterRequest[dto.ChatGetByPhoneNumberIDRequest]) (filter_request.FilterResponse[dto.ChatGetByPhoneNumberIDResponse], error) {
	var response filter_request.FilterResponse[dto.ChatGetByPhoneNumberIDResponse]
	filters, sort, paginate, err := filter_request.InitializeFilter(filter, r.chat.AllowedFilterFields(), r.chat.AllowedSortFields())
	if err != nil {
		return response, err
	}
	collection := r.db.Collection(r.chat.TableName())
	query := collection.Query
	docs, totalData, err := filter_request.ApplyFilterFirestore(ctx, query, filters, paginate, sort)
	if err != nil {
		return response, err
	}
	var result []dto.ChatGetByPhoneNumberIDResponse
	for _, doc := range docs {
		var chat model.Chat
		docData := doc.Data()
		docData[firestore.DocumentID] = doc.Ref.ID
		err := formatter.MapToStruct(docData, &chat)
		if err != nil {
			return response, err
		}
		var data dto.ChatGetByPhoneNumberIDResponse
		data.FromModel(chat)
		result = append(result, data)
	}
	response = filter_request.NewFilterResponse(result, paginate, totalData)
	return response, nil
}
