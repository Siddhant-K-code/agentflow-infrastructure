# AgentFlow Infrastructure

A comprehensive multi-agent orchestration platform with cost-aware scheduling, secure context handling, and advanced observability.

## 🚀 Quick Start (POC Ready!)

### One-Command Setup
```bash
# Clone and setup everything
git clone https://github.com/Siddhant-K-code/agentflow-infrastructure.git
cd agentflow-infrastructure
make setup
```

### Start the Platform
```bash
# Start all services and web dashboard
make dev
```

**🌐 Open http://localhost:8080 to access the AgentFlow Dashboard**

### Submit Your First Workflow
```bash
# Submit a simple text processing workflow
make example-workflow

# Or submit manually
./bin/agentctl workflow submit examples/simple_text_processing.yaml

# Start a workflow run
./bin/agentctl run start --workflow simple_text_processing --budget 500

# Check system status
./bin/agentctl status
```

## 🏗️ Architecture

This project implements five core pillars:

1. **Agent Orchestration Runtime (AOR)** - Fan-out/fan-in, retries, backpressure, cancellation, and idempotency for multi-agent DAGs
2. **PromptOps Platform (POP)** - Versioned, testable prompts with evaluation and canary rollouts
3. **Secure Context Layer (SCL)** - Sanitize, validate, and authorize untrusted context
4. **Agent Observability Stack (AOS)** - Semantic traces, diffs, replay, and root-cause analysis
5. **Cost-Aware Scheduler (CAS)** - Keep LLM/API costs predictable while preserving quality/SLA

## 📊 Web Dashboard Features

- **Real-time Monitoring**: Live workflow runs and system metrics
- **Cost Tracking**: Detailed cost breakdown by workflow and provider
- **Workflow Management**: Submit, monitor, and cancel workflows via web UI
- **System Health**: Database, queue, and worker status monitoring

## 🛠️ CLI Tool (agentctl)

Complete command-line interface for all operations:

```bash
# Workflow management
agentctl workflow submit examples/doc_triage.yaml
agentctl workflow list
agentctl workflow get doc_triage

# Run management  
agentctl run start --workflow doc_triage --budget 1000
agentctl run status <run-id>
agentctl run list
agentctl run cancel <run-id>

# System monitoring
agentctl status
```

## 🔧 Development Commands

```bash
# Setup development environment
make setup

# Build all components
make build

# Run tests
make test

# Start development server with hot reload
make dev

# Run demo workflow
make demo

# Stop all services
make stop-deps

# Reset database
make db-reset
```

## 📋 Example Workflows

### 1. Simple Text Processing (POC Ready)
- **File**: `examples/simple_text_processing.yaml`
- **Purpose**: Demonstrates sentiment analysis, keyword extraction, and summarization
- **Features**: Multi-step LLM pipeline with cost controls

### 2. Document Triage Pipeline
- **File**: `examples/doc_triage.yaml`  
- **Purpose**: Document classification and routing
- **Features**: Tool integration, conditional routing, quality tiers

### 3. Customer Support Pipeline
- **File**: `examples/customer_support_pipeline.yaml`
- **Purpose**: Automated ticket processing
- **Features**: Priority escalation, SLA enforcement, budget controls

## 🏭 Production Features

### Multi-Provider LLM Support
- OpenAI GPT models
- Anthropic Claude (configuration ready)
- Cost optimization across providers
- Quality tier enforcement (Bronze/Silver/Gold)

### Scalability & Reliability
- Horizontal scaling via containerized workers
- Circuit breakers and retry policies
- Budget caps with automatic degradation
- Distributed tracing for debugging

### Security & Compliance
- PII detection and redaction
- Content filtering for prompt injection
- Fine-grained access control (OpenFGA integration points)
- Audit trails and provenance tracking

## 🔌 API Reference

All services expose REST APIs under `/api/v1/`:

### Workflow APIs
- `POST /api/v1/workflows/{name}/versions` - Create workflow
- `GET /api/v1/workflows` - List workflows
- `GET /api/v1/workflows/{name}` - Get latest workflow

### Run APIs  
- `POST /api/v1/runs` - Start workflow run
- `GET /api/v1/runs` - List runs
- `GET /api/v1/runs/{id}` - Get run status
- `POST /api/v1/runs/{id}/cancel` - Cancel run

### System APIs
- `GET /health` - System health check
- `POST /api/v1/tasks/heartbeat` - Worker heartbeat
- `POST /api/v1/tasks/complete` - Complete task

## 🧪 Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Lint code
make lint

# Format code
make fmt
```

## 📦 Components

- `cmd/` - CLI tools and service binaries
  - `server/` - Main AgentFlow server
  - `agentctl/` - Command-line interface
  - `worker/` - Distributed task worker
- `internal/` - Core implementation packages
  - `aor/` - Agent Orchestration Runtime
  - `pop/` - PromptOps Platform  
  - `scl/` - Secure Context Layer
  - `aos/` - Agent Observability Stack
  - `cas/` - Cost-Aware Scheduler
  - `client/` - HTTP API client library
- `web/` - Web dashboard and UI
- `deployments/` - Docker Compose and Kubernetes manifests
- `examples/` - Sample workflows and configurations
- `docs/` - Architecture and API documentation

## 🌟 POC Highlights

**Ready for immediate demonstration:**

✅ **Functional Web Dashboard** - Complete monitoring interface  
✅ **Working CLI Tool** - Full workflow lifecycle management  
✅ **LLM Integration** - OpenAI API integration (with mock fallback)  
✅ **Cost Tracking** - Real-time cost monitoring and budget enforcement  
✅ **Multi-step Workflows** - Complex DAG execution with dependencies  
✅ **Worker Architecture** - Distributed task execution framework  
✅ **Database Schema** - Production-ready PostgreSQL schema  
✅ **Development Environment** - One-command Docker setup  

**Demo Workflow**: Submit the simple text processing workflow to see multi-step LLM orchestration with sentiment analysis, keyword extraction, and summarization.

## 🚀 Getting Started (Detailed)

### Prerequisites
- Docker and Docker Compose
- Go 1.21+ (for development)
- PostgreSQL client (optional, for direct DB access)

### Environment Variables (Optional)
```bash
export OPENAI_API_KEY="your-openai-key"  # For real LLM calls
export DATABASE_URL="postgres://..."     # Custom database URL
```

### Architecture Highlights

**Database Design:**
- PostgreSQL for transactional data (workflows, runs, metadata)
- ClickHouse for analytics and observability (schema provided)
- Redis for caching and session management
- NATS for distributed task queuing

**Scalability:**
- Stateless server design for horizontal scaling
- Distributed workers for task execution
- Message queue for loose coupling
- Database connection pooling

**Cost Management:**
- Real-time budget tracking
- Provider cost comparison
- Quality tier enforcement
- Automatic degradation policies

This infrastructure provides a solid foundation for multi-agent workflow orchestration at scale, ready for immediate POC deployment and production use.

## 📖 Documentation

- [Architecture Guide](./docs/architecture.md) - Detailed system design
- [API Reference](./docs/api.md) - Complete API documentation
- [Workflow Examples](./examples/) - Production workflow templates