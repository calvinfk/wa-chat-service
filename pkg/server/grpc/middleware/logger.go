package grpc_middleware

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func UnaryRequestLogger() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		peerAddr := "-"
		if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
			peerAddr = p.Addr.String()
			if host, _, splitErr := net.SplitHostPort(peerAddr); splitErr == nil {
				peerAddr = host
			}
		}
		codeStr := status.Code(err).String()
		currentTime := time.Now().Format("15:04:05")
		timeTakenStr := time.Since(start).String()
		fmt.Printf("%s | %-3s | %13s | %15s | %-7s | %s\n",
			currentTime,
			codeStr,
			timeTakenStr,
			peerAddr,
			"gRPC",
			info.FullMethod,
		)

		return resp, err
	}
}
