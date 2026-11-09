VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +%Y-%m-%d)
PKG := github.com/allanmedeiros71/vanitas/internal/cli
LDFLAGS := -s -w -X $(PKG).version=$(VERSION) -X $(PKG).commit=$(COMMIT) -X $(PKG).date=$(DATE)

.PHONY: build test test-integration lint fmt run clean

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o vanitas ./cmd/vanitas

test:
	CGO_ENABLED=0 go test ./...

test-integration:
	CGO_ENABLED=0 go test -tags integration ./...

lint:
	golangci-lint fmt --diff && golangci-lint run ./...

fmt:
	golangci-lint fmt

run:
	go run ./cmd/vanitas

clean:
	rm -f vanitas vanitas.exe coverage.out coverage.xml
	rm -rf dist
