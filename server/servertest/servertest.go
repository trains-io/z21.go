// Package servertest provides an in-process Z21 server for tests, in the style
// of net/http/httptest.
package servertest

import (
	"context"
	"fmt"
	"sync"

	"github.com/trains-io/z21.go/client"
	"github.com/trains-io/z21.go/server"
)

// Server is a Z21 server listening on a loopback UDP port.
type Server struct {
	// Addr is the server address as "127.0.0.1:port", ready for client.Dial.
	Addr string

	// Server is the running server, for example to call Broadcast or
	// BroadcastLocoInfo from a test.
	Server *server.Server

	cancel    context.CancelFunc
	done      chan struct{}
	closeOnce sync.Once
}

// NewServer starts a server with handler h on 127.0.0.1 and a free port. The
// caller must call Close when finished. It panics if the server cannot start.
func NewServer(h server.Handler, opts ...server.Option) *Server {
	opts = append([]server.Option{server.WithHandler(h)}, opts...)
	srv, err := server.New("127.0.0.1:0", opts...)
	if err != nil {
		panic(fmt.Sprintf("servertest: failed to start server: %v", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := &Server{
		Addr:   srv.Addr().String(),
		Server: srv,
		cancel: cancel,
		done:   make(chan struct{}),
	}
	go func() {
		defer close(s.done)
		_ = srv.Serve(ctx)
	}()
	return s
}

// Dial returns a client connected to the server.
func (s *Server) Dial(opts ...client.Option) (*client.Client, error) {
	return client.Dial(s.Addr, opts...)
}

// Close stops the server and waits until it has stopped. It is safe to call
// more than once.
func (s *Server) Close() {
	s.closeOnce.Do(func() {
		s.cancel()
		<-s.done
		_ = s.Server.Close()
	})
}
