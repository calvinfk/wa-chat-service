package grpc_v1

import (
	v1 "wa_chat_service/docs/proto/v1"
	"wa_chat_service/internal/usecase"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type HandlerGRPCV1 struct {
	App          *grpc.Server
	StorageMedia usecase.StorageMedia
	Message      usecase.Message
	ZSLog        *zap.SugaredLogger
}

func NewRouterGRPCV1(handler HandlerGRPCV1) {
	handler.newStorageMediaRoutes()
	handler.newMessageRoutes()
}

func (h *HandlerGRPCV1) newStorageMediaRoutes() {
	r := &StorageMediaGRPC{
		storageMediaUsecase: h.StorageMedia,
		zslog:               h.ZSLog,
	}
	v1.RegisterStorageMediaServer(h.App, r)
}

func (h *HandlerGRPCV1) newMessageRoutes() {
	r := &MessageGRPC{
		messageUsecase: h.Message,
		zslog:          h.ZSLog,
	}
	v1.RegisterMessageServer(h.App, r)
}
