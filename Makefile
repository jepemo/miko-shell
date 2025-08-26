BINARY_NAME=miko-shell
BINARY_PATH=./$(BINARY_NAME)
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build clean test install uninstall run demo help

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build -ldflags="-X 'github.com/jepemo/miko-shell/cmd.version=$(VERSION)'" -o $(BINARY_NAME) .
	@echo "Built $(BINARY_NAME) successfully!"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)
	rm -rf .bootstrap

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Install to system PATH
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(BINARY_NAME) /usr/local/bin/
	@echo "$(BINARY_NAME) installed successfully!"

# Uninstall from system PATH
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(BINARY_NAME) uninstalled successfully!"

# Run the demo
demo: build
	@echo "Running demo..."
	./demo.sh

# Initialize a miko-shell project in current directory
init: build
	./$(BINARY_NAME) init

# Development targets
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

lint:
	@echo "Running linters..."
	golangci-lint run

coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

dev-setup:
	@echo "Setting up development environment..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "Development environment ready!"

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X 'github.com/jepemo/miko-shell/cmd.version=$(VERSION)'" -o $(BUILD_DIR)/miko-shell-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X 'github.com/jepemo/miko-shell/cmd.version=$(VERSION)'" -o $(BUILD_DIR)/miko-shell-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X 'github.com/jepemo/miko-shell/cmd.version=$(VERSION)'" -o $(BUILD_DIR)/miko-shell-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X 'github.com/jepemo/miko-shell/cmd.version=$(VERSION)'" -o $(BUILD_DIR)/miko-shell-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X 'github.com/jepemo/miko-shell/cmd.version=$(VERSION)'" -o $(BUILD_DIR)/miko-shell-windows-amd64.exe .

# Run all checks (used in CI)
check: fmt lint test

# Help target
help:
	@echo "Available targets:"
	@echo "  build      - Build the application"
	@echo "  clean      - Clean build artifacts"
	@echo "  test       - Run tests"
	@echo "  coverage   - Run tests with coverage"
	@echo "  lint       - Run linters"
	@echo "  fmt        - Format code"
	@echo "  deps       - Download dependencies"
	@echo "  dev-setup  - Setup development environment"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  check      - Run all checks (fmt, lint, test)"
	@echo "  install    - Install to system PATH"
	@echo "  uninstall  - Uninstall from system PATH"
	@echo "  demo       - Run demo"
	@echo "  init       - Initialize project"
