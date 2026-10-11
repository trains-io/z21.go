//go:build !integration

package client

import "testing"

func TestEnsureUDPPort(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "ipv4 without port", in: "192.168.0.111", want: "192.168.0.111:21105"},
		{name: "ipv4 with port", in: "192.168.0.111:21106", want: "192.168.0.111:21106"},
		{name: "hostname without port", in: "z21.local", want: "z21.local:21105"},
		{name: "hostname with port", in: "z21.local:1234", want: "z21.local:1234"},
		{name: "bare ipv6", in: "::1", want: "[::1]:21105"},
		{name: "bracketed ipv6 without port", in: "[::1]", want: "[::1]:21105"},
		{name: "bracketed ipv6 with port", in: "[::1]:21106", want: "[::1]:21106"},
		{name: "empty", in: "", wantErr: true},
		{name: "malformed", in: "a:b:c", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ensureUDPPort(tt.in, DefaultPort)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ensureUDPPort(%q) = %q, want error", tt.in, got)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("ensureUDPPort(%q) = %q, %v, want %q", tt.in, got, err, tt.want)
			}
		})
	}
}
