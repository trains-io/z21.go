package z21

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Default Constants
const (
	DefaultName    = "main"
	DefaultURL     = "127.0.0.1:21105"
	DefaultPort    = 21105
	DefaultTimeout = 5 * time.Second
)

const (
	defaultBufSize      = 1472
	defaultEventBufSize = 500
	defaultPortString   = "21105"
)

var (
	ErrBadPacket         = errors.New("z21: invalid packet")
	ErrInvalidConnection = errors.New("z21: invalid connection")
)

type Option func(*Options) error

type CustomDialer interface {
	Dial(network, address string) (net.Conn, error)
}

type Options struct {
	// Url represents a Z21 control station url to which the client
	// will be connecting.
	Url          string
	Verbose      bool
	CustomDialer CustomDialer
	Timeout      time.Duration
	Logger       zerolog.Logger
}

type Response struct {
	Message Serializable
	Err     error
}

type requestEntry struct {
	key      string
	response chan Response
	timer    *time.Timer
}

type Conn struct {
	mu       sync.Mutex
	Opts     Options
	conn     net.Conn
	br       *z21Reader
	bw       *z21Writer
	requests map[string]*requestEntry
	events   chan Serializable
	done     chan struct{}
}

type z21Reader struct {
	r   io.Reader
	buf []byte
}

type z21Writer struct {
	w io.Writer
}

func GetDefaultOptions() Options {
	return Options{
		Verbose: false,
		Timeout: DefaultTimeout,
	}
}

func Timeout(t time.Duration) Option {
	return func(o *Options) error {
		o.Timeout = t
		return nil
	}
}

func SetCustomDialer(dialer CustomDialer) Option {
	return func(o *Options) error {
		o.CustomDialer = dialer
		return nil
	}
}

func Verbose(v bool) Option {
	return func(o *Options) error {
		o.Verbose = v
		return nil
	}
}

func Connect(url string, options ...Option) (*Conn, error) {
	opts := GetDefaultOptions()
	opts.Url = processUrlString(url)
	for _, opt := range options {
		if opt != nil {
			if err := opt(&opts); err != nil {
				return nil, err
			}
		}
	}

	initLogger(&opts)
	return opts.Connect()
}

func initLogger(o *Options) {
	if o.Verbose {
		o.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
			Level(zerolog.DebugLevel).
			With().
			Str("component", "z21lib").
			Timestamp().
			Logger()
	}
}

func processUrlString(url string) string {
	host, port, err := net.SplitHostPort(url)
	if err != nil {
		host = url // assume no port specified
	}

	if port == "" {
		port = defaultPortString
	}

	return net.JoinHostPort(host, port)
}

func (o Options) Connect() (*Conn, error) {
	nc := &Conn{
		Opts:     o,
		requests: make(map[string]*requestEntry),
		done:     make(chan struct{}),
	}

	nc.newReaderWriter()

	if err := nc.connect(); err != nil {
		return nil, err
	}

	nc.events = make(chan Serializable, defaultEventBufSize)
	go nc.Listen()

	return nc, nil
}

func (nc *Conn) connect() error {
	var err error

	dialer := nc.Opts.CustomDialer
	if dialer == nil {
		dialer = &net.Dialer{}
	}
	nc.conn, err = dialer.Dial("udp", nc.Opts.Url)
	nc.Opts.Logger.Debug().
		Str("from", nc.conn.LocalAddr().String()).
		Str("to", nc.Opts.Url).
		Str("status", "connected").
		Msg("Z21 conn")
	if err != nil {
		return err
	}

	nc.bindToNewConn()

	return nil
}

func (nc *Conn) GetLocalPort() net.Addr {
	return nc.conn.LocalAddr()
}

func (nc *Conn) newReaderWriter() {
	nc.bw = &z21Writer{}
	nc.br = &z21Reader{
		buf: make([]byte, defaultBufSize),
	}
}

func (nc *Conn) bindToNewConn() {
	bw := nc.bw
	bw.w = nc.newWriter()
	br := nc.br
	br.r = nc.conn
}

func (nc *Conn) newWriter() io.Writer {
	var w io.Writer = nc.conn

	return w
}

func (nc *Conn) Events() <-chan Serializable {
	return nc.events
}

func (nc *Conn) Close() {
	if nc != nil {
		nc.close()
	}
}

func (nc *Conn) close() {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	if nc.conn != nil {
		close(nc.done)
		nc.conn.Close()
		nc.conn = nil
	}
}

func (nc *Conn) removeRequest(key string) {
	delete(nc.requests, key)
}

func (nc *Conn) SendRcv(ctx context.Context, m Serializable) (Serializable, error) {
	if nc.conn == nil {
		return nil, ErrInvalidConnection
	}

	if m == nil {
		return nil, ErrBadPacket
	}

	respCh, err := nc.send(m)
	if err != nil {
		return nil, err
	}

	if _, wait := m.Key(); !wait {
		return nil, nil
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-respCh:
		return resp.Message, resp.Err
	}
}

func (nc *Conn) send(m Serializable) (<-chan Response, error) {
	w := nc.bw.w
	log := nc.Opts.Logger

	respCh := make(chan Response, 1)

	key, ok := m.Key()
	if !ok {
		log.Debug().
			Msg("fire and forget: response tracking disabled")
	}
	if key != "" {
		nc.mu.Lock()
		nc.requests[key] = &requestEntry{
			key:      key,
			response: respCh,
			timer: time.AfterFunc(nc.Opts.Timeout, func() {
				select {
				case respCh <- Response{Err: fmt.Errorf("request timeout")}:
				default:
				}

				nc.removeRequest(key)
			}),
		}
		nc.mu.Unlock()
	}

	frame, err := WrapMessage(m)
	if err != nil {
		return nil, err
	}

	bytes, err := frame.Pack()
	if err != nil {
		return nil, err
	}

	_, err = w.Write(bytes)
	if err != nil {
		return nil, ErrBadPacket
	}

	log.Debug().
		Str("fingerprint", key).
		Msgf("[TX] %s", frame.Name())
	log.Debug().
		Msgf("hexdump:\n%s", strings.TrimRight(hex.Dump(bytes), "\n"))

	if _, ok := m.Key(); !ok {
		close(respCh)
	}
	return respCh, nil
}

func (nc *Conn) Listen() {
	r := nc.br
	log := nc.Opts.Logger

	for {
		select {
		case <-nc.done:
			return
		default:
		}

		n, err := r.r.Read(r.buf)
		if err != nil {
			log.Debug().Err(err)
			continue
		}
		log.Debug().
			Int("len", n).
			Msgf("socket rcv")

		frames, err := ParseFrames(r.buf[:n])
		if err != nil {
			log.Error().Err(err)
			continue
		}

		for _, frame := range frames {
			m, err := DecodeFrame(frame)
			if err != nil {
				log.Error().Err(err)
				continue
			}

			if err := m.Unpack(frame.Payload); err != nil {
				log.Error().Err(err)
				continue
			}

			nc.mu.Lock()
			key, ok := m.Key()
			if !ok {
				log.Error().Msg("unable to track message")
			}
			if entry, ok := nc.requests[key]; ok {
				entry.timer.Stop()
				select {
				case entry.response <- Response{Message: m}:
				default:
				}
				delete(nc.requests, key)
			} else {
				select {
				case nc.events <- m:
				default:
					log.Warn().Msgf("dropped event: %s", m)
				}
			}
			nc.mu.Unlock()

			log.Debug().
				Str("fingerprint", key).
				Msgf("[RX] %s", frame.Name())
			log.Debug().
				MsgFunc(
					func() string {
						bytes, _ := frame.Pack()
						return fmt.Sprintf("hexdump:\n%s", strings.TrimRight(hex.Dump(bytes), "\n"))
					},
				)
		}
	}
}
