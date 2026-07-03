package client

import "log/slog"

// Option configures a Client.
type Option func(*Client)

// WithObserver attaches a wire-level observer. Pass nil to disable.
func WithObserver(o Observer) Option {
	return func(c *Client) {
		c.observer = o
	}
}

// WithLogger attaches a logger for operational messages (errors, timeouts).
// Debug-level events are suppressed when nil.
func WithLogger(l *slog.Logger) Option {
	return func(c *Client) {
		c.log = l
	}
}

func applyOptions(c *Client, opts []Option) {
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
}
