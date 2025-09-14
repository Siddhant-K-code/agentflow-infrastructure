# AgentFlow API Reference

## Authentication

All API requests require authentication via Bearer token in the Authorization header:

```
Authorization: Bearer <token>
```

## Agent Orchestration Runtime (AOR)

### Workflows

#### Create Workflow Version
```http
POST /api/v1/workflows/{name}/versions
Content-Type: application/json

{
  "dag": {
    "nodes": [
      {
        "id": "step1",
        "type": "llm",
        "config": {...},
        "policy": {...}
      }
    ],
    "edges": [...]
  },
  "metadata": {...}
}
```

#### Get Workflow
```http
GET /api/v1/workflows/{name}/versions/{version}
```

### Runs

#### Start Run
```http
POST /api/v1/runs
Content-Type: application/json

{
  "workflow_name": "doc_triage",
  "workflow_version": 1,
  "inputs": {...},
  "tags": ["prod", "batch"],
  "budget_cents": 500
}
```

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "queued"
}
```

#### Get Run Status
```http
GET /api/v1/runs/{id}
```

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "workflow_spec_id": "...",
  "status": "running",
  "started_at": "2024-01-01T12:00:00Z",
  "cost_cents": 150,
  "steps": [...]
}
```

#### Cancel Run
```http
POST /api/v1/runs/{id}/cancel
Content-Type: application/json

{
  "reason": "User cancellation"
}
```

## PromptOps Platform (POP)

### Prompts

#### Create Prompt Version
```http
POST /api/v1/prompts/{name}/versions
Content-Type: application/json

{
  "template": "Hello {{name}}! Please {{task}}.",
  "schema": {
    "type": "object",
    "properties": {
      "name": {"type": "string"},
      "task": {"type": "string"}
    },
    "required": ["name", "task"]
  },
  "metadata": {
    "description": "Greeting prompt",
    "tags": ["greeting"]
  }
}
```

#### Get Prompt
```http
GET /api/v1/prompts/{name}
GET /api/v1/prompts/{name}/versions/{version}
```

#### Evaluate Prompt
```http
POST /api/v1/prompts/{name}/evaluate
Content-Type: application/json

{
  "prompt_id": "550e8400-e29b-41d4-a716-446655440000",
  "suite_id": "660e8400-e29b-41d4-a716-446655440000",
  "provider": "openai",
  "model": "gpt-3.5-turbo"
}
```

### Deployments

#### Deploy Prompt
```http
POST /api/v1/deployments
Content-Type: application/json

{
  "prompt_name": "greeting",
  "stable_version": 2,
  "canary_version": 3,
  "canary_ratio": 0.1
}
```

#### Get Deployment
```http
GET /api/v1/deployments/{name}
```

## Error Responses

All API endpoints return errors in this format:

```json
{
  "error": "Description of the error"
}
```

HTTP status codes:
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (missing/invalid auth)
- `404` - Not Found (resource doesn't exist)
- `500` - Internal Server Error

## Rate Limiting

API requests are rate limited per organization:
- 1000 requests per minute for workflow operations
- 100 requests per minute for run creation
- 10 evaluations per minute per prompt

## Pagination

List endpoints support pagination:

```http
GET /api/v1/runs?page=1&limit=20
```

Response includes pagination metadata:
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "has_more": true
  }
}
```