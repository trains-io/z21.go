package discovery

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/trains-io/z21.go/client"
	"github.com/trains-io/z21.go/protocol"
)

const defaultConcurrency = 200

// Device is a reachable Z21 command station discovered on the network.
type Device struct {
	IP     net.IP
	Port   int
	Serial string
}

// Options configures a network scan.
type Options struct {
	Port        int
	Timeout     time.Duration
	Concurrency int
	Observer    client.Observer
}

// ParseTarget resolves a CIDR network (for example "192.168.2.0/24") or a network
// interface name (for example "eth0") to an IPv4 network.
func ParseTarget(target string) (*net.IPNet, error) {
	if _, ipNet, err := net.ParseCIDR(target); err == nil {
		if ipNet.IP.To4() == nil {
			return nil, fmt.Errorf("discovery: only IPv4 networks are supported")
		}
		return ipNet, nil
	}
	return NetworkFromInterface(target)
}

// NetworkFromInterface returns the IPv4 network for the named interface.
func NetworkFromInterface(name string) (*net.IPNet, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.To4() == nil {
			continue
		}
		return ipNet, nil
	}

	return nil, fmt.Errorf("discovery: no IPv4 network on interface %q", name)
}

// HostsInNetwork returns host addresses in n, excluding the network and broadcast addresses.
func HostsInNetwork(n *net.IPNet) []net.IP {
	ip := n.IP.Mask(n.Mask)
	if ip == nil || ip.To4() == nil {
		return nil
	}

	ip = ip.To4()
	var hosts []net.IP
	for ; n.Contains(ip); incIP(ip) {
		hosts = append(hosts, append(net.IP(nil), ip...))
	}

	if len(hosts) > 2 {
		return hosts[1 : len(hosts)-1]
	}
	return hosts
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] != 0 {
			break
		}
	}
}

// Probe sends LAN_GET_SERIAL_NUMBER to a single host and reports whether it is a Z21 device.
func Probe(ctx context.Context, ip net.IP, port int, observer client.Observer) (Device, error) {
	device := Device{IP: ip, Port: port}

	addr := net.JoinHostPort(ip.String(), strconv.Itoa(port))
	c, err := client.Dial(addr, client.WithObserver(observer))
	if err != nil {
		return device, err
	}
	defer c.Close()

	msgs, err := c.Call(ctx, protocol.GetSerialNumber())
	if err != nil {
		return device, err
	}

	serial, ok := protocol.SerialFromMessages(msgs)
	if !ok {
		return device, fmt.Errorf("discovery: no LAN_GET_SERIAL_NUMBER reply from %s", addr)
	}

	device.Serial = serial
	return device, nil
}

// Scan probes every host in network for Z21 devices.
func Scan(ctx context.Context, network *net.IPNet, opts Options) ([]Device, error) {
	if network == nil {
		return nil, fmt.Errorf("discovery: network is required")
	}
	if network.IP.To4() == nil {
		return nil, fmt.Errorf("discovery: only IPv4 networks are supported")
	}

	port := opts.Port
	if port == 0 {
		port = client.DefaultPort
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 2 * time.Second
	}
	concurrency := opts.Concurrency
	if concurrency <= 0 {
		concurrency = defaultConcurrency
	}

	hosts := HostsInNetwork(network)
	results := make([]Device, 0, len(hosts))
	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)
	sem := make(chan struct{}, concurrency)

	for _, host := range hosts {
		if err := ctx.Err(); err != nil {
			break
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(ip net.IP) {
			defer wg.Done()
			defer func() { <-sem }()

			probeCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			device, err := Probe(probeCtx, ip, port, opts.Observer)
			if err != nil {
				return
			}

			mu.Lock()
			results = append(results, device)
			mu.Unlock()
		}(host)
	}

	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		return compareIP(results[i].IP, results[j].IP) < 0
	})

	if err := ctx.Err(); err != nil {
		return results, err
	}
	return results, nil
}

func compareIP(a, b net.IP) int {
	aa := a.To4()
	bb := b.To4()
	if aa == nil || bb == nil {
		return 0
	}
	for i := range aa {
		if aa[i] < bb[i] {
			return -1
		}
		if aa[i] > bb[i] {
			return 1
		}
	}
	return 0
}
