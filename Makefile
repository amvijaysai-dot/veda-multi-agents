.PHONY: build test lint fmt clean coverage all

APP_NAME := veda-agent-runtime
BUILD_DIR := ./build
GO_FLAGS := -ldflags="-s -w"

all: lint test build

build:
	@echo "Building $(APP_NAME)..."
	go build $(GO_FLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./cmd/...

test:
	@echo "Running tests..."
	go test -count=1 -race ./... 2>&1

test-verbose:
	@echo "Running tests (verbose)..."
	go test -v -count=1 -race ./... 2>&1

test-short:
	@echo "Running short tests..."
	go test -short -count=1 -race ./... 2>&1

coverage:
	@echo "Running tests with coverage..."
	go test -count=1 -coverprofile=coverage.out ./... 2>&1
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint:
	@echo "Running linters..."
	golangci-lint run ./... 2>&1

fmt:
	@echo "Formatting code..."
	go fmt ./... 2>&1

vet:
	@echo "Running go vet..."
	go vet ./... 2>&1

clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out
	rm -f coverage.html

deps:
	@echo "Tidying dependencies..."
	go mod tidy 2>&1

check: fmt vet lint build test
	@echo "All checks passed!"

help:
	@echo "Available targets:"
	@echo "  all           - Run lint, test, and build"
	@echo "  build         - Build the binary"
	@echo "  test          - Run tests"
	@echo "  test-verbose  - Run tests (verbose)"
	@echo "  coverage      - Run tests with coverage report"
	@echo "  lint          - Run linters"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Tidy dependencies"
	@echo "  check         - Run all checks"