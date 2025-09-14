# AgentFlow Infrastructure

A comprehensive multi-agent orchestration platform with cost-aware scheduling, secure context handling, and advanced observability.

## Architecture

This project implements five core pillars:

1. **Agent Orchestration Runtime (AOR)** - Fan-out/fan-in, retries, backpressure, cancellation, and idempotency for multi-agent DAGs
2. **PromptOps Platform (POP)** - Versioned, testable prompts with evaluation and canary rollouts
3. **Secure Context Layer (SCL)** - Sanitize, validate, and authorize untrusted context
4. **Agent Observability Stack (AOS)** - Semantic traces, diffs, replay, and root-cause analysis
5. **Cost-Aware Scheduler (CAS)** - Keep LLM/API costs predictable while preserving quality/SLA

## Quick Start

```bash
# Start the infrastructure
make start

# Submit a workflow
agentctl workflow submit examples/doc_triage.yaml

# Monitor execution
agentctl run status <run-id>
```

## Development

```bash
# Install dependencies
go mod tidy

# Run tests
make test

# Build all components
make build

# Run locally
make dev
```

## Components

- `cmd/` - CLI tools and service binaries
- `internal/` - Core implementation packages
- `pkg/` - Public APIs and SDKs
- `deployments/` - Kubernetes manifests and Helm charts
- `examples/` - Sample workflows and configurations

## Documentation

See [docs/](./docs/) for detailed architecture and API documentation.