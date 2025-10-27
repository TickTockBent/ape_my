.PHONY: build test clean run install help

# Binary name
BINARY_NAME=ape_my
BINARY_PATH=bin/$(BINARY_NAME)

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_PATH) ./cmd/ape_my

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -cover ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@go clean

# Run the application (requires schema.json)
run: build
	@./$(BINARY_PATH)

# Install dependencies
install:
	@echo "Installing dependencies..."
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Run all checks (fmt, vet, test)
check: fmt vet test

# Display help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Remove build artifacts"
	@echo "  run           - Build and run the application"
	@echo "  install       - Install dependencies"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  check         - Run fmt, vet, and test"
	@echo "  help          - Display this help message"
