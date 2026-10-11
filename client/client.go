package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/trains-io/z21.go/protocol"
)

const (
	// DefaultPort is the Z21 LAN UDP port (spec §1.1).
	DefaultPort = 21105
	// MaintenancePort is the second UDP port a Z21 answers on (spec §1.1).
	MaintenancePort = 21106
)

// defaultCallTimeout bounds Call when its context has no deadline.
const defaultCallTimeout = 5 * time.Second

// readPollInterval bounds how long a single socket read blocks, so that
// ReadPacket notices context cancellation promptly.
const readPollInterval = 200 * time.Millisecond

// maxDatagramSize is the largest UDP payload that fits an Ethernet frame
// without IP fragmentation.
const maxDatagramSize = 1472

// ErrIncomplete reports that Call stopped waiting before every request was
// answered. It is returned together with the replies received so far and the
// context error that ended the wait.
var ErrIncomplete = errors.New("z21: not all requests answered")

// Client talks to a Z21 command station (or compatible test server) over UDP.
//
// A Client is not safe for concurrent use: Send, Call, ReadPacket, Collect and
// the setters share one socket and its deadlines.
type Client struct {
	conn        *net.UDPConn
	addr        *net.UDPAddr
	observer    Observer
	log         *slog.Logger
	prepend     []protocol.Message
	callTimeout time.Duration
}

// Dial connects to host or host:port (default port 21105 if omitted).
func Dial(hostPort string, opts ...Option) (*Client, error) {
	return DialLocal(hostPort, 0, opts...)
}

// DialLocal connects to host or host:port, optionally binding a fixed local UDP port.
func DialLocal(hostPort string, localPort int, opts ...Option) (*Client, error) {
	hostPort, err := ensureUDPPort(hostPort, DefaultPort)
	if err != nil {
		return nil, fmt.Errorf("z21: resolve address: %w", err)
	}

	raddr, err := net.ResolveUDPAddr("udp", hostPort)
	if err != nil {
		return nil, fmt.Errorf("z21: resolve address: %w", err)
	}

	var laddr *net.UDPAddr
	if localPort > 0 {
		laddr = &net.UDPAddr{IP: net.IPv4zero, Port: localPort}
	}

	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		return nil, fmt.Errorf("z21: dial: %w", err)
	}

	c := &Client{conn: conn, addr: raddr, callTimeout: defaultCallTimeout}
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
		return errors.New("z21: not connected")
	}
	if len(reqs) == 0 {
		return errors.New("z21: no request")
	}
	return c.writeOutbound(ctx, reqs)
}

// Call sends one or more requests in a single UDP datagram and waits until
// every request that expects a reply is answered. Replies may arrive in one or
// several datagrams (spec §1.3); unsolicited datasets such as broadcasts are
// skipped. The matched replies are returned in request order.
//
// If ctx has no deadline, Call gives up after the client's call timeout
// (5 seconds unless changed with WithCallTimeout). When the wait ends
// before every request is answered, Call returns the replies received so far
// and an error that matches both ErrIncomplete and the context error
// (context.DeadlineExceeded or context.Canceled).
func (c *Client) Call(ctx context.Context, reqs ...protocol.Message) ([]protocol.Message, error) {
	if c == nil || c.conn == nil {
		return nil, errors.New("z21: not connected")
	}
	if len(reqs) == 0 {
		return nil, errors.New("z21: no request")
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.callTimeout)
		defer cancel()
	}

	if err := c.writeOutbound(ctx, reqs); err != nil {
		return nil, err
	}

	expecting := protocol.RequestsExpectingReplies(reqs)
	if len(expecting) == 0 {
		return nil, nil
	}
	return c.awaitReplies(ctx, expecting)
}

func (c *Client) writeOutbound(ctx context.Context, reqs []protocol.Message) error {
	payload, err := c.marshalOutbound(reqs)
	if err != nil {
		return err
	}

	if deadline, ok := ctx.Deadline(); ok {
		if err := c.conn.SetWriteDeadline(deadline); err != nil {
			return fmt.Errorf("z21: set write deadline: %w", err)
		}
		defer func() { _ = c.conn.SetWriteDeadline(time.Time{}) }()
	}

	if _, err := c.conn.Write(payload); err != nil {
		return fmt.Errorf("z21: write: %w", err)
	}
	c.observePacket(ctx, DirectionTX, payload, nil)
	return nil
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

func (c *Client) awaitReplies(ctx context.Context, reqs []protocol.Message) ([]protocol.Message, error) {
	var all []protocol.Message
	for {
		msgs, err := c.ReadPacket(ctx)
		if err != nil {
			matched := protocol.MatchedReplies(reqs, all)
			if ctxErr := ctx.Err(); ctxErr != nil {
				return matched, fmt.Errorf("%w: %w", ErrIncomplete, ctxErr)
			}
			return matched, err
		}
		all = append(all, msgs...)

		matched := protocol.MatchedReplies(reqs, all)
		if len(matched) == len(reqs) {
			c.logIgnored(all, len(all)-len(matched))
			return matched, nil
		}
	}
}

// ReadPacket waits for the next UDP datagram and parses its datasets.
// It returns ctx.Err() when ctx is done, even if no datagram arrives.
func (c *Client) ReadPacket(ctx context.Context) ([]protocol.Message, error) {
	if c == nil || c.conn == nil {
		return nil, errors.New("z21: not connected")
	}
	defer func() { _ = c.conn.SetReadDeadline(time.Time{}) }()

	buf := make([]byte, maxDatagramSize)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		deadline := time.Now().Add(readPollInterval)
		if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
			deadline = d
		}
		if err := c.conn.SetReadDeadline(deadline); err != nil {
			return nil, fmt.Errorf("z21: set read deadline: %w", err)
		}

		n, err := c.conn.Read(buf)
		if err != nil {
			if isReadTimeout(err) {
				continue
			}
			return nil, fmt.Errorf("z21: read: %w", err)
		}

		raw := buf[:n]
		msgs, err := protocol.ParseAll(raw)
		if err != nil {
			return nil, fmt.Errorf("z21: parse: %w", err)
		}
		c.observePacket(ctx, DirectionRX, raw, msgs)
		return msgs, nil
	}
}

func (c *Client) observePacket(ctx context.Context, dir Direction, raw []byte, msgs []protocol.Message) {
	if c.observer == nil || len(raw) == 0 {
		return
	}
	c.observer.OnPacket(ctx, dir, c.Addr(), raw, msgs)
}

func (c *Client) logIgnored(all []protocol.Message, ignored int) {
	if c.log == nil || ignored == 0 {
		return
	}
	c.log.Debug("ignore unmatched datasets", "addr", c.Addr(), "count", ignored, "received", len(all))
}

func isReadTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

// ensureUDPPort appends defaultPort to hostPort unless it already has a port.
// It accepts host names, IPv4 addresses and IPv6 addresses with or without
// brackets.
func ensureUDPPort(hostPort string, defaultPort int) (string, error) {
	if hostPort == "" {
		return "", errors.New("empty address")
	}
	_, _, err := net.SplitHostPort(hostPort)
	if err == nil {
		return hostPort, nil
	}

	host := hostPort
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}
	if !strings.Contains(host, ":") || net.ParseIP(host) != nil {
		return net.JoinHostPort(host, strconv.Itoa(defaultPort)), nil
	}
	return "", err
}
