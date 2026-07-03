.PHONY: help test test-integration tidy

.DEFAULT_GOAL := help

# Default Z21 LAN simulator for integration tests (override: make test-integration Z21_TESTSERVER_IMAGE=...)
Z21_TESTSERVER_IMAGE ?= ghcr.io/trains-io/z21-sim:latest
export Z21_TESTSERVER_IMAGE

help: ## Show targets
	@grep -E '^[a-zA-Z0-9_-]+:.*##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  %-18s %s\n", $$1, $$2}'

test: ## Run unit tests
	go test ./... -count=1

test-integration: ## Run integration tests (requires Docker; uses ghcr.io/trains-io/z21-sim by default)
	go test -tags=integration ./... -count=1 -timeout=10m
