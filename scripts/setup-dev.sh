#!/bin/bash

set -e

echo "🚀 Setting up AgentFlow Development Environment"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running. Please start Docker first.${NC}"
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose >/dev/null 2>&1; then
    echo -e "${RED}❌ docker-compose is not installed. Please install it first.${NC}"
    exit 1
fi

echo -e "${YELLOW}📦 Starting dependencies (PostgreSQL, Redis, NATS, ClickHouse)...${NC}"

# Start infrastructure services
docker-compose -f deployments/docker-compose.dev.yml up -d

echo -e "${YELLOW}⏳ Waiting for services to be ready...${NC}"
sleep 10

# Wait for PostgreSQL to be ready
echo -e "${YELLOW}🐘 Waiting for PostgreSQL...${NC}"
until docker-compose -f deployments/docker-compose.dev.yml exec -T postgres pg_isready -U postgres >/dev/null 2>&1; do
    echo "PostgreSQL is not ready yet..."
    sleep 2
done

echo -e "${GREEN}✅ PostgreSQL is ready${NC}"

# Check if Go dependencies are available
echo -e "${YELLOW}📋 Installing Go dependencies...${NC}"
go mod tidy

# Build binaries
echo -e "${YELLOW}🔨 Building AgentFlow binaries...${NC}"
make build

echo -e "${GREEN}✅ AgentFlow development environment is ready!${NC}"
echo ""
echo "🎯 Quick Start:"
echo "  1. Start the server:    make dev"
echo "  2. Open dashboard:      http://localhost:8080"
echo "  3. Submit workflow:     ./bin/agentctl workflow submit examples/doc_triage.yaml"
echo "  4. Check status:        ./bin/agentctl status"
echo ""
echo "🛠️  Available commands:"
echo "  make dev                - Start development server"
echo "  make build              - Build all binaries"
echo "  make test               - Run tests"
echo "  make stop-deps          - Stop infrastructure services"
echo "  make db-reset           - Reset database"
echo ""
echo "📚 Documentation:"
echo "  API Docs:               http://localhost:8080/static/index.html"
echo "  Architecture:           docs/architecture.md"
echo ""

# Show service status
echo -e "${YELLOW}🔍 Service Status:${NC}"
docker-compose -f deployments/docker-compose.dev.yml ps

echo -e "${GREEN}🎉 Setup complete! Happy coding!${NC}"