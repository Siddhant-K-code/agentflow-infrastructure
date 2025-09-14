# AgentFlow API Reference

This document provides a comprehensive reference for the AgentFlow REST API and GraphQL endpoints.

## Base URL

**Development**: `http://localhost:8080/api/v1`  
**Production**: `https://api.agentflow.io/v1`

## Authentication

AgentFlow uses Bearer token authentication with JWT tokens.

### Authentication Header
```http
Authorization: Bearer <jwt_token>
```

### Get Authentication Token
```http
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-01-15T10:30:00Z",
  "user": {
    "id": "user-123",
    "email": "user@example.com",
    "name": "John Doe"
  }
}
```

## Core Resources

### Workflows

#### List Workflows
```http
GET /workflows
```

**Query Parameters:**
- `page` (integer): Page number (default: 1)
- `limit` (integer): Items per page (default: 20, max: 100)
- `status` (string): Filter by status (`active`, `inactive`, `error`)
- `search` (string): Search by name or description

**Response:**
```json
{
  "workflows": [
    {
      "id": "workflow-123",
      "name": "data-processing-pipeline",
      "version": "v1.0.0",
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "spec": {
        "agents": [...],
        "triggers": [...],
        "config": {...}
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "pages": 8
  }
}
```

#### Get Workflow
```http
GET /workflows/{workflow_id}
```

**Response:**
```json
{
  "id": "workflow-123",
  "name": "data-processing-pipeline",
  "version": "v1.0.0",
  "description": "Processes customer data through multiple AI agents",
  "status": "active",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "spec": {
    "agents": [
      {
        "name": "data-collector",
        "image": "agent:data-collector-v1",
        "llm": {
          "provider": "openai",
          "model": "gpt-4",
          "temperature": 0.1
        },
        "resources": {
          "memory": "512Mi",
          "cpu": "200m",
          "timeout": "300s"
        },
        "retry": {
          "max_attempts": 3,
          "backoff": "exponential"
        }
      }
    ],
    "triggers": [
      {
        "webhook": {
          "path": "/process-data",
          "method": "POST"
        }
      }
    ],
    "config": {
      "max_concurrent_executions": 10,
      "default_timeout": "600s"
    }
  },
  "stats": {
    "total_executions": 1250,
    "successful_executions": 1190,
    "failed_executions": 60,
    "average_duration": "45.2s",
    "total_cost": "$234.56"
  }
}
```

#### Create Workflow
```http
POST /workflows
Content-Type: application/json

{
  "name": "new-workflow",
  "version": "v1.0.0",
  "description": "A new AI workflow",
  "spec": {
    "agents": [...],
    "triggers": [...],
    "config": {...}
  }
}
```

**Response:**
```json
{
  "id": "workflow-456",
  "name": "new-workflow",
  "status": "active",
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### Update Workflow
```http
PUT /workflows/{workflow_id}
Content-Type: application/json

{
  "version": "v1.1.0",
  "spec": {
    "agents": [...],
    "triggers": [...],
    "config": {...}
  }
}
```

#### Delete Workflow
```http
DELETE /workflows/{workflow_id}
```

**Response:**
```json
{
  "message": "Workflow deleted successfully",
  "deleted_at": "2024-01-15T10:30:00Z"
}
```

### Executions

#### List Executions
```http
GET /workflows/{workflow_id}/executions
```

**Query Parameters:**
- `page` (integer): Page number
- `limit` (integer): Items per page
- `status` (string): Filter by status (`pending`, `running`, `completed`, `failed`)
- `from` (string): Start date (ISO 8601)
- `to` (string): End date (ISO 8601)

**Response:**
```json
{
  "executions": [
    {
      "id": "exec-789",
      "workflow_id": "workflow-123",
      "status": "completed",
      "input": {...},
      "output": {...},
      "started_at": "2024-01-15T10:30:00Z",
      "completed_at": "2024-01-15T10:32:15Z",
      "duration": "135s",
      "cost": "$0.156",
      "trace_id": "trace-abc123"
    }
  ],
  "pagination": {...}
}
```

#### Get Execution
```http
GET /executions/{execution_id}
```

**Response:**
```json
{
  "id": "exec-789",
  "workflow_id": "workflow-123",
  "status": "completed",
  "input": {
    "customer_id": "cust-123",
    "data_source": "api"
  },
  "output": {
    "processed_records": 150,
    "insights": [...],
    "recommendations": [...]
  },
  "error_message": null,
  "started_at": "2024-01-15T10:30:00Z",
  "completed_at": "2024-01-15T10:32:15Z",
  "duration": "135s",
  "cost": "$0.156",
  "trace_id": "trace-abc123",
  "agents": [
    {
      "name": "data-collector",
      "status": "completed",
      "started_at": "2024-01-15T10:30:00Z",
      "completed_at": "2024-01-15T10:30:45Z",
      "duration": "45s",
      "cost": "$0.023",
      "resource_usage": {
        "max_memory": "234Mi",
        "cpu_time": "12.3s",
        "io_operations": 156
      },
      "llm_calls": [
        {
          "provider": "openai",
          "model": "gpt-4",
          "tokens_used": 450,
          "cost": "$0.023",
          "latency": "1.2s"
        }
      ]
    }
  ]
}
```

#### Create Execution
```http
POST /workflows/{workflow_id}/executions
Content-Type: application/json

{
  "input": {
    "customer_id": "cust-456",
    "data_source": "file"
  },
  "config": {
    "timeout": "300s",
    "priority": "high"
  }
}
```

**Response:**
```json
{
  "id": "exec-101112",
  "workflow_id": "workflow-123",
  "status": "pending",
  "created_at": "2024-01-15T10:30:00Z",
  "input": {...}
}
```

#### Stop Execution
```http
POST /executions/{execution_id}/stop
```

**Response:**
```json
{
  "id": "exec-789",
  "status": "stopped",
  "stopped_at": "2024-01-15T10:31:00Z",
  "message": "Execution stopped by user request"
}
```

### Logs

#### Get Execution Logs
```http
GET /executions/{execution_id}/logs
```

**Query Parameters:**
- `agent` (string): Filter by agent name
- `level` (string): Log level (`debug`, `info`, `warn`, `error`)
- `limit` (integer): Max number of log entries
- `follow` (boolean): Stream logs (Server-Sent Events)

**Response:**
```json
{
  "logs": [
    {
      "timestamp": "2024-01-15T10:30:15.123Z",
      "level": "info",
      "agent": "data-collector",
      "message": "Starting data collection from API",
      "metadata": {
        "trace_id": "trace-abc123",
        "span_id": "span-def456"
      }
    },
    {
      "timestamp": "2024-01-15T10:30:16.456Z",
      "level": "debug",
      "agent": "data-collector", 
      "message": "Retrieved 150 records from API",
      "metadata": {
        "record_count": 150,
        "api_latency": "1.2s"
      }
    }
  ]
}
```

#### Stream Logs (Server-Sent Events)
```http
GET /executions/{execution_id}/logs?follow=true
Accept: text/event-stream
```

**Response Stream:**
```
data: {"timestamp":"2024-01-15T10:30:15.123Z","level":"info","agent":"data-collector","message":"Starting execution"}

data: {"timestamp":"2024-01-15T10:30:16.456Z","level":"debug","agent":"data-collector","message":"Processing batch 1/5"}

data: {"timestamp":"2024-01-15T10:30:18.789Z","level":"info","agent":"data-processor","message":"Agent started"}
```

### Metrics

#### Get Workflow Metrics
```http
GET /workflows/{workflow_id}/metrics
```

**Query Parameters:**
- `from` (string): Start time (ISO 8601)
- `to` (string): End time (ISO 8601)
- `granularity` (string): Time granularity (`minute`, `hour`, `day`)
- `metrics` (array): Specific metrics to retrieve

**Response:**
```json
{
  "time_range": {
    "from": "2024-01-01T00:00:00Z",
    "to": "2024-01-15T23:59:59Z"
  },
  "metrics": {
    "execution_count": {
      "total": 1250,
      "successful": 1190,
      "failed": 60,
      "success_rate": 0.952
    },
    "performance": {
      "average_duration": "45.2s",
      "p50_duration": "42.1s",
      "p95_duration": "78.5s",
      "p99_duration": "125.3s"
    },
    "cost": {
      "total": "$234.56",
      "average_per_execution": "$0.188",
      "by_provider": {
        "openai": "$198.34",
        "anthropic": "$36.22"
      }
    },
    "resource_usage": {
      "average_memory": "345Mi",
      "peak_memory": "512Mi",
      "average_cpu": "150m",
      "peak_cpu": "400m"
    }
  },
  "time_series": [
    {
      "timestamp": "2024-01-15T10:00:00Z",
      "executions": 12,
      "success_rate": 0.958,
      "average_duration": "43.2s",
      "cost": "$2.34"
    }
  ]
}
```

### Webhooks

#### List Webhooks
```http
GET /webhooks
```

**Response:**
```json
{
  "webhooks": [
    {
      "id": "webhook-123",
      "workflow_id": "workflow-123",
      "path": "/process-data",
      "method": "POST",
      "url": "https://api.agentflow.io/v1/webhooks/process-data",
      "created_at": "2024-01-01T00:00:00Z",
      "stats": {
        "total_calls": 450,
        "successful_calls": 425,
        "failed_calls": 25
      }
    }
  ]
}
```

#### Get Webhook
```http
GET /webhooks/{webhook_id}
```

#### Trigger Webhook
```http
POST /webhooks/{path}
Content-Type: application/json

{
  "data": {...},
  "metadata": {...}
}
```

## GraphQL API

AgentFlow also provides a GraphQL endpoint for complex queries and real-time subscriptions.

**Endpoint**: `/graphql`

### Schema Overview

```graphql
type Query {
  workflows(filter: WorkflowFilter, pagination: Pagination): WorkflowConnection
  workflow(id: ID!): Workflow
  executions(filter: ExecutionFilter, pagination: Pagination): ExecutionConnection
  execution(id: ID!): Execution
  metrics(filter: MetricsFilter): Metrics
}

type Mutation {
  createWorkflow(input: CreateWorkflowInput!): Workflow
  updateWorkflow(id: ID!, input: UpdateWorkflowInput!): Workflow
  deleteWorkflow(id: ID!): Boolean
  createExecution(workflowId: ID!, input: CreateExecutionInput!): Execution
  stopExecution(id: ID!): Execution
}

type Subscription {
  executionUpdates(workflowId: ID): ExecutionUpdate
  logs(executionId: ID!, filter: LogFilter): LogEntry
  metrics(workflowId: ID!, interval: Duration!): MetricsUpdate
}
```

### Example Queries

#### Get Workflow with Recent Executions
```graphql
query GetWorkflowDetails($id: ID!) {
  workflow(id: $id) {
    id
    name
    status
    spec {
      agents {
        name
        llm {
          provider
          model
        }
        resources {
          memory
          cpu
        }
      }
    }
    executions(limit: 10, orderBy: CREATED_AT_DESC) {
      edges {
        node {
          id
          status
          duration
          cost
          createdAt
        }
      }
    }
    metrics {
      executionCount {
        total
        successRate
      }
      performance {
        averageDuration
        p95Duration
      }
    }
  }
}
```

#### Subscribe to Execution Updates
```graphql
subscription ExecutionUpdates($workflowId: ID!) {
  executionUpdates(workflowId: $workflowId) {
    execution {
      id
      status
      duration
    }
    agent {
      name
      status
      progress
    }
    timestamp
  }
}
```

## Error Handling

AgentFlow uses standard HTTP status codes and provides detailed error information.

### Error Response Format
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Workflow validation failed",
    "details": [
      {
        "field": "spec.agents[0].llm.model",
        "message": "Model 'gpt-5' is not supported"
      }
    ],
    "request_id": "req-123456",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

### Common Error Codes

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | `VALIDATION_ERROR` | Request validation failed |
| 401 | `AUTHENTICATION_ERROR` | Invalid or missing authentication |
| 403 | `AUTHORIZATION_ERROR` | Insufficient permissions |
| 404 | `NOT_FOUND` | Resource not found |
| 409 | `CONFLICT` | Resource conflict (e.g., duplicate name) |
| 422 | `WORKFLOW_ERROR` | Workflow definition error |
| 429 | `RATE_LIMIT_ERROR` | Rate limit exceeded |
| 500 | `INTERNAL_ERROR` | Internal server error |
| 503 | `SERVICE_UNAVAILABLE` | Service temporarily unavailable |

## Rate Limiting

API requests are rate limited to ensure fair usage and system stability.

### Rate Limit Headers
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 995
X-RateLimit-Reset: 1642262400
X-RateLimit-Window: 3600
```

### Rate Limits by Plan

| Plan | Requests/Hour | Concurrent Executions |
|------|---------------|----------------------|
| Free | 1,000 | 5 |
| Professional | 10,000 | 50 |
| Enterprise | 100,000 | 500 |

## SDKs

Official SDKs are available for popular programming languages:

- **Python**: `pip install agentflow-sdk`
- **TypeScript/JavaScript**: `npm install @agentflow/sdk`
- **Go**: `go get github.com/Siddhant-K-code/agentflow-infrastructure/sdk/go`

### SDK Examples

#### Python
```python
from agentflow import AgentFlowClient

client = AgentFlowClient(
    endpoint="https://api.agentflow.io/v1",
    token="your-jwt-token"
)

# Create workflow
workflow = client.create_workflow({
    "name": "my-workflow",
    "spec": {...}
})

# Create execution
execution = client.create_execution(
    workflow_id=workflow.id,
    input={"key": "value"}
)

# Monitor execution
for update in client.stream_execution(execution.id):
    print(f"Status: {update.status}")
```

#### TypeScript
```typescript
import { AgentFlowClient } from '@agentflow/sdk';

const client = new AgentFlowClient({
  endpoint: 'https://api.agentflow.io/v1',
  token: 'your-jwt-token'
});

// Create and monitor execution
const execution = await client.createExecution(workflowId, {
  input: { key: 'value' }
});

const status = await client.getExecution(execution.id);
console.log(`Status: ${status.status}`);
```

## API Versioning

AgentFlow uses URL-based versioning:

- **Current Version**: `v1`
- **Base URL**: `https://api.agentflow.io/v1`
- **Deprecation Policy**: 12 months notice before version deprecation

### Version History

| Version | Release Date | Status | End of Life |
|---------|--------------|--------|-------------|
| v1 | 2024-01-01 | Current | TBD |
| v1-beta | 2023-12-01 | Deprecated | 2024-12-01 |

## Support

For API support and questions:

- **Documentation**: [https://docs.agentflow.io](https://docs.agentflow.io)
- **GitHub Issues**: [https://github.com/Siddhant-K-code/agentflow-infrastructure/issues](https://github.com/Siddhant-K-code/agentflow-infrastructure/issues)
- **Email**: api-support@agentflow.io
- **Discord**: [https://discord.gg/agentflow](https://discord.gg/agentflow)