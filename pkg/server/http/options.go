package server_http

import (
	"net"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Option -.
type Option func(*Server)

// Port -.
func Port(port string) Option {
	return func(s *Server) {
		s.address = net.JoinHostPort("", port)
	}
}

// Prefork -.
func Prefork(prefork bool) Option {
	return func(s *Server) {
		s.prefork = prefork
	}
}

func BodyLimit(limit int) Option {
	return func(s *Server) {
		s.bodyLimit = limit
	}
}

// ReadTimeout -.
func ReadTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.readTimeout = timeout
	}
}

// WriteTimeout -.
func WriteTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.writeTimeout = timeout
	}
}

// ShutdownTimeout -.
func ShutdownTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = timeout
	}
}

// StructValidator -.
func StructValidator(validator fiber.StructValidator) Option {
	return func(s *Server) {
		s.validator = validator
	}
}

func Middleware(middleware ...fiber.Handler) Option {
	return func(s *Server) {
		s.middleware = append(s.middleware, middleware...)
	}
}
