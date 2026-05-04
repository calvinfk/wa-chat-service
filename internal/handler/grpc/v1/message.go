package grpc_v1

import (
	"context"
	"time"
	v1 "wa_chat_service/docs/proto/v1"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/api_response"
	"wa_chat_service/pkg/utils"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MessageGRPC struct {
	v1.UnimplementedMessageServer
	messageUsecase           usecase.Message
	waBusinessAccountUsecase usecase.WaBusinessAccount
	chatUsecase              usecase.Chat
	zsLog                    *zap.SugaredLogger
}

func (h *MessageGRPC) SaveMessage(ctx context.Context, req *v1.SaveMessageRequest) (*emptypb.Empty, error) {
	message := req.GetMessage()
	phoneNumberId := req.GetPhoneNumberId()
	recipientId := req.GetRecipientId()
	var userLastMessageAt *time.Time
	if req.GetUserLastMessageAt() != nil {
		createdAt := req.GetUserLastMessageAt().AsTime()
		userLastMessageAt = &createdAt
	}
	// Map gRPC request to DTO for use case
	inputData := dto.MessageSaveRequest{
		ID:                &message.Id,
		Wamid:             message.Wamid,
		PhoneNumberId:     phoneNumberId,
		RecipientId:       recipientId,
		RecipientName:     req.GetRecipientName(),
		LastMessage:       req.GetLastMessage(),
		UserLastMessageAt: userLastMessageAt,
		MessageType:       message.MessageType,
		MessageCategory:   message.MessageCategory,
		SenderName:        message.SenderName,
		Payload:           message.Payload,
		StorageMediaID:    message.StorageMediaId,
		Status:            message.Status,
		Error:             message.Error,
		CreatedAt:         message.CreatedAt.AsTime(),
	}
	if message.SentAt != nil {
		timeAt := message.SentAt.AsTime()
		inputData.SentAt = &timeAt
	}

	if message.DeliveredAt != nil {
		timeAt := message.DeliveredAt.AsTime()
		inputData.DeliveredAt = &timeAt
	}
	if message.ReadAt != nil {
		timeAt := message.ReadAt.AsTime()
		inputData.ReadAt = &timeAt
	}
	validator := utils.NewValidator()
	if err := validator.Struct(inputData); err != nil {
		return nil, api_response.NewGRPCErrorResponse(false, err)
	}
	whatsappBusinessAccount, serverError, err := h.waBusinessAccountUsecase.GetByPhoneNumberId(ctx, phoneNumberId)
	if err != nil {
		return nil, api_response.NewGRPCErrorResponse(serverError, err)
	}
	serverError, err = h.messageUsecase.SaveMessage(ctx, whatsappBusinessAccount.TenantID, inputData)
	if err != nil {
		return nil, api_response.NewGRPCErrorResponse(serverError, err)
	}
	return &emptypb.Empty{}, nil
}

func (h *MessageGRPC) UpdateMessageStatus(ctx context.Context, req *v1.UpdateMessageStatusRequest) (*emptypb.Empty, error) {
	whatsappBusinessAccount, serverError, err := h.waBusinessAccountUsecase.GetByPhoneNumberId(ctx, req.GetPhoneNumberId())
	if err != nil {
		return nil, api_response.NewGRPCErrorResponse(serverError, err)
	}
	message, serverError, err := h.messageUsecase.GetByWamid(ctx, whatsappBusinessAccount.TenantID, req.GetPhoneNumberId(), req.GetRecipientId(), req.GetWamid())
	if err != nil {
		return nil, api_response.NewGRPCErrorResponse(serverError, err)
	}
	// Map gRPC request to DTO for use case
	inputData := dto.MessageSaveRequest{
		ID:              &message.DocumentID,
		ChatID:          &message.ChatID,
		StorageMediaID:  message.StorageMediaID,
		Wamid:           message.Wamid,
		PhoneNumberId:   req.PhoneNumberId,
		RecipientId:     req.RecipientId,
		MessageType:     message.MessageType,
		MessageCategory: req.MessageCategory,
		SenderName:      message.SenderName,
		Payload:         message.Payload,
		Status:          req.Status,
		Error:           req.Error,
		SentAt:          message.SentAt,
		DeliveredAt:     message.DeliveredAt,
		ReadAt:          message.ReadAt,
		CreatedAt:       req.Timestamp.AsTime(),
	}
	switch req.GetStatus() {
	case "sent":
		inputData.SentAt = &inputData.CreatedAt
	case "delivered":
		inputData.DeliveredAt = &inputData.CreatedAt
	case "read":
		inputData.ReadAt = &inputData.CreatedAt
	}
	validator := utils.NewValidator()
	if err := validator.Struct(inputData); err != nil {
		h.zsLog.Errorf("[UpdateMessageStatus] Validation error: %v", err)
		return nil, api_response.NewGRPCErrorResponse(false, err)
	}
	serverError, err = h.messageUsecase.SaveMessage(ctx, whatsappBusinessAccount.TenantID, inputData)
	if err != nil {
		h.zsLog.Errorf("[UpdateMessageStatus] Failed to update message status: %v", err)
		return nil, api_response.NewGRPCErrorResponse(serverError, err)
	}
	return &emptypb.Empty{}, nil
}
