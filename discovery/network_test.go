package discovery

import (
	"net"
	"testing"
)

func TestNetworkFromInterfaceLoopback(t *testing.T) {
	ipNet, err := NetworkFromInterface("lo")
	if err != nil {
		t.Skipf("loopback interface not available: %v", err)
	}
	if ipNet.IP.To4() == nil {
		t.Fatalf("expected IPv4 network, got %s", ipNet)
	}
}

func TestCompareIP(t *testing.T) {
	a := net.ParseIP("192.168.1.1")
	b := net.ParseIP("192.168.1.2")
	if compareIP(a, b) >= 0 {
		t.Fatalf("compareIP(%v, %v) = %d, want < 0", a, b, compareIP(a, b))
	}
	if compareIP(a, a) != 0 {
		t.Fatalf("compareIP equal IPs should be 0")
	}
}
