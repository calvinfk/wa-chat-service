package grpc_v1

import (
	v1 "wa_chat_service/docs/proto/v1"
	"wa_chat_service/internal/usecase"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type HandlerGRPCV1 struct {
	StorageMedia usecase.StorageMedia
	ZSLog        *zap.SugaredLogger
}

func NewStorageMediaRoutes(app *grpc.Server, handler HandlerGRPCV1) {
	r := &StorageMediaGRPC{
		storageMediaUsecase: handler.StorageMedia,
		zslog:               handler.ZSLog,
	}
	v1.RegisterStorageMediaServer(app, r)
}
