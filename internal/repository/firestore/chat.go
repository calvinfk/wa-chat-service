package repository_firestore

import (
	"context"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
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

func (r *ChatRepository) Upsert(ctx context.Context, tx *firestore.Transaction, chat model.Chat) (model.Chat, bool, error) {
	created := false
	docRef := r.db.Collection(r.chat.TableName()).Doc(chat.DocumentID)
	_, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return r.chat, created, err
		}
		created = true
	}
	updates := []firestore.Update{
		{Path: "chat_status", Value: chat.ChatStatus},
		{Path: "agent_id", Value: chat.AgentID},
		{Path: "last_message", Value: chat.LastMessage},
		{Path: "user_last_message_at", Value: chat.UserLastMessageAt},
		{Path: "updated_at", Value: chat.UpdatedAt},
	}
	if tx != nil {
		if created {
			err = tx.Set(docRef, chat)
		} else {
			err = tx.Update(docRef, updates)
		}
	} else {
		if created {
			_, err = docRef.Set(ctx, chat)
		} else {
			_, err = docRef.Update(ctx, updates)
		}
	}
	return chat, created, err
}

func (r *ChatRepository) GetChatByPhoneNumberId(ctx context.Context, filter filter_request.FilterRequest[dto.ChatGetByPhoneNumberIdRequest]) (filter_request.FilterResponse[dto.ChatGetByPhoneNumberIdResponse], error) {
	var response filter_request.FilterResponse[dto.ChatGetByPhoneNumberIdResponse]
	filters, sort, paginate, err := filter_request.InitializeFilter(filter, r.chat.AllowedFilterFields(), r.chat.AllowedSortFields())
	if err != nil {
		return response, err
	}
	query := r.db.Collection(r.chat.TableName()).Query
	docs, totalData, err := filter_request.ApplyFilterFirestore(ctx, query, filters, sort, paginate)
	if err != nil {
		return response, err
	}
	var result []dto.ChatGetByPhoneNumberIdResponse
	for _, doc := range docs {
		var chat model.Chat
		docData := doc.Data()
		docData["id"] = doc.Ref.ID
		err := utils.MapToStruct(docData, &chat)
		if err != nil {
			return response, err
		}
		result = append(result, dto.ChatGetByPhoneNumberIdResponse{}.FromModel(chat))
	}
	response = filter_request.NewFilterResponse(result, paginate, totalData)
	return response, nil
}

func (r *ChatRepository) GetRunningTicketChat(ctx context.Context, phoneNumberId string, recipientId string) (model.Chat, error) {
	doc, err := r.db.Collection(r.chat.TableName()).
		Where("phone_number_id", "==", phoneNumberId).
		Where("recipient_id", "==", recipientId).
		Where("chat_type", "==", "ticket").
		Where("chat_status", "in", []model.ChatStatus{model.ChatStatusOpen, model.ChatStatusInProgress}).
		OrderBy("created_at", firestore.Desc).
		Limit(1).Documents(ctx).Next()
	if err != nil {
		if err == iterator.Done {
			return r.chat, errs.ErrGenericNotFound
		}
		return r.chat, err
	}
	var chat model.Chat
	docData := doc.Data()
	docData["id"] = doc.Ref.ID
	err = utils.MapToStruct(docData, &chat)
	if err != nil {
		return r.chat, err
	}
	return chat, nil
}

func (r *ChatRepository) GetByID(ctx context.Context, chatID string) (model.Chat, error) {
	doc, err := r.db.Collection(r.chat.TableName()).Doc(chatID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return r.chat, errs.ErrGenericNotFound
		}
		return r.chat, err
	}
	var chat model.Chat
	docData := doc.Data()
	docData["id"] = doc.Ref.ID
	err = utils.MapToStruct(docData, &chat)
	if err != nil {
		return r.chat, err
	}
	return chat, nil
}

func (r *ChatRepository) GetChatTicketDataAnalytics(ctx context.Context, phoneNumberIds []string, startTime time.Time, endTime time.Time) ([]model.Chat, error) {
	var chats []model.Chat
	iter := r.db.Collection(r.chat.TableName()).
		Where("phone_number_id", "in", phoneNumberIds).
		Where("created_at", ">=", startTime).
		Where("created_at", "<=", endTime).
		OrderBy("created_at", firestore.Desc).
		Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var chat model.Chat
		docData := doc.Data()
		docData["id"] = doc.Ref.ID
		err = utils.MapToStruct(docData, &chat)
		if err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}
	return chats, nil
}
