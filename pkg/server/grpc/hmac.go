package server_grpc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type staticHeaderCreds struct {
	signature string
}

func (c staticHeaderCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"x-signature": c.signature}, nil
}

func (c staticHeaderCreds) RequireTransportSecurity() bool { return true }

func createHMACSignature(secret string, fullMethod string, payloadMessage proto.Message) string {
	h := hmac.New(sha256.New, []byte(secret))
	options := protojson.MarshalOptions{EmitUnpopulated: true, UseProtoNames: true, Indent: " "}
	payloadBytes, _ := options.Marshal(payloadMessage)
	payloadStr := fmt.Sprintf("%s | %s", fullMethod, string(payloadBytes))
	h.Write([]byte(payloadStr))
	return hex.EncodeToString(h.Sum(nil))
}

func HMACClientInterceptor(secret string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		m, ok := req.(proto.Message)
		if !ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		signature := createHMACSignature(secret, method, m)
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
		expectedSig := createHMACSignature(secret, info.FullMethod, m)
		if !hmac.Equal([]byte(clientSig[0]), []byte(expectedSig)) {
			return nil, status.Error(codes.Unauthenticated, "invalid signature")
		}
		return handler(ctx, req)
	}
}
