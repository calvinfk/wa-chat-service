package grpc_v1

import (
	"context"
	v1 "wa_chat_service/docs/proto/v1"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/api_response"

	"go.uber.org/zap"
)

type StorageMediaGRPC struct {
	v1.UnimplementedStorageMediaServer
	storageMediaUsecase      usecase.StorageMedia
	waBusinessAccountUsecase usecase.WaBusinessAccount
	zsLog                    *zap.SugaredLogger
}

func (h *StorageMediaGRPC) SaveMediaID(ctx context.Context, in *v1.SaveMediaIDRequest) (*v1.SaveMediaIDResponse, error) {
	// Map gRPC request to DTO for use case
	inputData := dto.StorageMediaSaveMediaIDRequest{
		MediaId:       in.MediaId,
		PhoneNumberId: in.PhoneNumberId,
	}
	whatsappBusinessAccount, serverError, err := h.waBusinessAccountUsecase.GetByPhoneNumberId(ctx, in.PhoneNumberId)
	if err != nil {
		return nil, api_response.NewGRPCErrorResponse(serverError, err)
	}
	data, serverError, err := h.storageMediaUsecase.SaveMediaID(ctx, whatsappBusinessAccount.TenantID, inputData)
	if err != nil {
		return nil, api_response.NewGRPCErrorResponse(serverError, err)
	}
	return &v1.SaveMediaIDResponse{
		Id: data.ID,
	}, nil
}
