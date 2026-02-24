BINARY_NAME=devnews
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: build install test lint clean run

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

install:
	go install $(LDFLAGS) .

run:
	go run $(LDFLAGS) .

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

.DEFAULT_GOAL := build
