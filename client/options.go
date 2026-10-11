package client

import (
	"log/slog"
	"time"
)

// Option configures a Client.
type Option func(*Client)

// WithObserver attaches a wire-level observer. Pass nil to disable.
func WithObserver(o Observer) Option {
	return func(c *Client) {
		c.observer = o
	}
}

// WithLogger attaches a logger for events not reported through return values,
// such as unsolicited datasets skipped by Call. Pass nil to disable logging.
func WithLogger(l *slog.Logger) Option {
	return func(c *Client) {
		c.log = l
	}
}

// WithCallTimeout sets how long Call waits for replies when its context has
// no deadline. Non-positive values keep the default of 5 seconds.
func WithCallTimeout(d time.Duration) Option {
	return func(c *Client) {
		if d > 0 {
			c.callTimeout = d
		}
	}
}

func applyOptions(c *Client, opts []Option) {
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
}
