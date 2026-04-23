package server_http

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	_defaultAddr            = ":80"
	_defaultPrefork         = false
	_defaultBodyLimit       = 16 * 1024 * 1024 // 16MB
	_defaultReadTimeout     = 5 * time.Second
	_defaultWriteTimeout    = 5 * time.Second
	_defaultShutdownTimeout = 3 * time.Second
)

// Server -.
type Server struct {
	ctx context.Context
	eg  *errgroup.Group

	App    *fiber.App
	notify chan error

	address         string
	prefork         bool
	bodyLimit       int
	readTimeout     time.Duration
	writeTimeout    time.Duration
	shutdownTimeout time.Duration

	zlog       *zap.Logger
	validator  fiber.StructValidator
	middleware []fiber.Handler
}

// New -.
func New(zlog *zap.Logger, opts ...Option) *Server {
	group, ctx := errgroup.WithContext(context.Background())
	group.SetLimit(1) // Run only one goroutine

	s := &Server{
		ctx:             ctx,
		eg:              group,
		App:             nil,
		notify:          make(chan error, 1),
		address:         _defaultAddr,
		prefork:         _defaultPrefork,
		bodyLimit:       _defaultBodyLimit,
		readTimeout:     _defaultReadTimeout,
		writeTimeout:    _defaultWriteTimeout,
		shutdownTimeout: _defaultShutdownTimeout,
		zlog:            zlog,
		validator:       nil,
		middleware:      nil,
	}

	// Custom options
	for _, opt := range opts {
		opt(s)
	}

	app := fiber.New(fiber.Config{
		StructValidator: s.validator,
		BodyLimit:       s.bodyLimit,
		ReadTimeout:     s.readTimeout,
		// WriteTimeout:    s.writeTimeout,
		JSONDecoder: json.Unmarshal,
		JSONEncoder: json.Marshal,
	})

	s.App = app

	// Register middleware
	if len(s.middleware) > 0 {
		middlewareArgs := make([]any, len(s.middleware))
		for i, m := range s.middleware {
			middlewareArgs[i] = m
		}
		s.App.Use(middlewareArgs...)
	}
	return s
}

// Start -.
func (s *Server) Start() {
	s.eg.Go(func() error {
		err := s.App.Listen(s.address, fiber.ListenConfig{
			EnablePrefork: s.prefork,
		})
		if err != nil {
			s.notify <- err

			close(s.notify)

			return err
		}

		return nil
	})

	s.zlog.Info("http - Server - Started")
}

// Notify -.
func (s *Server) Notify() <-chan error {
	return s.notify
}

// Shutdown -.
func (s *Server) Shutdown() error {
	var shutdownErrors []error

	err := s.App.ShutdownWithTimeout(s.shutdownTimeout)
	if err != nil && !errors.Is(err, context.Canceled) {
		s.zlog.Error("http - Server - Shutdown - s.App.ShutdownWithTimeout", zap.Error(err))

		shutdownErrors = append(shutdownErrors, err)
	}

	// Wait for all goroutines to finish and get any error
	err = s.eg.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		s.zlog.Error("http - Server - Shutdown - s.eg.Wait", zap.Error(err))
		shutdownErrors = append(shutdownErrors, err)
	}

	s.zlog.Info("http - Server - Shutdown")

	return errors.Join(shutdownErrors...)
}
