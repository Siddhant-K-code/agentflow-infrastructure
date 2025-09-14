# AgentFlow Architecture

## Overview

AgentFlow is a comprehensive multi-agent orchestration platform implementing five core pillars:

## 1. Agent Orchestration Runtime (AOR)

The AOR provides fan-out/fan-in, retries, backpressure, cancellation, and idempotency for multi-agent DAGs and map-reduce patterns.

### Key Components:
- **Control Plane**: API + scheduler + state machine (Go)
- **Workers**: Execute steps (LLM calls, functions, tools)
- **Queues**: NATS subjects per priority, per tenant
- **State**: Postgres (workflow spec, run state), Redis (in-flight), S3 (artifacts)

### API Endpoints:
- `POST /api/v1/workflows/{name}/versions` - Register workflow spec
- `POST /api/v1/runs` - Start workflow run
- `GET /api/v1/runs/{id}` - Get run status
- `POST /api/v1/runs/{id}/cancel` - Cancel run

## 2. PromptOps Platform (POP)

POP provides versioned, testable prompts with evaluation, canary rollouts, and composability.

### Features:
- Template engine with typed variables
- Evaluation runners (unit and dataset eval)
- Canary deployment manager
- Git-friendly export/import

### API Endpoints:
- `POST /api/v1/prompts/{name}/versions` - Create prompt version
- `GET /api/v1/prompts/{name}` - Get latest prompt
- `POST /api/v1/prompts/{name}/evaluate` - Run evaluation
- `POST /api/v1/deployments` - Deploy prompt version

## 3. Secure Context Layer (SCL)

SCL sanitizes, validates, and authorizes untrusted context passed to agents.

### Pipeline Stages:
1. Source validators
2. Schema validation  
3. Content filters
4. Policy enforcement
5. Context gating
6. Provenance tagging

## 4. Agent Observability Stack (AOS)

AOS provides semantic traces, diffs, replay, and root-cause analysis.

### Features:
- Trace UI with flame-graph-like DAG timeline
- Semantic diff between runs
- Deterministic replay capability
- Cost explorer and guardrails dashboard

## 5. Cost-Aware Scheduler (CAS)

CAS keeps LLM/API costs predictable while preserving quality/SLA.

### Capabilities:
- Provider routing by quality tier
- Request batching and caching
- Budget enforcement with graceful degradation
- Quota control per provider

## Data Models

### Workflows
```sql
workflow_spec: versioned DAG specifications
workflow_run: execution instances  
step_run: individual node executions
```

### Prompts
```sql
prompt_template: versioned prompt templates
prompt_suite: evaluation test suites
prompt_deployment: canary deployment configs
```

### Context
```sql
context_bundle: sanitized context with provenance
```

## Getting Started

1. **Setup Dependencies**
   ```bash
   make setup
   ```

2. **Start Services**
   ```bash
   make dev
   ```

3. **Submit Workflow**
   ```bash
   agentctl workflow submit examples/doc_triage.yaml
   ```

4. **Monitor Execution**
   ```bash
   agentctl run status <run-id>
   ```

## Configuration

Environment variables:
- `DATABASE_URL`: Postgres connection string
- `REDIS_URL`: Redis connection string  
- `NATS_URL`: NATS server URL
- `CLICKHOUSE_URL`: ClickHouse connection URL

## Development

See [Development Guide](./development.md) for detailed setup instructions.