package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/trains-io/z21.go/protocol"
)

const (
	DefaultPort     = 21105
	MaintenancePort = 21106
)

// readPollInterval bounds how long ReadPacket blocks when ctx has no deadline.
const readPollInterval = 200 * time.Millisecond

// Client talks to a Z21 command station (or compatible test server) over UDP.
type Client struct {
	conn     *net.UDPConn
	addr     *net.UDPAddr
	observer Observer
	log      *slog.Logger
	prepend  []protocol.Message
}

// Dial connects to host or host:port (default port 21105 if omitted).
func Dial(hostPort string, opts ...Option) (*Client, error) {
	return DialLocal(hostPort, 0, opts...)
}

// DialLocal connects to host or host:port, optionally binding a fixed local UDP port.
func DialLocal(hostPort string, localPort int, opts ...Option) (*Client, error) {
	hostPort, err := ensureUDPPort(hostPort, DefaultPort)
	if err != nil {
		return nil, fmt.Errorf("z21 client: resolve address: %w", err)
	}

	raddr, err := net.ResolveUDPAddr("udp", hostPort)
	if err != nil {
		return nil, fmt.Errorf("z21 client: resolve address: %w", err)
	}

	var laddr *net.UDPAddr
	if localPort > 0 {
		laddr = &net.UDPAddr{IP: net.IPv4zero, Port: localPort}
	}

	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		return nil, fmt.Errorf("z21 client: dial: %w", err)
	}

	c := &Client{conn: conn, addr: raddr}
	applyOptions(c, opts)
	return c, nil
}

// Close releases the UDP socket.
func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// Addr returns the remote address.
func (c *Client) Addr() string {
	if c == nil || c.addr == nil {
		return ""
	}
	return c.addr.String()
}

// LocalAddr returns the local UDP endpoint.
func (c *Client) LocalAddr() net.Addr {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.LocalAddr()
}

// SetPrependMessages configures datasets sent before every Send/Call payload.
func (c *Client) SetPrependMessages(msgs ...protocol.Message) {
	if c == nil {
		return
	}
	c.prepend = append([]protocol.Message(nil), msgs...)
}

// Send transmits req without waiting for a reply.
func (c *Client) Send(ctx context.Context, req protocol.Message) error {
	return c.SendMany(ctx, req)
}

// SendMany transmits reqs in one UDP datagram without waiting for a reply.
func (c *Client) SendMany(ctx context.Context, reqs ...protocol.Message) error {
	if c == nil || c.conn == nil {
		return fmt.Errorf("z21 client: not connected")
	}
	if len(reqs) == 0 {
		return fmt.Errorf("z21 client: no request")
	}

	payload, err := c.marshalOutbound(reqs)
	if err != nil {
		return err
	}

	if deadline, ok := ctx.Deadline(); ok {
		if err := c.conn.SetWriteDeadline(deadline); err != nil {
			return fmt.Errorf("z21 client: set write deadline: %w", err)
		}
		defer c.clearDeadlines()
	}

	if _, err := c.conn.Write(payload); err != nil {
		c.logDebug("z21 write failed", "addr", c.Addr(), "err", err)
		return fmt.Errorf("z21 client: write: %w", err)
	}
	c.observePacket(ctx, DirectionTX, payload, nil)
	return nil
}

// Call sends one or more requests in a single UDP datagram and waits for the
// response(s). A single request expects one UDP datagram (which may contain
// multiple datasets). Multiple requests are equivalent to separate UDP packets
// (spec §1.3); replies are collected until each request is matched or the
// context times out. Unsolicited broadcasts are ignored for completion.
func (c *Client) Call(ctx context.Context, reqs ...protocol.Message) ([]protocol.Message, error) {
	if c == nil || c.conn == nil {
		return nil, fmt.Errorf("z21 client: not connected")
	}
	if len(reqs) == 0 {
		return nil, fmt.Errorf("z21 client: no request")
	}

	payload, err := c.marshalOutbound(reqs)
	if err != nil {
		return nil, err
	}

	expecting := protocol.RequestsExpectingReplies(reqs)

	if deadline, ok := ctx.Deadline(); ok {
		if err := c.conn.SetDeadline(deadline); err != nil {
			return nil, fmt.Errorf("z21 client: set deadline: %w", err)
		}
		defer c.clearDeadlines()
	} else {
		if err := c.conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
			return nil, fmt.Errorf("z21 client: set deadline: %w", err)
		}
		defer c.clearDeadlines()
	}

	if _, err := c.conn.Write(payload); err != nil {
		c.logDebug("z21 write failed", "addr", c.Addr(), "err", err)
		return nil, fmt.Errorf("z21 client: write: %w", err)
	}
	c.observePacket(ctx, DirectionTX, payload, nil)

	if len(expecting) == 0 {
		return nil, nil
	}

	if len(expecting) == 1 && len(c.prepend) == 0 {
		msgs, err := c.readOneDatagram(ctx)
		if err != nil {
			return nil, err
		}
		return protocol.MatchedReplies(expecting, msgs), nil
	}

	// ReadPacket manages per-read deadlines; drop the write deadline on the socket.
	_ = c.conn.SetDeadline(time.Time{})
	return c.collectResponsesForRequests(ctx, expecting)
}

func (c *Client) marshalOutbound(reqs []protocol.Message) ([]byte, error) {
	outbound := make([]protocol.Message, 0, len(c.prepend)+len(reqs))
	outbound = append(outbound, c.prepend...)
	outbound = append(outbound, reqs...)

	if len(outbound) == 1 {
		return outbound[0].Marshal()
	}
	return protocol.MarshalAll(outbound...)
}

func (c *Client) readOneDatagram(ctx context.Context) ([]protocol.Message, error) {
	buf := make([]byte, 1472)
	n, err := c.conn.Read(buf)
	if err != nil {
		c.logDebug("z21 read failed", "addr", c.Addr(), "err", err)
		return nil, fmt.Errorf("z21 client: read: %w", err)
	}

	raw := buf[:n]
	msgs, err := protocol.ParseAll(raw)
	if err != nil {
		c.logDebug("z21 parse failed", "addr", c.Addr(), "err", err)
		return nil, fmt.Errorf("z21 client: parse: %w", err)
	}
	c.observePacket(ctx, DirectionRX, raw, msgs)
	return msgs, nil
}

func (c *Client) collectResponsesForRequests(ctx context.Context, reqs []protocol.Message) ([]protocol.Message, error) {
	var all []protocol.Message
	for {
		if protocol.AllRequestsAnswered(reqs, all) {
			return protocol.MatchedReplies(reqs, all), nil
		}
		if err := ctx.Err(); err != nil {
			if len(all) > 0 {
				return protocol.MatchedReplies(reqs, all), nil
			}
			return nil, err
		}

		packetCtx, cancel := context.WithTimeout(ctx, readPollInterval)
		msgs, err := c.ReadPacket(packetCtx)
		cancel()

		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				continue
			}
			if errors.Is(err, context.Canceled) {
				if len(all) > 0 {
					return protocol.MatchedReplies(reqs, all), nil
				}
				return nil, err
			}
			return protocol.MatchedReplies(reqs, all), err
		}
		all = append(all, msgs...)
	}
}

func (c *Client) clearDeadlines() {
	if c == nil || c.conn == nil {
		return
	}
	_ = c.conn.SetDeadline(time.Time{})
}

// ReadPacket waits for the next UDP datagram and parses its datasets.
// It returns when ctx is cancelled, even if no datagram arrives.
func (c *Client) ReadPacket(ctx context.Context) ([]protocol.Message, error) {
	if c == nil || c.conn == nil {
		return nil, fmt.Errorf("z21 client: not connected")
	}

	buf := make([]byte, 1472)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		deadline := time.Now().Add(readPollInterval)
		if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
			deadline = d
		}
		if err := c.conn.SetReadDeadline(deadline); err != nil {
			return nil, fmt.Errorf("z21 client: set read deadline: %w", err)
		}

		n, err := c.conn.Read(buf)
		if err != nil {
			if isReadTimeout(err) {
				continue
			}
			c.logDebug("z21 read failed", "addr", c.Addr(), "err", err)
			return nil, fmt.Errorf("z21 client: read: %w", err)
		}

		_ = c.conn.SetReadDeadline(time.Time{})
		raw := buf[:n]
		msgs, err := protocol.ParseAll(raw)
		if err != nil {
			c.logDebug("z21 parse failed", "addr", c.Addr(), "err", err)
			return nil, fmt.Errorf("z21 client: parse: %w", err)
		}
		c.observePacket(ctx, DirectionRX, raw, msgs)
		return msgs, nil
	}
}

func (c *Client) observePacket(ctx context.Context, dir Direction, raw []byte, msgs []protocol.Message) {
	if c == nil || c.observer == nil || len(raw) == 0 {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	c.observer.OnPacket(ctx, dir, c.Addr(), raw, msgs)
}

func (c *Client) logDebug(msg string, args ...any) {
	if c != nil && c.log != nil {
		c.log.Debug(msg, args...)
	}
}

func isReadTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func ensureUDPPort(hostPort string, defaultPort int) (string, error) {
	if hostPort == "" {
		return "", fmt.Errorf("empty address")
	}
	_, _, err := net.SplitHostPort(hostPort)
	if err == nil {
		return hostPort, nil
	}
	var addrErr *net.AddrError
	if errors.As(err, &addrErr) && addrErr.Err == "missing port in address" {
		return net.JoinHostPort(hostPort, strconv.Itoa(defaultPort)), nil
	}
	return "", err
}
