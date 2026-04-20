package grpc_middleware

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TimingServerInterceptor(maxTime time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// get timestamp from request
		type TimeStamper interface{ GetGrpcCreatedAt() *timestamppb.Timestamp }
		ts, ok := req.(TimeStamper)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "request missing security timestamp")
		}
		grpcCreatedAt := ts.GetGrpcCreatedAt()
		if grpcCreatedAt == nil {
			return nil, status.Error(codes.InvalidArgument, "missing grpc_created_at")
		}
		if err := grpcCreatedAt.CheckValid(); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid grpc_created_at")
		}
		// check if the request is too old
		requestTime := grpcCreatedAt.AsTime()
		if time.Since(requestTime) > maxTime || time.Until(requestTime) > maxTime {
			return nil, status.Error(codes.Unauthenticated, "request expired or clock drift")
		}
		return handler(ctx, req)
	}
}
