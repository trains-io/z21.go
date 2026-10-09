## z21 - Go library for client and server

A [Go](https://go.dev/) client for [z21](https://www.z21.eu/en) command station.

### Installation

```bash
# To get the latest released Go client:
go get github.com/trains-io/z21.go@latest

# To get a specific version:
go get github.com/trains-io/z21.go@v0.1.0
```

### Basic Usage

#### Client

```go
import "github.com/trains-io/z21.go/client"
import "github.com/trains-io/z21.go/protocol"

// Connect to a z21 command station
c, _ := client.Dial("192.168.1.100")

ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

// Request
msgs, _ := c.Call(ctx, protocol.GetHWInfo())

// Replies
info, _ := protocol.HWInfoFromMessages(msgs)
```

#### Server

```go
import "github.com/trains-io/z21.go/server"
import "github.com/trains-io/z21.go/protocol"

// Handle requests
handler := server.HandlerFunc(func(w server.ResponseWriter, r *server.Request) {
	if r.Message.Header == protocol.HeaderLANGetHWInfo {
		w.Reply(protocol.EncodeHWInfo(protocol.HWInfo{HwType: protocol.HwTypeZ21New, FirmwareVersion: 0x0143}))
	}
})

// Listen on the Z21 LAN port and serve until ctx is cancelled
srv, _ := server.New(":21105", server.WithHandler(handler))
defer srv.Close()

srv.Serve(ctx)
```

### Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for how to build, run tests, and run integration tests against the z21 reference implementation.

### License

[MIT](./LICENSE)
