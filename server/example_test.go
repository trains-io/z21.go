package server_test

import (
	"context"
	"fmt"
	"time"

	"github.com/trains-io/z21.go/protocol"
	"github.com/trains-io/z21.go/server"
	"github.com/trains-io/z21.go/server/servertest"
)

// A minimal command station: it reports its hardware, switches track power and
// drives locos while power is on. Handlers run one at a time, so the state
// needs no locking as long as only handlers touch it.
func Example() {
	trackPower := true
	locos := map[uint16]protocol.LocoInfo{}

	mux := server.NewMux()
	mux.Handle(protocol.HeaderLANGetHWInfo, server.HandlerFunc(func(w server.ResponseWriter, _ *server.Request) {
		w.Reply(protocol.EncodeHWInfo(protocol.HWInfo{HwType: protocol.HwTypeZ21New, FirmwareVersion: 0x0143}))
	}))
	setPower := func(on bool) server.Handler {
		return server.HandlerFunc(func(w server.ResponseWriter, _ *server.Request) {
			trackPower = on
			w.Reply(protocol.EncodeTrackPowerBC(on))
			w.Broadcast(protocol.BroadcastFlagXpressNet, protocol.EncodeTrackPowerBC(on))
		})
	}
	mux.HandleLANXDB0(0x21, 0x80, setPower(false))
	mux.HandleLANXDB0(0x21, 0x81, setPower(true))
	mux.HandleLANXDB0(0xE3, 0xF0, server.HandlerFunc(func(w server.ResponseWriter, r *server.Request) {
		addr, err := protocol.ParseGetLocoInfo(r.Message.Data)
		if err != nil {
			return
		}
		w.Reply(protocol.EncodeLocoInfo(protocol.LocoInfo{Address: addr, SpeedSteps: 4, Forward: locos[addr].Forward, Speed: locos[addr].Speed}))
	}))
	mux.HandleLANX(0xE4, server.HandlerFunc(func(w server.ResponseWriter, r *server.Request) {
		d, err := protocol.ParseSetLocoDrive(r.Message.Data)
		if err != nil || !trackPower {
			return
		}
		info := protocol.LocoInfo{Address: d.Address, SpeedSteps: 4, Forward: d.Forward, Speed: d.Speed}
		locos[d.Address] = info
		w.BroadcastLocoInfo(info)
	}))

	// In production: srv, _ := server.New(":21105", server.WithHandler(mux)); srv.Serve(ctx)
	ts := servertest.NewServer(mux)
	defer ts.Close()

	c, _ := ts.Dial()
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	msgs, _ := c.Call(ctx, protocol.GetHWInfo())
	hw, _ := protocol.HWInfoFromMessages(msgs)
	fmt.Printf("firmware %x\n", hw.FirmwareVersion)

	_ = c.Send(ctx, protocol.SetBroadcastFlags(protocol.BroadcastFlagXpressNet))
	_ = c.Send(ctx, protocol.GetLocoInfo(3)) // also subscribes to loco 3
	msgs, _ = c.ReadPacket(ctx)
	info, _ := protocol.LocoInfoFromMessages(msgs)
	fmt.Printf("loco %d speed %d\n", info.Address, info.Speed)

	_ = c.Send(ctx, protocol.SetLocoDrive(3, protocol.LocoSpeedSteps128, true, 42))
	msgs, _ = c.ReadPacket(ctx)
	info, _ = protocol.LocoInfoFromMessages(msgs)
	fmt.Printf("loco %d speed %d forward %t\n", info.Address, info.Speed, info.Forward)

	msgs, _ = c.Call(ctx, protocol.SetTrackPower(false))
	on, _ := protocol.ParseTrackPowerBC(msgs[0].Data)
	fmt.Printf("track power on: %t\n", on)

	// Output:
	// firmware 143
	// loco 3 speed 0
	// loco 3 speed 42 forward true
	// track power on: false
}
