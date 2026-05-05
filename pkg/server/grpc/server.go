package server_grpc

import (
	"context"
	"errors"
	"net"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const (
	_defaultAddr = ":80"
)

// Server - represents a gRPC server instance.
type Server struct {
	ctx context.Context
	eg  *errgroup.Group

	App        *grpc.Server
	notify     chan error
	address    string
	serverOpts []grpc.ServerOption

	zlog *zap.Logger
}

// New - creates a new instance of the Server struct, initializing its fields and applying any provided options.
// It sets up the gRPC server with the specified options and prepares it for handling incoming requests.
// The server will listen on the specified address and log its activities using the provided zap.Logger instance.
func New(zlog *zap.Logger, opts ...Option) *Server {
	// Stop the server if any of the goroutines returns an error, or if the context is canceled.
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

	s.App = grpc.NewServer(s.serverOpts...)

	return s
}

// Start - starts the gRPC server in a separate goroutine, allowing it to handle incoming requests concurrently while the main program continues to execute.
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

// Notify - returns a channel that will receive an error if the server encounters an issue during startup or runtime.
// The channel will be closed after sending the error, ensuring that it can only be read once.
func (s *Server) Notify() <-chan error {
	return s.notify
}

// Shutdown - shuts down the gRPC server gracefully, stopping it from accepting new requests and waiting for existing requests to complete.
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
