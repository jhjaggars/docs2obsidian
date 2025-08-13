# Makefile for pkm-sync
# ----------------------
# This file provides local commands to ensure code quality and correctness,
# mirroring the checks run in the CI/CD pipeline.

# Go parameters
GO_PACKAGES := ./...
GO_BUILD_CMD := go build -v $(GO_PACKAGES)
GO_TEST_CMD := go test -v -race $(GO_PACKAGES)
GOLANGCI_LINT := golangci-lint

# Default target: Run all CI checks.
.PHONY: all
all: ci

# Target: ci - Runs all the checks that are performed in the CI pipeline.
# This is the command you should run locally before pushing code.
.PHONY: ci
ci: lint test build
	@echo "✅ All CI checks passed."

# Target: lint - Runs the golangci-lint linter.
# It uses the .golangci.yml configuration file for its settings.
.PHONY: lint
lint:
	@echo "🔍 Running linters..."
	@$(GOLANGCI_LINT) run ./... --timeout=5m

# Target: test - Runs unit tests with the race detector.
.PHONY: test
test:
	@echo "🧪 Running unit tests..."
	@$(GO_TEST_CMD)

# Target: build - Compiles the Go project to ensure it builds correctly.
.PHONY: build
build:
	@echo "🏗️ Building project..."
	@$(GO_BUILD_CMD)

# Target: tidy - Tidies up the go.mod and go.sum files.
.PHONY: tidy
tidy:
	@echo "🧹 Tidying go modules..."
	@go mod tidy

# Target: help - Displays help for the Makefile targets.
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all     - Run all CI checks (default)."
	@echo "  ci      - Alias for 'all'."
	@echo "  lint    - Run golangci-lint."
	@echo "  test    - Run unit tests."
	@echo "  build   - Build the project."
	@echo "  tidy    - Tidy go modules."
	@echo "  help    - Show this help message."

