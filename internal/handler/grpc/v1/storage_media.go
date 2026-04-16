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
	storageMediaUsecase usecase.StorageMedia
	zslog               *zap.SugaredLogger
}

func (h *StorageMediaGRPC) SaveMediaID(ctx context.Context, in *v1.SaveMediaIDRequest) (*v1.SaveMediaIDResponse, error) {
	inputData := dto.StorageMediaSaveMediaIDRequest{
		MediaID:       in.MediaId,
		PhoneNumberID: in.PhoneNumberId,
	}
	data, serverError, err := h.storageMediaUsecase.SaveMediaID(ctx, inputData)
	if err != nil {
		return nil, api_response.NewGRPCErrorResponse(serverError, err)
	}
	return &v1.SaveMediaIDResponse{
		Id: data.ID,
	}, nil
}
