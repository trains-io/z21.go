# Contributing

Thanks for contributing to `github.com/trains-io/z21.go`. This document covers how to run tests in this 
repository.

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.24+ | Build and test |
| Docker | recent | Integration tests (testcontainers) |

Integration tests need a running Docker daemon (`docker info`).

## Quick start

From the repository root:

```bash
make test              # unit tests (no Docker)
make test-integration  # client tests against a Z21 simulator in Docker
```

## Unit tests

Run all packages:

```bash
make test
```

Or directly:

```bash
go test ./... -count=1
```

### Single package

```bash
go test ./protocol/... -count=1
go test ./client/... -count=1
go test ./discovery/... -count=1
```

Filter by test name:

```bash
go test ./protocol/... -run TestReadCV -v -count=1
```

Use `-count=1` while iterating to disable the test cache.

## Integration tests

Integration tests are in `client/` and gated by the `integration` build tag. They start a Z21 LAN simulator in Docker via [testcontainers](https://golang.testcontainers.org/) and exercise the UDP client.

### Run integration tests

```bash
make test-integration
```

By default, `make test-integration` uses `ghcr.io/trains-io/z21-sim:latest`. Override with:

```bash
make test-integration Z21_TESTSERVER_IMAGE=z21-testserver:local
```

Manual equivalent:

```bash
go test -tags=integration ./client/... -count=1 -timeout=10m
```

### Run one integration test

```bash
go test -tags=integration ./client/... -run TestGetHWInfo -v -count=1
```

If Docker is not installed or not running, integration tests skip rather than fail.

### Local Dockerfile

```bash
export Z21_TESTSERVER_DOCKERFILE=/path/to/z21-sim
go test -tags=integration ./client/... -count=1 -timeout=10m
```

## CI

GitHub Actions runs unit and integration tests on every push and pull request.
