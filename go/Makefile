# ACP Gateway Go - Makefile

.PHONY: build run test clean deps fmt vet

# Build the application
build:
	@echo "🔨 Building ACP Gateway Go..."
	go build -o bin/acp-gateway-go ./cmd/server

# Run the application
run:
	@echo "🚀 Starting ACP Gateway Go server..."
	go run ./cmd/server

# Run tests
test:
	@echo "🧪 Running tests..."
	go test ./...

# Run the test script
test-api:
	@echo "🧪 Testing API endpoints..."
	go run ./test/test_go.go

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/

# Download dependencies
deps:
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "✨ Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "🔍 Running go vet..."
	go vet ./...

# Install dependencies and build
install: deps build

# Development setup
dev-setup: deps fmt vet
	@echo "✅ Development environment ready!"

# Production build
prod-build:
	@echo "🏭 Building for production..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/acp-gateway-go ./cmd/server

# Help
help:
	@echo "Available targets:"
	@echo "  build      - Build the application"
	@echo "  run        - Run the application"
	@echo "  test       - Run tests"
	@echo "  test-api   - Test API endpoints"
	@echo "  clean      - Clean build artifacts"
	@echo "  deps       - Download dependencies"
	@echo "  fmt        - Format code"
	@echo "  vet        - Run go vet"
	@echo "  install    - Install dependencies and build"
	@echo "  dev-setup  - Setup development environment"
	@echo "  prod-build - Build for production"
	@echo "  help       - Show this help"
