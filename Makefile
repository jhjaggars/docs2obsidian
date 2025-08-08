# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint
KO=ko

# Container registry
REGISTRY?=ghcr.io/$(shell git config --get remote.origin.url | sed 's|.*github.com[:/]||; s|\.git$$||')

# Binary info
BINARY_NAME=pkm-sync
BINARY_PATH=./cmd
BUILD_DIR=./build

# Version info (can be overridden)
VERSION?=$(shell git describe --tags --always --dirty)
COMMIT?=$(shell git rev-parse HEAD)
DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)"

.PHONY: all build clean test coverage lint fmt gofmt gofumpt vet deps help ko-build ko-push ko-run ko-install security

## Build the binary
build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(BINARY_PATH)

## Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(BINARY_PATH)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(BINARY_PATH)

build-darwin:
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(BINARY_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(BINARY_PATH)

build-windows:
	mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(BINARY_PATH)

## Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)

## Run tests
test:
	$(GOTEST) -v -race ./...

## Run tests with coverage
coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## Run linter
lint:
	@if ! command -v $(GOLINT) > /dev/null; then \
		echo "golangci-lint not found. Installing..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin latest; \
	fi
	@$$($(GOCMD) env GOPATH)/bin/$(GOLINT) run --timeout=5m

## Format code with gofmt
fmt:
	$(GOFMT) -s -w .

## Check gofmt formatting (without changing files)
gofmt:
	@unformatted=$$($(GOFMT) -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "The following files are not gofmt'ed:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

## Format code with gofumpt (stricter than gofmt)
gofumpt:
	@if ! command -v gofumpt > /dev/null; then \
		echo "gofumpt not found. Installing..."; \
		$(GOGET) mvdan.cc/gofumpt@latest; \
	fi
	gofumpt -l -w .

## Run go vet
vet:
	$(GOCMD) vet ./...

## Install/update dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

## Install goimports if not present and format imports
imports:
	@if ! command -v goimports > /dev/null; then \
		echo "goimports not found. Installing..."; \
		$(GOGET) golang.org/x/tools/cmd/goimports; \
	fi
	@$$($(GOCMD) env GOPATH)/bin/goimports -w .

## Run security checks
security:
	@echo "Running Go vulnerability check..."
	@$(GOCMD) install golang.org/x/vuln/cmd/govulncheck@latest
	@$$($(GOCMD) env GOPATH)/bin/govulncheck ./...
	@echo "Running go vet security checks..."
	$(GOCMD) vet ./...

## Run all checks (formatting, linting, vetting, testing, security)
check: gofmt imports vet lint test security

## Install the binary
install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

## Run the application (requires setup first)
run:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(BINARY_PATH)
	./$(BINARY_NAME)

## Setup development environment
dev-setup:
	@echo "Setting up development environment..."
	@if ! command -v $(GOLINT) > /dev/null; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin latest; \
	fi
	@if ! command -v goimports > /dev/null; then \
		echo "Installing goimports..."; \
		$(GOGET) golang.org/x/tools/cmd/goimports; \
	fi
	@if ! command -v $(KO) > /dev/null; then \
		echo "Installing ko..."; \
		$(GOGET) github.com/google/ko; \
	fi
	@if ! command -v govulncheck > /dev/null; then \
		echo "Installing govulncheck..."; \
		$(GOGET) golang.org/x/vuln/cmd/govulncheck@latest; \
	fi
	@if ! command -v gofumpt > /dev/null; then \
		echo "Installing gofumpt..."; \
		$(GOGET) mvdan.cc/gofumpt@latest; \
	fi
	$(GOMOD) download
	@echo "Development environment setup complete!"

## Build container image with ko
ko-build:
	@if ! command -v $(KO) > /dev/null; then \
		echo "ko not found. Installing..."; \
		$(GOGET) github.com/google/ko; \
	fi
	VERSION=$(VERSION) COMMIT=$(COMMIT) DATE=$(DATE) $(KO) build --local $(BINARY_PATH)

## Build and push container image with ko
ko-push:
	@if ! command -v $(KO) > /dev/null; then \
		echo "ko not found. Installing..."; \
		$(GOGET) github.com/google/ko; \
	fi
	KO_DOCKER_REPO=$(REGISTRY) VERSION=$(VERSION) COMMIT=$(COMMIT) DATE=$(DATE) $(KO) build --bare $(BINARY_PATH)

## Run container image locally with ko
ko-run:
	@if ! command -v $(KO) > /dev/null; then \
		echo "ko not found. Installing..."; \
		$(GOGET) github.com/google/ko; \
	fi
	VERSION=$(VERSION) COMMIT=$(COMMIT) DATE=$(DATE) $(KO) run --local $(BINARY_PATH) -- --help

## Install ko if not present
ko-install:
	@if ! command -v $(KO) > /dev/null; then \
		echo "Installing ko..."; \
		$(GOGET) github.com/google/ko; \
	fi
	@echo "ko is installed at: $(shell command -v $(KO) 2>/dev/null || echo 'not found')"

## Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  build-all    - Build for all platforms"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  coverage     - Run tests with coverage report"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code with gofmt"
	@echo "  gofmt        - Check gofmt formatting (CI-friendly)"
	@echo "  gofumpt      - Format code with gofumpt (stricter)"
	@echo "  imports      - Format imports"
	@echo "  vet          - Run go vet"
	@echo "  check        - Run all checks (gofmt, imports, vet, lint, test, security)"
	@echo "  deps         - Install/update dependencies"
	@echo "  install      - Install binary to GOPATH/bin"
	@echo "  run          - Build and run the application"
	@echo "  ko-build     - Build container image locally with ko"
	@echo "  ko-push      - Build and push container image with ko"
	@echo "  ko-run       - Run container image locally with ko"
	@echo "  ko-install   - Install ko if not present"
	@echo "  security     - Run security checks (govulncheck + go vet)"
	@echo "  dev-setup    - Setup development environment"
	@echo "  help         - Show this help"

# Default target
all: check build