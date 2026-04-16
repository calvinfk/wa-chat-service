package handler_grpc

import (
	grpc_v1 "wa_chat_service/internal/handler/grpc/v1"

	"google.golang.org/grpc"
)

func NewRouter(app *grpc.Server, handlerGRPCV1 grpc_v1.HandlerGRPCV1) {
	{
		grpc_v1.NewStorageMediaRoutes(app, handlerGRPCV1)
	}
}
