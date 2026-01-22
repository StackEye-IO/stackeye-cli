# StackEye CLI Makefile
# Build, test, and lint targets for the CLI

.PHONY: all build build-all build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-linux-arm64 build-windows-amd64 \
        build-local install clean test test-verbose test-e2e test-integration coverage coverage-html lint fmt fmt-check vet validate validate-quick tidy help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOVET=$(GOCMD) vet
GOFMT=gofmt
GOMOD=$(GOCMD) mod
GOINSTALL=$(GOCMD) install

# Binary configuration
BINARY_NAME=stackeye
BINARY_DIR=bin
CMD_DIR=./cmd/stackeye

# Version information (injected via ldflags)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short=7 HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Package path for ldflags injection
VERSION_PKG=github.com/StackEye-IO/stackeye-cli/internal/cmd

# Build flags
LDFLAGS=-ldflags "-s -w \
	-X $(VERSION_PKG).Version=$(VERSION) \
	-X $(VERSION_PKG).GitCommit=$(GIT_COMMIT) \
	-X $(VERSION_PKG).BuildTime=$(BUILD_TIME)"
BUILD_FLAGS=-trimpath

# Coverage output
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Tools (install if missing)
GOLINT=$(shell which golangci-lint 2>/dev/null || echo "go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest")

# Packages
PACKAGES=./...

#==============================================================================
# Default target
#==============================================================================

all: validate build

#==============================================================================
# Build targets
#==============================================================================

## build: Build CLI binary for current platform
build:
	@echo "Building $(BINARY_NAME)..."
	@echo "  Version:    $(VERSION)"
	@echo "  Commit:     $(GIT_COMMIT)"
	@echo "  Build Time: $(BUILD_TIME)"
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Binary: $(BINARY_DIR)/$(BINARY_NAME)"

## build-all: Build binaries for all supported platforms
build-all: build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-linux-arm64 build-windows-amd64
	@echo ""
	@echo "All binaries built in $(BINARY_DIR)/"
	@ls -la $(BINARY_DIR)/

## build-darwin-amd64: Build for macOS Intel
build-darwin-amd64:
	@echo "Building for darwin/amd64..."
	@mkdir -p $(BINARY_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)

## build-darwin-arm64: Build for macOS Apple Silicon
build-darwin-arm64:
	@echo "Building for darwin/arm64..."
	@mkdir -p $(BINARY_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)

## build-linux-amd64: Build for Linux x86_64
build-linux-amd64:
	@echo "Building for linux/amd64..."
	@mkdir -p $(BINARY_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)

## build-linux-arm64: Build for Linux ARM64
build-linux-arm64:
	@echo "Building for linux/arm64..."
	@mkdir -p $(BINARY_DIR)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)

## build-windows-amd64: Build for Windows x86_64
build-windows-amd64:
	@echo "Building for windows/amd64..."
	@mkdir -p $(BINARY_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)

## build-local: Build with race detector for local development
build-local:
	@echo "Building with race detector..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -race -o $(BINARY_DIR)/$(BINARY_NAME) $(CMD_DIR)

## install: Install CLI to GOBIN
install:
	@echo "Installing $(BINARY_NAME) to GOBIN..."
	$(GOINSTALL) $(BUILD_FLAGS) $(LDFLAGS) $(CMD_DIR)
	@echo "Installed: $$(which $(BINARY_NAME) 2>/dev/null || echo '$(GOPATH)/bin/$(BINARY_NAME)')"

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)/
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@echo "Clean complete"

#==============================================================================
# Test targets
#==============================================================================

## test: Run all tests with race detector
test:
	@echo "Running tests..."
	$(GOTEST) -race $(PACKAGES)

## test-verbose: Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	$(GOTEST) -race -v $(PACKAGES)

## test-e2e: Run end-to-end tests (requires binary to be built first)
test-e2e: build
	@echo "Running E2E tests..."
	$(GOTEST) -race -v ./test/e2e/...

## test-integration: Run integration tests against live dev API (requires login first)
test-integration: build
	@echo "Running integration tests against live API..."
	@echo "Note: Run 'stackeye login --api-url https://api-dev.stackeye.io' first"
	STACKEYE_E2E_LIVE=true $(GOTEST) -race -v -tags=integration -count=1 ./test/e2e/...

## coverage: Generate coverage report
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic $(PACKAGES)
	@echo "Coverage report: $(COVERAGE_FILE)"
	@$(GOCMD) tool cover -func=$(COVERAGE_FILE) | tail -1

## coverage-html: Generate HTML coverage report
coverage-html: coverage
	@echo "Generating HTML coverage report..."
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "HTML report: $(COVERAGE_HTML)"

#==============================================================================
# Code quality targets
#==============================================================================

## lint: Run golangci-lint
lint:
	@echo "Running linter..."
	$(GOLINT) run $(PACKAGES)

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

## fmt-check: Check code formatting (fails if not formatted)
fmt-check:
	@echo "Checking code format..."
	@UNFORMATTED=$$($(GOFMT) -s -l .); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "The following files are not formatted:"; \
		echo "$$UNFORMATTED"; \
		echo "Run 'make fmt' to fix"; \
		exit 1; \
	fi

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) $(PACKAGES)

## tidy: Run go mod tidy
tidy:
	@echo "Tidying modules..."
	$(GOMOD) tidy

#==============================================================================
# Validation targets
#==============================================================================

## validate: Full validation (fmt-check, vet, lint, test)
validate: fmt-check vet lint test
	@echo "Validation complete"

## validate-quick: Quick validation (fmt-check, vet, lint - no tests)
validate-quick: fmt-check vet lint
	@echo "Quick validation complete"

#==============================================================================
# Help
#==============================================================================

## help: Show this help message
help:
	@echo "StackEye CLI Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build targets:"
	@grep -E '^## build|^## install|^## clean' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""
	@echo "Test targets:"
	@grep -E '^## test|^## coverage' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""
	@echo "Code quality targets:"
	@grep -E '^## (lint|fmt|vet|tidy)' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""
	@echo "Validation targets:"
	@grep -E '^## validate' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""
	@echo "Cross-compilation:"
	@echo "  make build-all           Build for all platforms"
	@echo "  make build-darwin-amd64  Build for macOS Intel"
	@echo "  make build-darwin-arm64  Build for macOS Apple Silicon"
	@echo "  make build-linux-amd64   Build for Linux x86_64"
	@echo "  make build-linux-arm64   Build for Linux ARM64"
	@echo "  make build-windows-amd64 Build for Windows x86_64"
	@echo ""
	@echo "Version info:"
	@echo "  VERSION=$(VERSION)"
	@echo "  GIT_COMMIT=$(GIT_COMMIT)"
