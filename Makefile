.PHONY: build test lint coverage clean install help

# Build variables
BINARY_NAME := opencligen
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Default target
all: lint test build

## build: Build the binary
build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/opencligen

## test: Run all tests
test:
	go test -race ./...

## lint: Run golangci-lint
lint:
	golangci-lint run

## coverage: Run tests with coverage report
coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out
	@echo ""
	@echo "To view HTML coverage report, run: go tool cover -html=coverage.out"

## coverage-html: Generate and open HTML coverage report
coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## clean: Remove build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

## install: Install the binary to GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/opencligen

## fmt: Format code
fmt:
	go fmt ./...
	goimports -w -local github.com/crunchloop/opencligen .

## vet: Run go vet
vet:
	go vet ./...

## mod-tidy: Tidy go modules
mod-tidy:
	go mod tidy

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /'
