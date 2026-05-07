package grpc_v1

import (
	"context"
	"time"
	v1 "wa_chat_service/docs/proto/v1"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/api_response"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/utils"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MessageGRPC struct {
	v1.UnimplementedMessageServer
	waBusinessAccountUsecase usecase.WaBusinessAccount
	tenantUsecase            usecase.Tenant
	chatUsecase              usecase.Chat
	ticketUsecase            usecase.Ticket
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
	serverError, err = h.chatUsecase.SaveMessage(ctx, whatsappBusinessAccount.TenantID, inputData)
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
	tenant, serverError, err := h.tenantUsecase.GetByID(ctx, whatsappBusinessAccount.TenantID)
	if err != nil {
		return nil, api_response.NewGRPCErrorResponse(serverError, err)
	}
	// Map gRPC request to DTO for use case
	inputData := dto.MessageSaveRequest{
		PhoneNumberId:   req.PhoneNumberId,
		RecipientId:     req.RecipientId,
		MessageCategory: req.MessageCategory,
		Status:          req.Status,
		Error:           req.Error,
		CreatedAt:       req.Timestamp.AsTime(),
	}
	// if tenant chat type is ticket, try to get message from ticket message collection, if not found then get from chat message collection.
	// if tenant chat type is not ticket, get message from chat message collection.
	if tenant.ChatType == model.TenantChatTypeTicket {
		ticketMessage, serverError, err := h.ticketUsecase.GetTicketMessageByWamid(ctx, whatsappBusinessAccount.TenantID, req.GetPhoneNumberId(), req.GetRecipientId(), req.GetWamid())
		if err != nil {
			if serverError {
				h.zsLog.Errorf("[UpdateMessageStatus] Failed to get ticket message by WAMID: %v", err)
				return nil, api_response.NewGRPCErrorResponse(serverError, err)
			}
			message, serverError, err := h.chatUsecase.GetMessageByWamid(ctx, whatsappBusinessAccount.TenantID, req.GetPhoneNumberId(), req.GetRecipientId(), req.GetWamid())
			if err != nil {
				h.zsLog.Errorf("[UpdateMessageStatus] Failed to get message by WAMID: %v", err)
				return nil, api_response.NewGRPCErrorResponse(serverError, err)
			}
			inputData.ID = &message.DocumentID
			inputData.ChatID = &message.ChatID
			inputData.StorageMediaID = message.StorageMediaID
			inputData.Wamid = message.Wamid
			inputData.MessageType = message.MessageType
			inputData.SenderName = message.SenderName
			inputData.Payload = message.Payload
			inputData.SentAt = message.SentAt
			inputData.DeliveredAt = message.DeliveredAt
			inputData.ReadAt = message.ReadAt
		} else {
			inputData.ID = &ticketMessage.DocumentID
			inputData.TicketID = &ticketMessage.TicketID
			inputData.StorageMediaID = ticketMessage.StorageMediaID
			inputData.Wamid = ticketMessage.Wamid
			inputData.MessageType = ticketMessage.MessageType
			inputData.SenderName = ticketMessage.SenderName
			inputData.Payload = ticketMessage.Payload
			inputData.SentAt = ticketMessage.SentAt
			inputData.DeliveredAt = ticketMessage.DeliveredAt
			inputData.ReadAt = ticketMessage.ReadAt
		}
	} else {
		message, serverError, err := h.chatUsecase.GetMessageByWamid(ctx, whatsappBusinessAccount.TenantID, req.GetPhoneNumberId(), req.GetRecipientId(), req.GetWamid())
		if err != nil {
			h.zsLog.Errorf("[UpdateMessageStatus] Failed to get message by WAMID: %v", err)
			return nil, api_response.NewGRPCErrorResponse(serverError, err)
		}
		inputData.ID = &message.DocumentID
		inputData.ChatID = &message.ChatID
		inputData.StorageMediaID = message.StorageMediaID
		inputData.Wamid = message.Wamid
		inputData.MessageType = message.MessageType
		inputData.SenderName = message.SenderName
		inputData.Payload = message.Payload
		inputData.SentAt = message.SentAt
		inputData.DeliveredAt = message.DeliveredAt
		inputData.ReadAt = message.ReadAt
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
		return nil, api_response.NewGRPCErrorResponse(false, errs.ErrGenericInvalidBody)
	}
	serverError, err = h.chatUsecase.SaveMessage(ctx, whatsappBusinessAccount.TenantID, inputData)
	if err != nil {
		h.zsLog.Errorf("[UpdateMessageStatus] Failed to update message status: %v", err)
		return nil, api_response.NewGRPCErrorResponse(serverError, err)
	}
	return &emptypb.Empty{}, nil
}
