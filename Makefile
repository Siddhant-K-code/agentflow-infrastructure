.PHONY: build test clean dev deploy install

# Build variables
BINARY_NAME=agentflow
BUILD_DIR=bin
GO_FILES=$(shell find . -name "*.go" -not -path "./vendor/*")

# Version and build info
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT ?= $(shell git rev-parse HEAD)

# Go build flags
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.Commit=${COMMIT}"

# Default target
all: build

# Build the CLI binary
build:
	@echo "Building AgentFlow CLI..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/agentflow

# Build all components
build-all: build build-orchestrator build-dashboard build-agents

# Build orchestrator
build-orchestrator:
	@echo "Building orchestrator..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/orchestrator ./core/orchestrator

# Build dashboard server
build-dashboard:
	@echo "Building dashboard server..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/dashboard-server ./dashboard

# Build mock agents
build-agents:
	@echo "Building mock agents..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/mock-agent ./agents/hello-world

# Build LLM router
build-llm-router:
	@echo "Building LLM router..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/llm-router ./core/llm-router

# Build Rust runtime
build-runtime:
	@echo "Building Rust runtime..."
	@cd core/runtime && cargo build --release
	@mkdir -p $(BUILD_DIR)
	@cp core/runtime/target/release/agentflow-runtime $(BUILD_DIR)/

# Install CLI globally
install: build
	@echo "Installing AgentFlow CLI..."
	go install $(LDFLAGS) ./cmd/agentflow

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@cd core/runtime && cargo clean

# Development environment
dev: build-all
	@echo "Starting development environment..."
	docker-compose -f deploy/docker/docker-compose.dev.yml up -d

# Stop development environment
dev-stop:
	@echo "Stopping development environment..."
	docker-compose -f deploy/docker/docker-compose.dev.yml down

# Deploy to Kubernetes
deploy:
	@echo "Deploying to Kubernetes..."
	kubectl apply -f deploy/kubernetes/

# Undeploy from Kubernetes
undeploy:
	@echo "Removing from Kubernetes..."
	kubectl delete -f deploy/kubernetes/

# Format code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...
	@cd core/runtime && cargo fmt

# Lint code
lint:
	@echo "Linting Go code..."
	golangci-lint run
	@cd core/runtime && cargo clippy

# Generate API documentation
docs:
	@echo "Generating API documentation..."
	@mkdir -p docs/api
	swag init -g cmd/agentflow/main.go -o docs/api

# Update dependencies
deps:
	@echo "Updating Go dependencies..."
	go mod tidy
	@echo "Updating Rust dependencies..."
	@cd core/runtime && cargo update

# POC Demo - Start all services locally
demo:
	@echo "🚀 Starting AgentFlow POC Demo..."
	@make build-all
	@echo ""
	@echo "Starting services..."
	@echo "📊 Dashboard: http://localhost:3001"
	@echo "🔧 Orchestrator: http://localhost:8080"
	@echo ""
	@echo "Press Ctrl+C to stop all services"
	@$(BUILD_DIR)/orchestrator & \
	$(BUILD_DIR)/dashboard-server & \
	wait

# Quick POC test
poc-test: build
	@echo "🧪 Running POC test..."
	@echo "Building and deploying hello-world example..."
	@$(BUILD_DIR)/agentflow deploy examples/hello-world-pipeline.yaml || echo "⚠️  Orchestrator might not be running"
# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the CLI binary"
	@echo "  build-all      - Build all components"
	@echo "  install        - Install CLI globally"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  clean          - Clean build artifacts"
	@echo "  dev            - Start development environment"
	@echo "  dev-stop       - Stop development environment"
	@echo "  demo           - Start POC demo (orchestrator + dashboard)"
	@echo "  poc-test       - Run quick POC test"
	@echo "  deploy         - Deploy to Kubernetes"
	@echo "  undeploy       - Remove from Kubernetes"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  docs           - Generate API documentation"
	@echo "  deps           - Update dependencies"
	@echo "  help           - Show this help"