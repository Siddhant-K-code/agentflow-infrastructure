# AgentFlow Infrastructure Makefile

.PHONY: help build test clean docker-build docker-push k8s-deploy k8s-delete dev-up dev-down

# Default target
help:
	@echo "AgentFlow Infrastructure"
	@echo "========================"
	@echo ""
	@echo "Available targets:"
	@echo "  build         - Build all binaries"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker images"
	@echo "  docker-push   - Push Docker images"
	@echo "  k8s-deploy    - Deploy to Kubernetes"
	@echo "  k8s-delete    - Delete from Kubernetes"
	@echo "  dev-up        - Start development environment"
	@echo "  dev-down      - Stop development environment"
	@echo "  lint          - Run linters"
	@echo "  fmt           - Format code"

# Variables
DOCKER_REGISTRY ?= agentflow
VERSION ?= latest
NAMESPACE ?= agentflow

# Go build variables
GOOS ?= linux
GOARCH ?= amd64
CGO_ENABLED ?= 0

# Build all binaries
build:
	@echo "Building binaries..."
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/control-plane ./cmd/control-plane
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/worker ./cmd/worker
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/agentctl ./cmd/agentctl

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	docker system prune -f

# Build Docker images
docker-build:
	@echo "Building Docker images..."
	docker build -f Dockerfile.control-plane -t $(DOCKER_REGISTRY)/control-plane:$(VERSION) .
	docker build -f Dockerfile.worker -t $(DOCKER_REGISTRY)/worker:$(VERSION) .

# Push Docker images
docker-push: docker-build
	@echo "Pushing Docker images..."
	docker push $(DOCKER_REGISTRY)/control-plane:$(VERSION)
	docker push $(DOCKER_REGISTRY)/worker:$(VERSION)

# Deploy to Kubernetes
k8s-deploy:
	@echo "Deploying to Kubernetes..."
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/postgres.yaml
	kubectl apply -f k8s/clickhouse.yaml
	kubectl apply -f k8s/redis.yaml
	kubectl apply -f k8s/nats.yaml
	kubectl apply -f k8s/control-plane.yaml
	kubectl apply -f k8s/workers.yaml
	@echo "Waiting for deployments to be ready..."
	kubectl wait --for=condition=available --timeout=300s deployment/control-plane -n $(NAMESPACE)
	kubectl wait --for=condition=available --timeout=300s deployment/workers -n $(NAMESPACE)

# Delete from Kubernetes
k8s-delete:
	@echo "Deleting from Kubernetes..."
	kubectl delete -f k8s/workers.yaml --ignore-not-found=true
	kubectl delete -f k8s/control-plane.yaml --ignore-not-found=true
	kubectl delete -f k8s/nats.yaml --ignore-not-found=true
	kubectl delete -f k8s/redis.yaml --ignore-not-found=true
	kubectl delete -f k8s/clickhouse.yaml --ignore-not-found=true
	kubectl delete -f k8s/postgres.yaml --ignore-not-found=true
	kubectl delete -f k8s/namespace.yaml --ignore-not-found=true

# Start development environment
dev-up:
	@echo "Starting development environment..."
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	sleep 30
	@echo "Development environment is ready!"
	@echo "Control Plane: http://localhost:8080"
	@echo "Grafana: http://localhost:3000 (admin/admin)"
	@echo "Prometheus: http://localhost:9090"

# Stop development environment
dev-down:
	@echo "Stopping development environment..."
	docker-compose down -v

# Run linters
lint:
	@echo "Running linters..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Initialize Go modules
mod-init:
	@echo "Initializing Go modules..."
	go mod tidy
	go mod verify

# Generate code
generate:
	@echo "Generating code..."
	go generate ./...

# Run database migrations
migrate-up:
	@echo "Running database migrations..."
	migrate -path migrations -database "postgres://agentflow:agentflow_password@localhost:5432/agentflow?sslmode=disable" up

# Rollback database migrations
migrate-down:
	@echo "Rolling back database migrations..."
	migrate -path migrations -database "postgres://agentflow:agentflow_password@localhost:5432/agentflow?sslmode=disable" down 1

# Create new migration
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations $$name

# Install development dependencies
dev-deps:
	@echo "Installing development dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run security scan
security:
	@echo "Running security scan..."
	gosec ./...

# Build and run locally
run-control-plane: build
	@echo "Running control plane locally..."
	./bin/control-plane

run-worker: build
	@echo "Running worker locally..."
	./bin/worker

# Install CLI globally
install-cli: build
	@echo "Installing agentctl CLI..."
	cp bin/agentctl /usr/local/bin/agentctl
	chmod +x /usr/local/bin/agentctl

# Uninstall CLI
uninstall-cli:
	@echo "Uninstalling agentctl CLI..."
	rm -f /usr/local/bin/agentctl

# Show logs from development environment
logs:
	docker-compose logs -f

# Show status of development environment
status:
	docker-compose ps

# Backup development data
backup:
	@echo "Backing up development data..."
	docker-compose exec postgres pg_dump -U agentflow agentflow > backup_$(shell date +%Y%m%d_%H%M%S).sql

# Restore development data
restore:
	@read -p "Enter backup file path: " file; \
	docker-compose exec -T postgres psql -U agentflow agentflow < $$file