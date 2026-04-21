package handler_grpc

import (
	grpc_v1 "wa_chat_service/internal/handler/grpc/v1"
)

func NewRouter(handlerGRPCV1 grpc_v1.HandlerGRPCV1) {
	grpc_v1.NewRouterGRPCV1(handlerGRPCV1)
}
