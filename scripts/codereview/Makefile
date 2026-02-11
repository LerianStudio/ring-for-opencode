.PHONY: all build build-legacy build-scr test test-coverage clean install fmt vet golangci-lint lint \
	scope-detector static-analysis ast-extractor call-graph data-flow compile-context run-all scr \
	build-context generate

# Binary output directory
BIN_DIR := bin

# Version info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Ldflags for version injection
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.GitCommit=$(COMMIT) -X main.BuildDate=$(BUILD_DATE)"

# Legacy binaries (backward compatibility)
LEGACY_BINARIES := scope-detector static-analysis ast-extractor call-graph data-flow compile-context run-all

all: build

# Build the unified CLI (primary target)
build: scr

# Build all binaries including legacy wrappers
build-all: scr build-legacy

# Build only legacy binaries
build-legacy: $(LEGACY_BINARIES)

# Build unified scr binary
scr:
	@echo "Building scr (unified CLI)..."
	@mkdir -p $(BIN_DIR)
	@go build $(LDFLAGS) -o $(BIN_DIR)/scr ./cmd/scr

# Pattern rule for building legacy phase binaries
# Replaces individual targets with identical echo/mkdir/go-build pattern
$(LEGACY_BINARIES):
	@echo "Building $@ (legacy)..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$@ ./cmd/$@

# Convenience target for Phase 5 binaries only
build-context: compile-context run-all
	@echo "Context binaries built."

test:
	@echo "Running tests..."
	@go test -v -race ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out coverage.html

install: build
	@echo "Installing scr to $(GOPATH)/bin..."
	@go install $(LDFLAGS) ./cmd/scr

install-all: build-all
	@echo "Installing all binaries..."
	@go install $(LDFLAGS) ./cmd/scr
	@chmod +x $(BIN_DIR)/*

# Development helpers
fmt:
	@go fmt ./...

vet:
	@go vet ./...

golangci-lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

lint: fmt vet golangci-lint

generate:
	@echo "Generating mocks..."
	@go generate ./...
