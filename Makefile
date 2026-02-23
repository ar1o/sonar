.PHONY: build test lint vet install clean

VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := -X github.com/ar1o/sonar/internal/cli.version=$(VERSION) -X github.com/ar1o/sonar/internal/cli.commit=$(COMMIT) -X github.com/ar1o/sonar/internal/cli.buildDate=$(BUILD_DATE)

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o ./bin/sonar ./cmd/sonar

test:
	go test ./...

lint: vet
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "staticcheck not found, skipping"

vet:
	go vet ./...

install:
	CGO_ENABLED=0 go install -ldflags "$(LDFLAGS)" ./cmd/sonar

clean:
	rm -rf ./bin/
