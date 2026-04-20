package grpc_middleware

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TimingServerInterceptor(maxTime time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// extract timestamp from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
		}
		timestamps := md.Get("x-timestamp")
		if len(timestamps) == 0 {
			return nil, status.Errorf(codes.InvalidArgument, "missing x-timestamp")
		}
		timestamp, err := time.Parse(time.RFC3339Nano, timestamps[0])
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid x-timestamp format: %v", err)
		}
		// check if timestamp is within acceptable range
		if time.Since(timestamp) > maxTime {
			return nil, status.Errorf(codes.InvalidArgument, "request timestamp is too old")
		}
		// allow some clock skew by checking if timestamp is not too far in the future
		if timestamp.After(time.Now().Add(5 * time.Second)) {
			return nil, status.Errorf(codes.InvalidArgument, "request timestamp is in the future")
		}
		return handler(ctx, req)
	}
}

func TimingClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// add timestamp to request
		currentTime := time.Now().UTC().Format(time.RFC3339Nano)
		ctx = metadata.AppendToOutgoingContext(ctx, "x-timestamp", currentTime)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
