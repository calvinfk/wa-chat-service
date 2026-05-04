package grpc_v1

import (
	v1 "wa_chat_service/docs/proto/v1"
	"wa_chat_service/internal/usecase"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type HandlerGRPCV1 struct {
	App               *grpc.Server
	StorageMedia      usecase.StorageMedia
	Message           usecase.Message
	WaBusinessAccount usecase.WaBusinessAccount
	Chat              usecase.Chat
	ZSLog             *zap.SugaredLogger
}

func NewRouterGRPCV1(handler HandlerGRPCV1) {
	storageMediaGRPC := &StorageMediaGRPC{
		storageMediaUsecase:      handler.StorageMedia,
		waBusinessAccountUsecase: handler.WaBusinessAccount,
		zsLog:                    handler.ZSLog,
	}
	v1.RegisterStorageMediaServer(handler.App, storageMediaGRPC)
	r := &MessageGRPC{
		messageUsecase:           handler.Message,
		waBusinessAccountUsecase: handler.WaBusinessAccount,
		chatUsecase:              handler.Chat,
		zsLog:                    handler.ZSLog,
	}
	v1.RegisterMessageServer(handler.App, r)
}
