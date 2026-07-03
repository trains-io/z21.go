## z21 - Go Client

A [Go](https://go.dev/) client for [z21](https://www.z21.eu/en) command station.

### Installation

```bash
# To get the latest released Go client:
go get github.com/trains-io/z21.go@latest

# To get a specific version:
go get github.com/trains-io/z21.go@v0.1.0
```

### Basic Usage

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

### Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for how to build, run tests, and run integration tests against the z21 simulator.

### License

[MIT](./LICENSE)
