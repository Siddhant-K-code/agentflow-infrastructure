.PHONY: build test clean run dev docker-build docker-run

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME_SERVER=agentflow-server
BINARY_NAME_CLI=agentctl
BINARY_NAME_WORKER=agentflow-worker

# Build targets
all: build

build: build-server build-cli build-worker

build-server:
	$(GOBUILD) -o bin/$(BINARY_NAME_SERVER) -v ./cmd/server

build-cli:
	$(GOBUILD) -o bin/$(BINARY_NAME_CLI) -v ./cmd/agentctl

build-worker:
	$(GOBUILD) -o bin/$(BINARY_NAME_WORKER) -v ./cmd/worker

# Development
dev:
	@echo "Running in development mode..."
	@export DATABASE_URL="postgres://postgres:password@localhost:5432/agentflow?sslmode=disable" && \
	go run cmd/server/main.go -debug

run: build-server
	./bin/$(BINARY_NAME_SERVER)

# Testing
test:
	$(GOTEST) -v ./...

test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Linting
lint:
	golangci-lint run

fmt:
	$(GOCMD) fmt ./...

# Dependencies
deps:
	$(GOGET) -u ./...
	$(GOCMD) mod tidy

# Docker
docker-build:
	docker build -t agentflow:latest .

docker-run:
	docker-compose up -d

# Database
db-setup:
	@echo "Setting up development database..."
	docker run --name agentflow-postgres -e POSTGRES_PASSWORD=password -e POSTGRES_DB=agentflow -p 5432:5432 -d postgres:15
	sleep 5
	@echo "Creating schema..."
	PGPASSWORD=password psql -h localhost -U postgres -d agentflow -f scripts/schema.sql

db-migrate:
	PGPASSWORD=password psql -h localhost -U postgres -d agentflow -f scripts/schema.sql

db-reset:
	docker stop agentflow-postgres || true
	docker rm agentflow-postgres || true
	$(MAKE) db-setup

# Cleanup
clean:
	$(GOCLEAN)
	rm -rf bin/

# Installation
install: build
	@echo "Installing binaries to /usr/local/bin..."
	sudo cp bin/$(BINARY_NAME_SERVER) /usr/local/bin/
	sudo cp bin/$(BINARY_NAME_CLI) /usr/local/bin/
	sudo cp bin/$(BINARY_NAME_WORKER) /usr/local/bin/

# Generate protobuf (if we add gRPC later)
proto:
	@echo "Generating protobuf files..."
	# protoc commands would go here

# Start services for development
start-deps:
	docker-compose -f deployments/docker-compose.dev.yml up -d postgres redis nats clickhouse

stop-deps:
	docker-compose -f deployments/docker-compose.dev.yml down

# Full development setup
setup: deps db-setup start-deps
	@echo "Development environment ready!"

# Examples
example-workflow:
	$(MAKE) build-cli
	./bin/$(BINARY_NAME_CLI) workflow submit examples/doc_triage.yaml

# Help
help:
	@echo "Available targets:"
	@echo "  build       - Build all binaries"
	@echo "  test        - Run tests"
	@echo "  dev         - Run server in development mode"
	@echo "  setup       - Setup development environment"
	@echo "  db-setup    - Setup development database"
	@echo "  lint        - Run linter"
	@echo "  clean       - Clean build artifacts"
	@echo "  help        - Show this help"