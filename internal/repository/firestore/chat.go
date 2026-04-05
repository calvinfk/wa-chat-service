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

func (r *ChatRepository) Upsert(ctx context.Context, tx *firestore.Transaction, chat model.Chat) (model.Chat, error) {
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		doc := r.db.Collection("chats").Doc(chat.DocumentID)
		_, getErr := tx.Get(doc)
		if getErr != nil {
			if status.Code(getErr) != codes.NotFound {
				return getErr
			}

			setErr := tx.Set(doc, chat)
			if setErr != nil {
				return setErr
			}
			return nil
		}

		updateErr := tx.Update(doc, []firestore.Update{
			{Path: "last_message", Value: chat.LastMessage},
			{Path: "updated_at", Value: chat.UpdatedAt},
		})
		if updateErr != nil {
			return updateErr
		}

		return nil
	}

	var err error
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
