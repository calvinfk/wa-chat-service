package server_grpc

import (
	"context"
	"errors"
	"net"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	pbgrpc "google.golang.org/grpc"
)

const (
	_defaultAddr = ":80"
)

// Server -.
type Server struct {
	ctx context.Context
	eg  *errgroup.Group

	App        *pbgrpc.Server
	notify     chan error
	address    string
	serverOpts []pbgrpc.ServerOption

	zlog *zap.Logger
}

// New -.
func New(zlog *zap.Logger, opts ...Option) *Server {
	group, ctx := errgroup.WithContext(context.Background())
	group.SetLimit(1)
	s := &Server{
		ctx:     ctx,
		eg:      group,
		notify:  make(chan error, 1),
		address: _defaultAddr,
		zlog:    zlog,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.App = pbgrpc.NewServer(s.serverOpts...)

	return s
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

	s.zlog.Info("grpc - Server - Started")
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
