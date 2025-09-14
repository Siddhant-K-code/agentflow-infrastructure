# Getting Started with AgentFlow

Welcome to AgentFlow! This guide will help you deploy your first AI agent workflow in just a few minutes.

## Prerequisites

Before you begin, ensure you have:

- **Docker** and **Docker Compose** installed
- **Go 1.21+** (if building from source)
- **kubectl** (for Kubernetes deployments)
- An **LLM API key** (OpenAI, Anthropic, or others)

## Quick Start

### 1. Install AgentFlow CLI

#### Option A: Download Binary (Recommended)
```bash
# macOS (Intel)
curl -L "https://github.com/Siddhant-K-code/agentflow-infrastructure/releases/latest/download/agentflow-darwin-amd64" -o agentflow
chmod +x agentflow && sudo mv agentflow /usr/local/bin/

# macOS (Apple Silicon)
curl -L "https://github.com/Siddhant-K-code/agentflow-infrastructure/releases/latest/download/agentflow-darwin-arm64" -o agentflow
chmod +x agentflow && sudo mv agentflow /usr/local/bin/

# Linux (x86_64)
curl -L "https://github.com/Siddhant-K-code/agentflow-infrastructure/releases/latest/download/agentflow-linux-amd64" -o agentflow
chmod +x agentflow && sudo mv agentflow /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/Siddhant-K-code/agentflow-infrastructure/releases/latest/download/agentflow-windows-amd64.exe" -OutFile "agentflow.exe"
```

#### Option B: Build from Source
```bash
git clone https://github.com/Siddhant-K-code/agentflow-infrastructure.git
cd agentflow-infrastructure
make build
sudo mv bin/agentflow /usr/local/bin/
```

#### Option C: Go Install
```bash
go install github.com/Siddhant-K-code/agentflow-infrastructure/cmd/agentflow@latest
```

### 2. Verify Installation

```bash
agentflow version
```

Expected output:
```
AgentFlow CLI v1.0.0
Build: abc123def456
Go version: go1.21.0
```

### 3. Start Development Environment

```bash
# Clone the repository (if not already done)
git clone https://github.com/Siddhant-K-code/agentflow-infrastructure.git
cd agentflow-infrastructure

# Start all services with Docker Compose
make dev
```

This command starts:
- PostgreSQL database
- NATS message queue  
- Redis cache
- AgentFlow orchestrator
- AgentFlow runtime
- LLM router service

Wait for all services to start (about 30-60 seconds). You should see:
```
✅ PostgreSQL ready
✅ NATS ready  
✅ Redis ready
✅ Orchestrator ready
✅ Runtime ready
✅ LLM Router ready
🚀 AgentFlow development environment running!
```

### 4. Configure API Keys

Set your LLM provider API keys:

```bash
# OpenAI
export OPENAI_API_KEY="sk-..."

# Anthropic (optional)
export ANTHROPIC_API_KEY="sk-ant-..."

# Configure AgentFlow
agentflow config set api.endpoint http://localhost:8080
agentflow config set llm.openai.api_key $OPENAI_API_KEY
agentflow config set llm.anthropic.api_key $ANTHROPIC_API_KEY
```

### 5. Deploy Your First Workflow

Create a simple workflow file:

```bash
cat > hello-world-workflow.yaml << EOF
apiVersion: agentflow.io/v1
kind: Workflow
metadata:
  name: hello-world
  description: "A simple greeting workflow"
spec:
  agents:
    - name: greeter
      image: agent:greeter-v1
      llm:
        provider: openai
        model: gpt-3.5-turbo
        temperature: 0.7
      resources:
        memory: 256Mi
        cpu: 100m
        timeout: 30s
      config:
        prompt: "Say hello and introduce yourself as an AI agent"
        
    - name: responder  
      image: agent:responder-v1
      llm:
        provider: openai
        model: gpt-3.5-turbo
        temperature: 0.5
      dependsOn: [greeter]
      resources:
        memory: 256Mi
        cpu: 100m
        timeout: 30s
      config:
        prompt: "Respond to the greeting in a friendly way"
        
  triggers:
    - webhook:
        path: /hello
        method: POST
EOF
```

Deploy the workflow:

```bash
agentflow deploy hello-world-workflow.yaml
```

Expected output:
```
✅ Workflow validated successfully
✅ Workflow deployed: hello-world
✅ Execution ID: exec-abc123def456
🔗 Webhook: http://localhost:8080/api/v1/webhooks/hello
```

### 6. Monitor Workflow Execution

Check the workflow status:

```bash
agentflow status hello-world
```

Output:
```
┌─────────────────┬──────────┬────────────┬──────────┬─────────┐
│ Agent           │ Status   │ Duration   │ Cost     │ Tokens  │
├─────────────────┼──────────┼────────────┼──────────┼─────────┤
│ greeter         │ ✅ Complete│ 2.1s      │ $0.002   │ 45      │
│ responder       │ ✅ Complete│ 1.8s      │ $0.003   │ 52      │
└─────────────────┴──────────┴────────────┴──────────┴─────────┘

Workflow: hello-world
Status: ✅ Completed
Total Duration: 4.2s
Total Cost: $0.005
```

### 7. View Live Execution (Optional)

For future workflow runs, watch live execution:

```bash
agentflow live-view hello-world
```

This opens an interactive terminal UI showing real-time workflow progress.

### 8. Trigger Workflow via Webhook

Test the webhook trigger:

```bash
curl -X POST http://localhost:8080/api/v1/webhooks/hello \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello from webhook!"}'
```

Response:
```json
{
  "execution_id": "exec-def456ghi789",
  "status": "started",
  "workflow": "hello-world",
  "webhook": "/hello"
}
```

### 9. View Logs

Check agent execution logs:

```bash
# View all logs for the workflow
agentflow logs hello-world

# View logs for specific agent
agentflow logs hello-world greeter

# Follow logs in real-time
agentflow logs hello-world -f
```

## Next Steps

Congratulations! You've successfully deployed and executed your first AgentFlow workflow. Here's what to explore next:

### Explore Example Workflows

```bash
# View available examples
ls examples/

# Deploy a more complex workflow
agentflow deploy examples/data-processing-pipeline.yaml
```

### Learn Key Concepts

1. **[Workflow Definition](workflow-definition.md)** - Learn YAML syntax and configuration options
2. **[Agent Development](agent-development.md)** - Build custom agents with SDKs
3. **[LLM Configuration](llm-configuration.md)** - Configure providers and optimize costs
4. **[Resource Management](resource-management.md)** - Set limits and monitor usage
5. **[Security Model](security-model.md)** - Understand sandboxing and permissions

### Advanced Features

1. **[Debugging Workflows](debugging.md)** - Debug and replay executions
2. **[Cost Optimization](cost-optimization.md)** - Minimize LLM expenses
3. **[Production Deployment](deployment-guide.md)** - Deploy to Kubernetes
4. **[Monitoring & Observability](monitoring.md)** - Set up comprehensive monitoring

### SDKs and Integration

1. **[Python SDK](python-sdk.md)** - Programmatic workflow management
2. **[TypeScript SDK](typescript-sdk.md)** - JavaScript/Node.js integration
3. **[REST API](api-reference.md)** - Direct API usage
4. **[CI/CD Integration](cicd-integration.md)** - Automate deployments

## Troubleshooting

### Common Issues

#### 1. "Connection refused" error
```bash
# Check if services are running
docker-compose ps

# Restart development environment
make dev-restart
```

#### 2. "Invalid API key" error
```bash
# Verify API key is set
agentflow config get llm.openai.api_key

# Test API key directly
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

#### 3. Workflow validation fails
```bash
# Validate workflow syntax
agentflow validate hello-world-workflow.yaml

# Check for common issues
agentflow lint hello-world-workflow.yaml
```

#### 4. Agents timeout or fail
```bash
# Check resource limits
agentflow describe workflow hello-world

# View detailed error logs
agentflow logs hello-world --level error
```

### Getting Help

- **Documentation**: [https://docs.agentflow.io](https://docs.agentflow.io)
- **GitHub Issues**: [https://github.com/Siddhant-K-code/agentflow-infrastructure/issues](https://github.com/Siddhant-K-code/agentflow-infrastructure/issues)
- **Discord Community**: [https://discord.gg/agentflow](https://discord.gg/agentflow)
- **Email Support**: support@agentflow.io

## What's Next?

Now that you have AgentFlow running locally, you might want to:

1. **Build Custom Agents**: Create agents tailored to your specific use cases
2. **Integrate with Your Services**: Connect workflows to your existing systems
3. **Deploy to Production**: Set up AgentFlow in your Kubernetes cluster
4. **Optimize Costs**: Configure intelligent LLM routing to minimize expenses
5. **Monitor at Scale**: Set up comprehensive observability for production workloads

Ready to build something amazing with AI agents? Let's go! 🚀