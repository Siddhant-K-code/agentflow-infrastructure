# AgentFlow Infrastructure

A comprehensive agent orchestration platform with five interconnected pillars: Agent Orchestration Runtime (AOR), PromptOps Platform (POP), Secure Context Layer (SCL), Agent Observability Stack (AOS), and Cost-Aware Scheduler (CAS).

## 🏗️ Architecture Overview

AgentFlow provides a complete infrastructure for building, deploying, and managing AI agent workflows at scale. The platform is designed with five core pillars:

### 1. Agent Orchestration Runtime (AOR)
- **Purpose**: Fan-out/fan-in, retries, backpressure, cancellation, and idempotency for multi-agent DAGs
- **Features**: 
  - DAG-based workflow execution
  - Map-reduce patterns
  - Exactly-once semantics
  - Deterministic replay
  - WASI sandbox execution

### 2. PromptOps Platform (POP)
- **Purpose**: Versioned, testable prompts with evaluation, canary rollouts, and composability
- **Features**:
  - Template versioning and composition
  - Automated evaluation suites
  - Canary deployments with traffic splitting
  - Git-friendly workflow

### 3. Secure Context Layer (SCL)
- **Purpose**: Sanitize, validate, and authorize untrusted context
- **Features**:
  - Schema validation and content filtering
  - PII redaction with reversible tokenization
  - Prompt injection detection
  - Policy-based access control

### 4. Agent Observability Stack (AOS)
- **Purpose**: Semantic traces, diffs, replay, and root-cause analysis
- **Features**:
  - Distributed tracing with semantic context
  - Workflow replay and comparison
  - Cost and performance analytics
  - Quality drift detection

### 5. Cost-Aware Scheduler (CAS)
- **Purpose**: Keep LLM/API costs predictable while preserving quality/SLA
- **Features**:
  - Multi-armed bandit provider routing
  - Budget enforcement and quotas
  - Intelligent caching and batching
  - Cost optimization recommendations

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for development)
- Kubernetes cluster (for production deployment)

### Development Environment

1. **Clone the repository**:
   ```bash
   git clone https://github.com/Siddhant-K-code/agentflow-infrastructure.git
   cd agentflow-infrastructure
   ```

2. **Start the development environment**:
   ```bash
   make dev-up
   ```

3. **Install the CLI**:
   ```bash
   make build
   make install-cli
   ```

4. **Configure the CLI**:
   ```bash
   agentctl config init
   agentctl config login
   ```

### Your First Workflow

1. **Create a simple workflow**:
   ```bash
   cat > example-workflow.yaml << EOF
   name: document_analysis
   version: 1
   nodes:
     - id: ingest
       type: tool
       config:
         tool_name: s3.fetch
         tool_args:
           bucket: documents
           key: "{{document_key}}"
     - id: analyze
       type: llm
       config:
         prompt_ref: document_analyzer@3
         inputs:
           content: "ingest.output"
       policy:
         quality: Gold
         sla_ms: 30000
   edges:
     - from: ingest
       to: analyze
   EOF
   ```

2. **Submit the workflow**:
   ```bash
   agentctl workflow submit document_analysis \
     --inputs '{"document_key": "sample.pdf"}' \
     --budget 1000 \
     --wait
   ```

3. **Monitor execution**:
   ```bash
   agentctl workflow status <run-id>
   agentctl trace get <run-id>
   ```

## 📊 Monitoring and Observability

The platform includes comprehensive monitoring:

- **Grafana Dashboard**: http://localhost:3000 (admin/admin)
- **Prometheus Metrics**: http://localhost:9090
- **Control Plane API**: http://localhost:8080

### Key Metrics

- Workflow success rates and latency
- Cost breakdown by provider/model
- Cache hit rates and optimization opportunities
- Resource utilization and scaling metrics

## 🛠️ Development

### Project Structure

```
├── cmd/                    # Application entrypoints
│   ├── agentctl/          # CLI application
│   ├── control-plane/     # Control plane service
│   └── worker/            # Worker service
├── internal/              # Internal packages
│   ├── aor/              # Agent Orchestration Runtime
│   ├── pop/              # PromptOps Platform
│   ├── scl/              # Secure Context Layer
│   ├── aos/              # Agent Observability Stack
│   ├── cas/              # Cost-Aware Scheduler
│   ├── cli/              # CLI implementation
│   ├── config/           # Configuration management
│   └── db/               # Database abstractions
├── sdk/                   # Client SDKs
│   └── go/               # Go SDK
├── k8s/                  # Kubernetes manifests
├── monitoring/           # Monitoring configurations
├── migrations/           # Database migrations
└── docs/                 # Documentation
```

### Building and Testing

```bash
# Build all components
make build

# Run tests
make test

# Run linters
make lint

# Format code
make fmt
```

### Database Migrations

```bash
# Run migrations
make migrate-up

# Create new migration
make migrate-create

# Rollback migration
make migrate-down
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_USER` | PostgreSQL user | `agentflow` |
| `DB_PASSWORD` | PostgreSQL password | - |
| `CLICKHOUSE_HOST` | ClickHouse host | `localhost` |
| `REDIS_HOST` | Redis host | `localhost` |
| `NATS_URL` | NATS connection URL | `nats://localhost:4222` |

### Configuration File

Create `configs/config.yaml`:

```yaml
database:
  host: localhost
  port: 5432
  user: agentflow
  password: your_password
  database: agentflow
  ssl_mode: disable

clickhouse:
  host: localhost
  port: 9000
  user: agentflow
  password: your_password
  database: agentflow

redis:
  host: localhost
  port: 6379
  password: your_password
  db: 0

nats:
  url: nats://localhost:4222

server:
  host: 0.0.0.0
  port: 8080
```

## 🚢 Production Deployment

### Kubernetes

1. **Deploy infrastructure**:
   ```bash
   make k8s-deploy
   ```

2. **Verify deployment**:
   ```bash
   kubectl get pods -n agentflow
   kubectl get services -n agentflow
   ```

3. **Access the API**:
   ```bash
   kubectl port-forward svc/control-plane 8080:8080 -n agentflow
   ```

### Scaling

The platform supports horizontal scaling:

- **Control Plane**: Stateless, can run multiple replicas
- **Workers**: Auto-scaling based on queue depth and CPU/memory
- **Databases**: Use managed services (RDS, Cloud SQL) for production

## 📚 API Reference

### Workflow Management

```bash
# Submit workflow
POST /api/v1/workflows/runs
{
  "workflow_name": "document_analysis",
  "workflow_version": 1,
  "inputs": {"document_key": "sample.pdf"},
  "budget_cents": 1000
}

# Get workflow status
GET /api/v1/workflows/runs/{run_id}

# Cancel workflow
POST /api/v1/workflows/runs/{run_id}/cancel
```

### Prompt Management

```bash
# Create prompt version
POST /api/v1/prompts
{
  "name": "document_analyzer",
  "template": "Analyze: {{content}}",
  "schema": {"type": "object", "properties": {...}}
}

# Deploy prompt
POST /api/v1/prompts/deployments
{
  "prompt_name": "document_analyzer",
  "stable_version": 3,
  "canary_version": 4,
  "canary_ratio": 0.1
}
```

### Budget Management

```bash
# Create budget
POST /api/v1/budgets
{
  "period_type": "monthly",
  "limit_cents": 100000
}

# Get budget status
GET /api/v1/budgets/status
```

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

```bash
# Start test environment
make dev-up

# Run integration tests
go test -tags=integration ./tests/integration/...
```

### Load Testing

```bash
# Install k6
brew install k6  # macOS
# or
sudo apt install k6  # Ubuntu

# Run load tests
k6 run tests/load/workflow_submission.js
```

## 🔒 Security

### Authentication

The platform supports multiple authentication methods:

- **JWT tokens** for API access
- **OIDC integration** for enterprise SSO
- **Service accounts** for programmatic access

### Authorization

Fine-grained permissions using OpenFGA:

```yaml
# Example authorization model
model
  schema 1.1

type user

type organization
  relations
    define member: [user]
    define admin: [user]

type project
  relations
    define parent: [organization]
    define viewer: [user] or member from parent
    define editor: [user] or admin from parent
    define admin: [user] or admin from parent

type workflow
  relations
    define parent: [project]
    define viewer: [user] or viewer from parent
    define runner: [user] or editor from parent
    define editor: [user] or admin from parent
```

### Data Protection

- **Encryption at rest** for sensitive data
- **PII redaction** with reversible tokenization
- **Audit logging** for all operations
- **Network policies** for service isolation

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run the test suite
6. Submit a pull request

### Code Standards

- Follow Go best practices
- Write comprehensive tests
- Document public APIs
- Use conventional commits

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: [docs.agentflow.dev](https://docs.agentflow.dev)
- **Issues**: [GitHub Issues](https://github.com/Siddhant-K-code/agentflow-infrastructure/issues)
- **Discussions**: [GitHub Discussions](https://github.com/Siddhant-K-code/agentflow-infrastructure/discussions)
- **Slack**: [AgentFlow Community](https://agentflow.slack.com)

## 🗺️ Roadmap

### M1 (Weeks 0-4): Foundations ✅
- [x] CP skeleton, worker prototype, NATS wiring
- [x] Minimal DAG: map, reduce, retries, cancel
- [x] Basic model providers, token estimation
- [x] CLI + simple web UI

### M2 (Weeks 5-8): Scheduling & Cost
- [ ] CAS v1: provider routing, budgets, quotas, caching
- [ ] AOR integration with CAS
- [ ] Canary deployment object in POP

### M3 (Weeks 9-14): PromptOps + Secure Context
- [ ] POP v1: versioning, linting, eval suites, canary rollout
- [ ] SCL v1: schema validation, redaction, injection heuristics
- [ ] AOR uses POP and SCL integration

### M4 (Weeks 15-20): Observability
- [ ] AOS v1: trace timeline, semantic diff, replay
- [ ] Cost explorer dashboards
- [ ] Shadow routing in CAS

### M5 (Weeks 21-28): Hardening & Scale
- [ ] WASI sandbox for user steps
- [ ] Multi-region optionality
- [ ] Performance: >50k step/s across cluster

---

**Built with ❤️ by the AgentFlow team**