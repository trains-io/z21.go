package server

import "log/slog"

// Option configures a Server.
type Option func(*Server)

// WithHandler sets the handler for incoming requests. Without a handler, requests
// are read and dropped.
func WithHandler(h Handler) Option {
	return func(s *Server) {
		s.handler = h
	}
}

// WithLogger attaches a logger for operational messages (malformed packets, write
// errors). The server is silent when nil.
func WithLogger(l *slog.Logger) Option {
	return func(s *Server) {
		s.log = l
	}
}

func applyOptions(s *Server, opts []Option) {
	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}
}
