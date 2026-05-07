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
	WaBusinessAccount usecase.WaBusinessAccount
	Chat              usecase.Chat
	Tenant            usecase.Tenant
	Ticket            usecase.Ticket
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
		waBusinessAccountUsecase: handler.WaBusinessAccount,
		chatUsecase:              handler.Chat,
		tenantUsecase:            handler.Tenant,
		ticketUsecase:            handler.Ticket,
		zsLog:                    handler.ZSLog,
	}
	v1.RegisterMessageServer(handler.App, r)
}
