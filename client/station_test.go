//go:build !integration

package client_test

import (
	"testing"

	"github.com/trains-io/z21.go/client"
	"github.com/trains-io/z21.go/protocol"
	"github.com/trains-io/z21.go/server"
	"github.com/trains-io/z21.go/server/servertest"
)

var (
	testHWInfo    = protocol.HWInfo{HwType: protocol.HwTypeZ21New, FirmwareVersion: 0x0143}
	testSerial    = uint32(307214)
	testXVersion  = protocol.XVersion{XBusVersion: 0x30, CommandStationID: 0x12}
	testXFirmware = protocol.XFirmware{VersionMSB: 0x01, VersionLSB: 0x42}
	testXStatus   = protocol.XStatus{CentralState: 0x01} // emergency stop
)

func reply(m protocol.Message) server.Handler {
	return server.HandlerFunc(func(w server.ResponseWriter, _ *server.Request) { w.Reply(m) })
}

// stationMux answers the device information requests with fixed values.
func stationMux() *server.Mux {
	mux := server.NewMux()
	mux.Handle(protocol.HeaderLANGetSerialNumber, reply(protocol.EncodeSerialNumber(testSerial)))
	mux.Handle(protocol.HeaderLANGetHWInfo, reply(protocol.EncodeHWInfo(testHWInfo)))
	mux.Handle(protocol.HeaderLANGetCode, reply(protocol.EncodeCode(protocol.CodeNoLock)))
	mux.HandleLANXDB0(0x21, 0x21, reply(protocol.EncodeXVersion(testXVersion)))
	mux.HandleLANXDB0(0x21, 0x24, reply(protocol.EncodeXStatus(testXStatus)))
	mux.HandleLANXDB0(0xF1, 0x0A, reply(protocol.EncodeXFirmware(testXFirmware)))
	return mux
}

// startStation runs h (stationMux if nil) on an in-process server until the
// test ends.
func startStation(t *testing.T, h server.Handler) *servertest.Server {
	t.Helper()
	if h == nil {
		h = stationMux()
	}
	srv := servertest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

func dialStation(t *testing.T, srv *servertest.Server, opts ...client.Option) *client.Client {
	t.Helper()
	c, err := srv.Dial(opts...)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}
