.PHONY: all build test test-coverage clean install fmt vet golangci-lint lint \
	scope-detector static-analysis ast-extractor call-graph data-flow compile-context run-all \
	build-context

# Binary output directory
BIN_DIR := bin

# All binaries to build
BINARIES := scope-detector static-analysis ast-extractor call-graph data-flow compile-context run-all

all: build

build: $(BINARIES)

# Pattern rule for building all phase binaries
# Replaces individual targets with identical echo/mkdir/go-build pattern
$(BINARIES):
	@echo "Building $@..."
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
	@echo "Installing binaries to $(BIN_DIR)..."
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
