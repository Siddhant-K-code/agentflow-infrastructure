# AgentFlow Infrastructure

## Building Kubernetes for AI Agents

AgentFlow is a comprehensive platform for deploying and orchestrating multi-agent AI workflows. It provides a Kubernetes-like experience for AI agents, allowing you to deploy complex agent workflows using simple YAML configurations.

## Features

- **YAML-based Workflow Definition**: Deploy multi-agent workflows using docker-compose-style YAML files
- **DAG Execution Engine**: Go-based orchestrator with directed acyclic graph execution
- **WASM Sandboxing**: Rust runtime with WebAssembly sandboxing for secure agent execution
- **Multi-LLM Router**: Cost optimization through intelligent LLM routing
- **Distributed Tracing**: Full observability with distributed tracing capabilities
- **Time-travel Debugging**: Debug workflows by replaying historical states
- **Automatic Retries/Fallbacks**: Built-in resilience mechanisms
- **Multi-language SDKs**: Python and TypeScript SDKs for easy integration
- **Live Monitoring**: CLI with real-time workflow visualization

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Tool      │    │  Go Orchestrator│    │  Rust Runtime   │
│  (agentflow)    │────│   (DAG Engine)  │────│ (WASM Sandbox)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │              ┌─────────────────┐              │
         └──────────────│  LLM Router     │──────────────┘
                        │ (Cost Optimize) │
                        └─────────────────┘
                                 │
        ┌────────────────────────┼────────────────────────┐
        │                       │                        │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ NATS Messaging  │    │   PostgreSQL    │    │   ClickHouse    │
│   (Events)      │    │    (State)      │    │   (Metrics)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Quick Start

```bash
# Install AgentFlow CLI
go install github.com/Siddhant-K-code/agentflow-infrastructure/cmd/agentflow@latest

# Deploy a workflow
agentflow deploy workflow.yaml

# Monitor workflows
agentflow status

# View live workflow execution
agentflow live-view my-workflow
```

## Example Workflow

```yaml
# workflow.yaml
apiVersion: agentflow.io/v1
kind: Workflow
metadata:
  name: data-processing-pipeline
spec:
  agents:
    - name: data-collector
      image: agent:data-collector
      llm: 
        provider: openai
        model: gpt-4
      resources:
        memory: 512Mi
        cpu: 100m
    - name: data-processor
      image: agent:data-processor
      llm:
        provider: anthropic
        model: claude-3-sonnet
      dependsOn: [data-collector]
    - name: data-publisher
      image: agent:data-publisher
      llm:
        provider: openai
        model: gpt-3.5-turbo
      dependsOn: [data-processor]
  triggers:
    - schedule: "0 */6 * * *"
    - webhook: "/api/v1/trigger/data-pipeline"
```

## Development

```bash
# Build the project
make build

# Run tests
make test

# Start development environment
make dev

# Deploy to Kubernetes
make deploy
```

## Components

- **Core Orchestrator** (`core/orchestrator/`): Go-based DAG execution engine
- **Rust Runtime** (`core/runtime/`): WASM-based agent execution environment
- **LLM Router** (`core/llm-router/`): Multi-provider LLM routing and optimization
- **CLI Tool** (`cmd/agentflow/`): Command-line interface for managing workflows
- **SDKs** (`sdk/`): Python and TypeScript client libraries
- **Infrastructure** (`deploy/`): Kubernetes and Docker deployment configurations

## License

MIT License - see LICENSE file for details.