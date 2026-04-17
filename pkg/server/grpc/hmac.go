package server_grpc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type staticHeaderCreds struct {
	signature string
}

func (c staticHeaderCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"x-signature": c.signature}, nil
}

func (c staticHeaderCreds) RequireTransportSecurity() bool { return true }

func HMACClientInterceptor(secret string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		m, ok := req.(proto.Message)
		if !ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		// create HMAC signature
		options := protojson.MarshalOptions{EmitUnpopulated: true, UseProtoNames: true, Indent: " "}
		payload, _ := options.Marshal(m)
		h := hmac.New(sha256.New, []byte(secret))
		h.Write(payload)
		signature := hex.EncodeToString(h.Sum(nil))
		// attach signature to outgoing context
		ctx = metadata.AppendToOutgoingContext(ctx, "x-signature", signature)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func HMACServerInterceptor(secret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		clientSig := md.Get("x-signature")
		if len(clientSig) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing signature")
		}
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
		if time.Since(requestTime) > 30*time.Second || time.Until(requestTime) > 30*time.Second {
			return nil, status.Error(codes.Unauthenticated, "request expired or clock drift")
		}
		// re-calculate HMAC
		options := protojson.MarshalOptions{EmitUnpopulated: true, UseProtoNames: true, Indent: " "}
		payload, _ := options.Marshal(req.(proto.Message))
		h := hmac.New(sha256.New, []byte(secret))
		h.Write(payload)
		expectedSig := hex.EncodeToString(h.Sum(nil))
		if !hmac.Equal([]byte(clientSig[0]), []byte(expectedSig)) {
			return nil, status.Error(codes.Unauthenticated, "invalid signature")
		}
		return handler(ctx, req)
	}
}
