.PHONY: help check fmt-check vet test fuzz test-integration

.DEFAULT_GOAL := help

FUZZTIME ?= 30s

# Default Z21 LAN reference server for integration tests (override: make test-integration Z21_TESTSERVER_IMAGE=...)
Z21_TESTSERVER_IMAGE ?= ghcr.io/trains-io/z21-ref:latest
export Z21_TESTSERVER_IMAGE

help: ## Show targets
	@grep -E '^[a-zA-Z0-9_-]+:.*##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  %-18s %s\n", $$1, $$2}'

check: fmt-check vet test ## Run the same checks as CI (gofmt, go vet, unit tests with -race)

fmt-check: ## Fail if any Go file needs gofmt
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "Files need gofmt:"; echo "$$unformatted"; exit 1; \
	fi

vet: ## Run go vet, including integration-tagged files
	go vet ./...
	go vet -tags=integration ./...

test: ## Run unit tests with the race detector
	go test -race ./... -count=1

fuzz: ## Fuzz the protocol codecs, one target at a time (override: make fuzz FUZZTIME=5m)
	@for target in $$(go test ./protocol -list '^Fuzz' | grep '^Fuzz'); do \
		go test ./protocol -run '^$$' -fuzz "^$$target$$" -fuzztime $(FUZZTIME) || exit 1; \
	done

test-integration: ## Run integration tests (requires Docker; uses ghcr.io/trains-io/z21-ref by default)
	go test -tags=integration ./client/... -count=1 -timeout=10m
