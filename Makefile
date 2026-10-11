.PHONY: help build test test-live test-cover fmt vet lint clean

GO ?= go

help:
	@echo "Usage: make <target>"
	@echo "Targets:"
	@echo "  build      - Build the Tile wrapper"
	@echo "  test       - Run the default fixture-backed test suite"
	@echo "  test-live  - Run opt-in live tests when TILE_LIVE_TESTS=1"
	@echo "  test-cover - Run tests with coverage"
	@echo "  fmt        - Format Go code"
	@echo "  vet        - Run go vet"
	@echo "  lint       - Run pinned Staticcheck analysis"
	@echo "  clean      - Clean Go test cache"

build:
	$(GO) build ./...

test:
	$(GO) test ./...

test-live:
	@echo "No live Tile tests are retained; default tests use redacted fixtures only"

test-cover:
	$(GO) test -cover ./...

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

lint:
	GOTOOLCHAIN=go1.25.0 $(GO) run honnef.co/go/tools/cmd/staticcheck@v0.6.1 ./...

clean:
	$(GO) clean -testcache
