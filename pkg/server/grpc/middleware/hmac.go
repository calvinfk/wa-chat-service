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

// createHMACSignature - generates an HMAC signature for a given gRPC method and its associated protobuf message using a secret key.
// The signature is created by hashing the method name and the serialized protobuf message together
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

// HMACClientInterceptor - returns a gRPC unary client interceptor that adds an HMAC signature to the outgoing request metadata for authentication purposes.
// The interceptor calculates the HMAC signature based on the method name and the protobuf message being sent, using a provided secret key.
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

// HMACServerInterceptor - returns a gRPC unary server interceptor that validates the HMAC signature of incoming requests for authentication purposes.
// The interceptor calculates the expected HMAC signature based on the method name and the protobuf message, using a provided secret key, and compares it with the received signature.
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
