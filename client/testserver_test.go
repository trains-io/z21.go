//go:build integration

package client_test

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/trains-io/z21.go/client"
	"github.com/trains-io/z21.go/protocol"
)

func requireDocker(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not installed; skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "info")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("docker not available: %v\n%s", err, out)
	}
}

func startTestServer(t *testing.T) (addr string, terminate func()) {
	t.Helper()
	requireDocker(t)

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		ExposedPorts: []string{"21105/udp"},
	}
	if image := os.Getenv("Z21_TESTSERVER_IMAGE"); image != "" {
		req.Image = image
	} else if ctxDir := os.Getenv("Z21_TESTSERVER_DOCKERFILE"); ctxDir != "" {
		req.FromDockerfile = testcontainers.FromDockerfile{
			Context:    ctxDir,
			Dockerfile: "Dockerfile",
			KeepImage:  true,
		}
	} else {
		t.Skip("set Z21_TESTSERVER_IMAGE (recommended: ghcr.io/trains-io/z21-sim:latest) or Z21_TESTSERVER_DOCKERFILE")
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		if isDockerInfraError(err) {
			t.Skipf("cannot start z21 test server container: %v", err)
		}
		require.NoError(t, err)
	}

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "21105/udp")
	require.NoError(t, err)

	return host + ":" + port.Port(), func() {
		require.NoError(t, container.Terminate(ctx))
	}
}

func waitForServer(t *testing.T, addr string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for {
		c, err := client.Dial(addr)
		if err == nil {
			_, err = c.Call(ctx, protocol.GetHWInfo())
			_ = c.Close()
			if err == nil {
				return
			}
		}

		select {
		case <-ctx.Done():
			t.Fatalf("z21 test server not ready at %s: %v", addr, ctx.Err())
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func TestGetHWInfo(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	waitForServer(t, addr)

	c, err := client.Dial(addr)
	require.NoError(t, err)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msgs, err := c.Call(ctx, protocol.GetHWInfo())
	require.NoError(t, err)
	require.NotEmpty(t, msgs)

	var hwInfo *protocol.Message
	for i := range msgs {
		if msgs[i].Header == protocol.HeaderLANGetHWInfo {
			hwInfo = &msgs[i]
			break
		}
	}
	require.NotNil(t, hwInfo, "expected LAN_GET_HWINFO reply, got %#v", msgs)
	require.GreaterOrEqual(t, len(hwInfo.Data), 8, "expected HwType + firmware version")
}

func TestGetSerialNumber(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	waitForServer(t, addr)

	c, err := client.Dial(addr)
	require.NoError(t, err)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msgs, err := c.Call(ctx, protocol.GetSerialNumber())
	if err != nil {
		t.Skipf("simulator did not reply to LAN_GET_SERIAL_NUMBER: %v", err)
	}

	serial, ok := protocol.SerialFromMessages(msgs)
	require.True(t, ok, "expected LAN_GET_SERIAL_NUMBER reply, got %#v", msgs)
	require.NotEmpty(t, serial)
}

func TestSystemStateGetData(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	waitForServer(t, addr)

	c, err := client.Dial(addr)
	require.NoError(t, err)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msgs, err := c.Call(ctx, protocol.SystemStateGetData())
	require.NoError(t, err)
	require.NotEmpty(t, msgs)

	var state *protocol.Message
	for i := range msgs {
		if msgs[i].Header == protocol.HeaderLANSystemStateChanged {
			state = &msgs[i]
			break
		}
	}
	require.NotNil(t, state, "expected LAN_SYSTEMSTATE_DATACHANGED reply, got %#v", msgs)

	parsed, err := protocol.ParseSystemState(state.Data)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(state.Data), 16, "expected 16-byte system state payload")
	_ = parsed
}

func TestGetXFirmware(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	waitForServer(t, addr)

	c, err := client.Dial(addr)
	require.NoError(t, err)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msgs, err := c.Call(ctx, protocol.GetXFirmware())
	require.NoError(t, err)

	fw, err := protocol.XFirmwareFromMessages(msgs)
	require.NoError(t, err, "expected LAN_X_GET_FIRMWARE_VERSION reply, got %#v", msgs)
	require.NotEmpty(t, protocol.FormatXFirmwareVersion(fw))
}

func TestGetCode(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	waitForServer(t, addr)

	c, err := client.Dial(addr)
	require.NoError(t, err)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msgs, err := c.Call(ctx, protocol.GetCode())
	require.NoError(t, err)

	code, err := protocol.CodeFromMessages(msgs)
	require.NoError(t, err, "expected LAN_GET_CODE reply, got %#v", msgs)
	require.Equal(t, protocol.CodeNoLock, code)
}

func TestGetBroadcastFlags(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	waitForServer(t, addr)

	c, err := client.Dial(addr)
	require.NoError(t, err)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Send(ctx, protocol.SetBroadcastFlags(0x103)); err != nil {
		t.Fatalf("SetBroadcastFlags() error = %v", err)
	}

	msgs, err := c.Call(ctx, protocol.GetBroadcastFlags())
	require.NoError(t, err)

	flags, err := protocol.BroadcastFlagsFromMessages(msgs)
	require.NoError(t, err, "expected LAN_GET_BROADCASTFLAGS reply, got %#v", msgs)
	require.Equal(t, uint32(0x103), flags)
}

func TestGetXStatus(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	waitForServer(t, addr)

	c, err := client.Dial(addr)
	require.NoError(t, err)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msgs, err := c.Call(ctx, protocol.GetXStatus())
	require.NoError(t, err)

	status, err := protocol.XStatusFromMessages(msgs)
	require.NoError(t, err, "expected LAN_X_GET_STATUS reply, got %#v", msgs)
	require.NotEmpty(t, protocol.FormatXStatusFlags(status.CentralState))
}

func isDockerInfraError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "build image") ||
		strings.Contains(msg, "certificate") ||
		strings.Contains(msg, "connect: connection refused") ||
		strings.Contains(msg, "Cannot connect to the Docker daemon")
}
