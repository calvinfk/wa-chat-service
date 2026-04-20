package grpc_middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type staticHeaderCreds struct {
	signature string
}

func (c staticHeaderCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"x-signature": c.signature}, nil
}

func (c staticHeaderCreds) RequireTransportSecurity() bool { return true }

func createHMACSignature(secret, fullMethod string, msg proto.Message) (string, error) {
	if msg == nil {
		return "", errors.New("nil proto message")
	}

	b, err := (proto.MarshalOptions{Deterministic: true}).Marshal(msg)
	if err != nil {
		return "", err
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(fullMethod))
	h.Write([]byte{0}) // stable delimiter
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func HMACClientInterceptor(secret string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		m, ok := req.(proto.Message)
		if !ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		signature, err := createHMACSignature(secret, method, m)
		if err != nil {
			return status.Error(codes.Internal, "failed to create HMAC signature")
		}
		ctx = metadata.AppendToOutgoingContext(ctx, "x-signature", signature)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func HMACServerInterceptor(secret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		m, ok := req.(proto.Message)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "request is not a proto message")
		}
		md, _ := metadata.FromIncomingContext(ctx)
		clientSig := md.Get("x-signature")
		if len(clientSig) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing signature")
		}
		// re-calculate HMAC
		expectedSig, err := createHMACSignature(secret, info.FullMethod, m)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to create HMAC signature")
		}
		if !hmac.Equal([]byte(clientSig[0]), []byte(expectedSig)) {
			return nil, status.Error(codes.Unauthenticated, "invalid signature")
		}
		return handler(ctx, req)
	}
}
