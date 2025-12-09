.PHONY: all build test clean fmt lint examples help

all: fmt build test

build:
	@echo "Building..."
	@go build ./...

test:
	@echo "Running tests..."
	@go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

fmt:
	@echo "Formatting code..."
	@go fmt ./...

lint:
	@echo "Running linter..."
	@golangci-lint run ./...

examples:
	@echo "Running basic example..."
	@cd examples/basic && go run main.go
	@echo "\nRunning transfer example..."
	@cd examples/transfer && go run main.go

clean:
	@echo "Cleaning..."
	@rm -f coverage.out coverage.html
	@go clean

deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

docs:
	@echo "Generating documentation..."
	@godoc -http=:6060

test-integration:
	@echo "Running integration tests..."
	@go test -v -tags=integration ./...

help:
	@echo "Available targets:"
	@echo "  all              - Format, build, and test (default)"
	@echo "  build            - Build the project"
	@echo "  test             - Run unit tests"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  test-integration - Run integration tests"
	@echo "  fmt              - Format code"
	@echo "  lint             - Run linter"
	@echo "  examples         - Run example programs"
	@echo "  clean            - Clean build artifacts"
	@echo "  deps             - Install dependencies"
	@echo "  docs             - Generate and serve documentation"
	@echo "  help             - Show this help message"
