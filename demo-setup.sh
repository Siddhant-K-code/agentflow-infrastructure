#!/bin/bash

# AgentFlow Demo Setup Script
echo "🚀 Setting up AgentFlow Demo Environment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

print_status "Docker is running ✓"

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose is not installed. Please install Docker Compose and try again."
    exit 1
fi

print_status "Docker Compose is available ✓"

# Create demo environment file
print_status "Creating demo environment configuration..."
cat > .env.demo << EOF
# AgentFlow Demo Environment
DB_HOST=postgres
DB_USER=agentflow
DB_PASSWORD=agentflow_password
DB_NAME=agentflow
DB_SSL_MODE=disable

CLICKHOUSE_HOST=clickhouse
CLICKHOUSE_USER=agentflow
CLICKHOUSE_PASSWORD=agentflow_password
CLICKHOUSE_DB=agentflow

REDIS_HOST=redis
REDIS_PASSWORD=agentflow_password

NATS_URL=nats://nats:4222

# Demo API Keys (replace with real keys for production)
OPENAI_API_KEY=demo-key-replace-with-real
ANTHROPIC_API_KEY=demo-key-replace-with-real
GOOGLE_API_KEY=demo-key-replace-with-real
EOF

print_success "Environment file created"

# Start infrastructure services
print_status "Starting infrastructure services (PostgreSQL, Redis, ClickHouse, NATS)..."
docker-compose up -d postgres redis clickhouse nats

# Wait for services to be ready
print_status "Waiting for services to be ready..."
sleep 15

# Check if services are running
print_status "Checking service health..."

# Check PostgreSQL
if docker-compose exec -T postgres pg_isready -U agentflow > /dev/null 2>&1; then
    print_success "PostgreSQL is ready"
else
    print_error "PostgreSQL is not ready"
    exit 1
fi

# Check Redis
if docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; then
    print_success "Redis is ready"
else
    print_error "Redis is not ready"
    exit 1
fi

# Check ClickHouse
if curl -s http://localhost:8123/ping > /dev/null 2>&1; then
    print_success "ClickHouse is ready"
else
    print_error "ClickHouse is not ready"
    exit 1
fi

# Check NATS
if curl -s http://localhost:8222/healthz > /dev/null 2>&1; then
    print_success "NATS is ready"
else
    print_error "NATS is not ready"
    exit 1
fi

# Build the application
print_status "Building AgentFlow application..."
if go build -o bin/control-plane ./cmd/control-plane; then
    print_success "Control plane built successfully"
else
    print_error "Failed to build control plane"
    exit 1
fi

if go build -o bin/worker ./cmd/worker; then
    print_success "Worker built successfully"
else
    print_error "Failed to build worker"
    exit 1
fi

if go build -o bin/agentctl ./cmd/agentctl; then
    print_success "CLI built successfully"
else
    print_error "Failed to build CLI"
    exit 1
fi

# Start the control plane
print_status "Starting AgentFlow Control Plane..."
export $(cat .env.demo | xargs)
./bin/control-plane &
CONTROL_PLANE_PID=$!

# Wait for control plane to start
sleep 5

# Check if control plane is running
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    print_success "Control plane is running on http://localhost:8080"
else
    print_error "Control plane failed to start"
    kill $CONTROL_PLANE_PID 2>/dev/null
    exit 1
fi

# Start a worker
print_status "Starting AgentFlow Worker..."
./bin/worker &
WORKER_PID=$!

# Wait for worker to start
sleep 3

print_success "AgentFlow Demo Environment is ready!"
echo ""
echo "🎉 Demo URLs:"
echo "   • Control Plane API: http://localhost:8080"
echo "   • Demo UI: http://localhost:8080/"
echo "   • Health Check: http://localhost:8080/health"
echo "   • API Scenarios: http://localhost:8080/api/v1/demo/scenarios"
echo ""
echo "🔧 Infrastructure:"
echo "   • PostgreSQL: localhost:5432"
echo "   • Redis: localhost:6379"
echo "   • ClickHouse: localhost:8123"
echo "   • NATS: localhost:4222"
echo ""
echo "📝 Demo Commands:"
echo "   • Submit workflow: curl -X POST http://localhost:8080/api/v1/workflows/submit -H 'Content-Type: application/json' -d '{\"workflow_name\":\"document_analysis\",\"workflow_version\":1,\"inputs\":{\"document_url\":\"https://example.com/sample.pdf\"},\"budget_cents\":1000}'"
echo "   • Check status: curl http://localhost:8080/api/v1/workflows/{run_id}"
echo "   • List workflows: curl http://localhost:8080/api/v1/workflows"
echo ""
echo "🛑 To stop the demo:"
echo "   kill $CONTROL_PLANE_PID $WORKER_PID"
echo "   docker-compose down"
echo ""
echo "Press Ctrl+C to stop the demo..."

# Keep the script running and handle cleanup
trap 'echo ""; print_status "Stopping demo..."; kill $CONTROL_PLANE_PID $WORKER_PID 2>/dev/null; docker-compose down; print_success "Demo stopped"; exit 0' INT

# Wait for user to stop
wait