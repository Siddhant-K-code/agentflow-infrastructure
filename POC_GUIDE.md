# AgentFlow POC Demo Guide

## Quick Start POC

This guide shows how to run a working AgentFlow POC with real workflow execution.

### 1. Build All Components

```bash
make build-all
```

This builds:
- `bin/agentflow` - CLI tool
- `bin/orchestrator` - Workflow orchestrator
- `bin/dashboard-server` - Web dashboard
- `bin/mock-agent` - Sample agent for testing

### 2. Start the POC Demo

```bash
make demo
```

This starts:
- 🔧 **Orchestrator** at http://localhost:8080
- 📊 **Dashboard** at http://localhost:3001

### 3. Deploy a Workflow

Open a new terminal and deploy the hello-world example:

```bash
# Deploy a sample workflow
./bin/agentflow deploy examples/hello-world-pipeline.yaml

# Check status
./bin/agentflow status

# View specific workflow status
./bin/agentflow status hello-world-pipeline
```

### 4. Monitor in Dashboard

Open http://localhost:3001 to see:
- Real-time workflow status
- Agent execution progress
- Success/failure metrics
- Live updates

## What's Working in this POC

### ✅ Core Features
- [x] YAML workflow parsing and validation
- [x] CLI deployment with backend integration
- [x] Workflow orchestrator with DAG execution
- [x] Mock agent execution with dependency resolution
- [x] Real-time status monitoring
- [x] Web dashboard with live updates
- [x] Proper error handling and retries

### 🔄 Example Workflow Execution
1. CLI parses and validates YAML
2. Orchestrator receives workflow via HTTP API
3. DAG executor analyzes dependencies
4. Agents execute in correct order (greeter → processor → publisher)
5. Status updates in real-time
6. Dashboard shows live progress

### 📊 Dashboard Features
- Workflow statistics (running, completed, failed)
- Live workflow table with status
- Auto-refresh every 30 seconds
- Error display with connection status

## Next Steps for Full Implementation

### 🚧 Enhance Agent Execution
- Replace mock agents with actual WASM runtime
- Add real LLM provider integration
- Implement agent-to-agent communication
- Add resource limits and sandboxing

### 🔧 Production Features
- Add persistent storage (PostgreSQL)
- Implement proper authentication
- Add distributed tracing
- Scale with Kubernetes
- Add metrics and alerting

### 📈 Advanced Workflow Features
- Time-travel debugging
- Workflow templates
- Conditional execution
- Event-driven triggers
- Cost tracking

## Testing the POC

```bash
# Quick test
make poc-test

# Manual testing
./bin/agentflow deploy examples/hello-world-pipeline.yaml
./bin/agentflow status hello-world-pipeline
./bin/agentflow status

# View in browser
open http://localhost:3001
```

## POC Value Demonstration

This POC shows:
1. **End-to-end workflow** from YAML → execution → monitoring
2. **Real orchestration** with dependency resolution
3. **Production-ready architecture** with proper APIs
4. **User experience** similar to Kubernetes/Docker Compose
5. **Monitoring & observability** with live dashboard

Perfect for demos, investor presentations, and validating the core concept!