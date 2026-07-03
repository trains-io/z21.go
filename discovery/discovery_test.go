package discovery

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/trains-io/z21.go/client"
	"github.com/trains-io/z21.go/protocol"
)

var testHWInfoReply = []byte{0x01, 0x02, 0x00, 0x00, 0x43, 0x01, 0x00, 0x00}

func TestHostsInNetwork(t *testing.T) {
	_, network, err := net.ParseCIDR("192.168.2.0/30")
	if err != nil {
		t.Fatal(err)
	}

	hosts := HostsInNetwork(network)
	if len(hosts) != 2 {
		t.Fatalf("len(hosts) = %d, want 2", len(hosts))
	}
	if hosts[0].String() != "192.168.2.1" || hosts[1].String() != "192.168.2.2" {
		t.Fatalf("hosts = %v", hosts)
	}
}

func TestParseTargetCIDR(t *testing.T) {
	n, err := ParseTarget("10.0.0.0/24")
	if err != nil {
		t.Fatal(err)
	}
	if n.String() != "10.0.0.0/24" {
		t.Fatalf("network = %q", n)
	}
}

func TestProbe(t *testing.T) {
	conn, _ := startHWInfoStub(t)
	defer conn.Close()

	host, port := udpHostPort(t, conn.LocalAddr())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	device, err := Probe(ctx, net.ParseIP(host), port, nil)
	if err != nil {
		t.Fatalf("Probe() error = %v", err)
	}
	if device.HWInfo.HwType != protocol.HwTypeZ21New {
		t.Fatalf("HwType = %#x, want %#x", device.HWInfo.HwType, protocol.HwTypeZ21New)
	}
	if device.HWInfo.FirmwareVersion != 0x00000143 {
		t.Fatalf("FirmwareVersion = %#x, want 0x00000143", device.HWInfo.FirmwareVersion)
	}
}

func TestScanFindsDevice(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	reply, err := protocol.Message{
		Header: protocol.HeaderLANGetHWInfo,
		Data:   testHWInfoReply,
	}.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		buf := make([]byte, 1472)
		for {
			_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				return
			}
			msgs, err := protocol.ParseAll(buf[:n])
			if err != nil || len(msgs) != 1 || msgs[0].Header != protocol.HeaderLANGetHWInfo {
				continue
			}
			_, _ = conn.WriteToUDP(reply, addr)
		}
	}()

	host, port := udpHostPort(t, conn.LocalAddr())

	_, ipNet, err := net.ParseCIDR(host + "/32")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	devices, err := Scan(ctx, ipNet, Options{Port: port, Timeout: time.Second, Concurrency: 1})
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("Scan() found %d devices, want 1", len(devices))
	}
	if devices[0].HWInfo.HwType != protocol.HwTypeZ21New {
		t.Fatalf("Scan() HwType = %#x, want %#x", devices[0].HWInfo.HwType, protocol.HwTypeZ21New)
	}
}

func startHWInfoStub(t *testing.T) (*net.UDPConn, []byte) {
	t.Helper()

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatal(err)
	}

	reply, err := protocol.Message{
		Header: protocol.HeaderLANGetHWInfo,
		Data:   testHWInfoReply,
	}.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 1472)
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		msgs, err := protocol.ParseAll(buf[:n])
		if err != nil || len(msgs) != 1 || msgs[0].Header != protocol.HeaderLANGetHWInfo {
			return
		}
		_, _ = conn.WriteToUDP(reply, addr)
	}()

	t.Cleanup(func() { <-done })

	return conn, reply
}

func udpHostPort(t *testing.T, addr net.Addr) (string, int) {
	t.Helper()
	host, portStr, err := net.SplitHostPort(addr.String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := net.LookupPort("udp", portStr)
	if err != nil {
		t.Fatal(err)
	}
	return host, port
}

func TestScanDefaults(t *testing.T) {
	_, network, err := net.ParseCIDR("127.0.0.0/31")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	devices, err := Scan(ctx, network, Options{})
	if err != nil && err != context.DeadlineExceeded && err != context.Canceled {
		t.Fatalf("Scan() error = %v", err)
	}
	_ = devices
	_ = client.DefaultPort
}
