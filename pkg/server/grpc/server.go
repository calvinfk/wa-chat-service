package server_grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

const (
	_defaultAddr = ":80"
	_defaultLog  = true
)

// Server -.
type Server struct {
	ctx context.Context
	eg  *errgroup.Group

	App        *grpc.Server
	notify     chan error
	address    string
	enableLog  bool
	serverOpts []grpc.ServerOption

	zlog *zap.Logger
}

// New -.
func New(zlog *zap.Logger, opts ...Option) *Server {
	group, ctx := errgroup.WithContext(context.Background())
	group.SetLimit(1)
	s := &Server{
		ctx:       ctx,
		eg:        group,
		notify:    make(chan error, 1),
		address:   _defaultAddr,
		enableLog: _defaultLog,
		zlog:      zlog,
	}

	for _, opt := range opts {
		opt(s)
	}

	serverOpts := append([]grpc.ServerOption{}, s.serverOpts...)
	if s.enableLog {
		serverOpts = append(serverOpts, grpc.ChainUnaryInterceptor(s.unaryRequestLogger()))
	}

	s.App = grpc.NewServer(serverOpts...)

	return s
}

func (s *Server) unaryRequestLogger() grpc.UnaryServerInterceptor {
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
		fmt.Printf("%s | %-3s | %13s | %15s | %-7s | %s\n\n",
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

// Start -.
func (s *Server) Start() {
	s.eg.Go(func() error {
		var lc net.ListenConfig

		ln, err := lc.Listen(s.ctx, "tcp", s.address)
		if err != nil {
			s.notify <- err

			close(s.notify)

			return err
		}

		err = s.App.Serve(ln)
		if err != nil {
			s.notify <- err

			close(s.notify)

			return err
		}

		return nil
	})

	s.zlog.Info("grpc - Server - Started on" + s.address)
}

// Notify -.
func (s *Server) Notify() <-chan error {
	return s.notify
}

// Shutdown -.
func (s *Server) Shutdown() error {
	var shutdownErrors []error

	s.App.GracefulStop()

	err := s.eg.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		s.zlog.Error("grpc - Server - Shutdown - s.eg.Wait", zap.Error(err))
		shutdownErrors = append(shutdownErrors, err)
	}

	s.zlog.Info("grpc - Server - Shutdown")

	return errors.Join(shutdownErrors...)
}
