package server_test

import (
	"reflect"
	"testing"

	"github.com/trains-io/z21.go/protocol"
	"github.com/trains-io/z21.go/server"
)

// recorder is a ResponseWriter that records replies.
type recorder struct {
	replies []protocol.Message
}

func (r *recorder) Reply(msgs ...protocol.Message)        { r.replies = append(r.replies, msgs...) }
func (r *recorder) Broadcast(uint32, ...protocol.Message) {}
func (r *recorder) BroadcastLocoInfo(protocol.LocoInfo)   {}

// named replies with LAN_GET_CODE carrying id, so tests can tell handlers apart.
func named(id byte) server.Handler {
	return server.HandlerFunc(func(w server.ResponseWriter, _ *server.Request) {
		w.Reply(protocol.EncodeCode(id))
	})
}

func TestMuxRouting(t *testing.T) {
	const (
		hwInfo = iota + 1
		xGetVersionFamily
		xTrackPowerOn
		xStop
		lanXFallback
	)
	mux := server.NewMux()
	mux.Handle(protocol.HeaderLANGetHWInfo, named(hwInfo))
	mux.HandleLANX(0x21, named(xGetVersionFamily))
	mux.HandleLANXDB0(0x21, 0x81, named(xTrackPowerOn))
	mux.HandleLANX(0x80, named(xStop))

	withFallback := server.NewMux()
	withFallback.Handle(protocol.HeaderLANX, named(lanXFallback))

	unknown := []protocol.Message{protocol.EncodeUnknownCommand()}
	code := func(id byte) []protocol.Message { return []protocol.Message{protocol.EncodeCode(id)} }

	tests := []struct {
		name string
		mux  *server.Mux
		msg  protocol.Message
		want []protocol.Message
	}{
		{name: "LAN header", mux: mux, msg: protocol.GetHWInfo(), want: code(hwInfo)},
		{name: "unregistered LAN header is ignored", mux: mux, msg: protocol.GetSerialNumber(), want: nil},
		{name: "X-header and DB0 beat X-header", mux: mux, msg: protocol.SetTrackPower(true), want: code(xTrackPowerOn)},
		{name: "X-header for other DB0", mux: mux, msg: protocol.SetTrackPower(false), want: code(xGetVersionFamily)},
		{name: "X-header without DB0", mux: mux, msg: protocol.EncodeLANX(0x80), want: code(xStop)},
		{name: "unmatched LAN_X", mux: mux, msg: protocol.GetLocoInfo(3), want: unknown},
		{
			name: "bad checksum",
			mux:  mux,
			msg:  protocol.Message{Header: protocol.HeaderLANX, Data: []byte{0x21, 0x81, 0x00}},
			want: unknown,
		},
		{name: "too short", mux: mux, msg: protocol.Message{Header: protocol.HeaderLANX, Data: []byte{0x21}}, want: unknown},
		{name: "LAN_X fallback", mux: withFallback, msg: protocol.GetLocoInfo(3), want: code(lanXFallback)},
		{
			name: "malformed LAN_X skips fallback",
			mux:  withFallback,
			msg:  protocol.Message{Header: protocol.HeaderLANX, Data: []byte{0x21}},
			want: unknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &recorder{}
			tt.mux.ServeZ21(w, &server.Request{Message: tt.msg})
			if !reflect.DeepEqual(w.replies, tt.want) {
				t.Fatalf("replies = %+v, want %+v", w.replies, tt.want)
			}
		})
	}
}

func TestMuxRegistrationPanics(t *testing.T) {
	tests := []struct {
		name     string
		register func(m *server.Mux)
	}{
		{name: "nil handler", register: func(m *server.Mux) { m.Handle(protocol.HeaderLANGetHWInfo, nil) }},
		{name: "duplicate LAN header", register: func(m *server.Mux) {
			m.Handle(protocol.HeaderLANGetHWInfo, named(1))
			m.Handle(protocol.HeaderLANGetHWInfo, named(2))
		}},
		{name: "duplicate X-header", register: func(m *server.Mux) {
			m.HandleLANX(0x21, named(1))
			m.HandleLANX(0x21, named(2))
		}},
		{name: "duplicate X-header and DB0", register: func(m *server.Mux) {
			m.HandleLANXDB0(0x21, 0x81, named(1))
			m.HandleLANXDB0(0x21, 0x81, named(2))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			tt.register(server.NewMux())
		})
	}
}

func TestServerWithMuxRepliesUnknownCommand(t *testing.T) {
	mux := server.NewMux()
	mux.Handle(protocol.HeaderLANGetHWInfo, server.HandlerFunc(hwInfoHandler))
	srv := startServer(t, server.WithHandler(mux))
	c := dial(t, srv)

	if _, err := c.Call(testContext(t), protocol.GetHWInfo()); err != nil {
		t.Fatalf("Call GetHWInfo: %v", err)
	}
	if err := c.Send(testContext(t), protocol.GetLocoInfo(3)); err != nil {
		t.Fatalf("Send GetLocoInfo: %v", err)
	}
	msgs, err := c.ReadPacket(testContext(t))
	if err != nil {
		t.Fatalf("ReadPacket: %v", err)
	}
	if !reflect.DeepEqual(msgs, []protocol.Message{protocol.EncodeUnknownCommand()}) {
		t.Fatalf("got %+v, want LAN_X_UNKNOWN_COMMAND", msgs)
	}
}
