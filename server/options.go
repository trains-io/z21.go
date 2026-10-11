package server

import (
	"log/slog"
	"time"
)

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

// WithObserver attaches an observer for every packet received and sent, for
// example NewTracer(os.Stderr). Pass nil to disable.
func WithObserver(o Observer) Option {
	return func(s *Server) {
		s.observer = o
	}
}

// WithSessionTimeout sets how long a client may stay silent before the server
// forgets its session (default DefaultSessionTimeout). Zero or negative keeps
// sessions until LAN_LOGOFF.
func WithSessionTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.sessionTimeout = d
	}
}

// WithAdditionalAddr makes New also bind a socket on addr, for example
// ":21106" (AlternatePort). All sockets share the same client sessions; a
// client gets replies and broadcasts on the socket it last sent to. Repeat the
// option for more addresses.
func WithAdditionalAddr(addr string) Option {
	return func(s *Server) {
		s.additionalAddrs = append(s.additionalAddrs, addr)
	}
}

func applyOptions(s *Server, opts []Option) {
	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}
}
