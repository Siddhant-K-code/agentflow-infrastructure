# AgentFlow Architecture

## Overview

AgentFlow is designed as a cloud-native, microservices-based platform that brings Kubernetes-like orchestration capabilities to AI agent workloads. The architecture emphasizes security, scalability, observability, and cost optimization while maintaining developer simplicity.

## Core Architecture Principles

### 1. Microservices Design
- **Separation of Concerns**: Each service has a single, well-defined responsibility
- **Independent Scaling**: Components can scale independently based on load
- **Technology Diversity**: Best tool for each job (Go for orchestration, Rust for runtime)
- **Fault Isolation**: Failures in one component don't cascade to others

### 2. Event-Driven Architecture
- **Async Communication**: Services communicate via events rather than direct calls
- **Loose Coupling**: Services can evolve independently
- **Scalability**: Natural load distribution and horizontal scaling
- **Observability**: Complete event audit trail

### 3. Security-First Design
- **Zero Trust**: No implicit trust between components
- **Sandboxing**: Agent code runs in isolated WASM environments
- **Principle of Least Privilege**: Minimal permissions for each component
- **Defense in Depth**: Multiple security layers

### 4. Cloud-Native Patterns
- **Container-Ready**: All components designed for containerization
- **12-Factor App**: Stateless, config-driven, portable applications
- **Health Checks**: Built-in readiness and liveness probes
- **Graceful Degradation**: Continues operation under partial failures

## System Architecture

```
                                    ┌─────────────────────────────────────┐
                                    │            Load Balancer            │
                                    │         (ingress-nginx)             │
                                    └─────────────┬───────────────────────┘
                                                  │
                                    ┌─────────────▼───────────────────────┐
                                    │         API Gateway                 │
                                    │     (Kong/Ambassador)               │
                                    └─────────────┬───────────────────────┘
                                                  │
                ┌─────────────────────────────────┼─────────────────────────────────┐
                │                                 │                                 │
    ┌───────────▼────────────┐        ┌──────────▼─────────────┐        ┌─────────▼──────────┐
    │     CLI Clients        │        │    Web Dashboard       │        │   External APIs    │
    │   (agentflow CLI)      │        │    (Next.js SPA)       │        │  (Webhooks/REST)   │
    └────────────────────────┘        └────────────────────────┘        └────────────────────┘
                │                                 │                                 │
                └─────────────────────────────────┼─────────────────────────────────┘
                                                  │
                                    ┌─────────────▼───────────────────────┐
                                    │       Core Orchestrator             │
                                    │         (Go Service)                │
                                    │                                     │
                                    │  ┌─────────────────────────────┐    │
                                    │  │     API Server              │    │
                                    │  │   (REST + GraphQL)          │    │
                                    │  └─────────────────────────────┘    │
                                    │  ┌─────────────────────────────┐    │
                                    │  │   Workflow Engine           │    │
                                    │  │   (DAG Execution)           │    │
                                    │  └─────────────────────────────┘    │
                                    │  ┌─────────────────────────────┐    │
                                    │  │    Scheduler                │    │
                                    │  │  (Task Distribution)        │    │
                                    │  └─────────────────────────────┘    │
                                    └─────────────┬───────────────────────┘
                                                  │
                                    ┌─────────────▼───────────────────────┐
                                    │      Message Bus (NATS)             │
                                    │   ┌─────────────────────────────┐   │
                                    │   │     Event Streaming         │   │
                                    │   └─────────────────────────────┘   │
                                    │   ┌─────────────────────────────┐   │
                                    │   │    Pub/Sub Messaging        │   │
                                    │   └─────────────────────────────┘   │
                                    └─────────────┬───────────────────────┘
                                                  │
                ┌─────────────────────────────────┼─────────────────────────────────┐
                │                                 │                                 │
    ┌───────────▼────────────┐        ┌──────────▼─────────────┐        ┌─────────▼──────────┐
    │   Agent Runtime        │        │     LLM Router         │        │   Storage Layer    │
    │   (Rust Service)       │        │   (Go Service)         │        │                    │
    │                        │        │                        │        │                    │
    │ ┌────────────────────┐ │        │ ┌────────────────────┐ │        │ ┌────────────────┐ │
    │ │  WASM Executor     │ │        │ │ Provider Router    │ │        │ │  PostgreSQL    │ │
    │ │                    │ │        │ │                    │ │        │ │  (State Store) │ │
    │ └────────────────────┘ │        │ └────────────────────┘ │        │ └────────────────┘ │
    │ ┌────────────────────┐ │        │ ┌────────────────────┐ │        │ ┌────────────────┐ │
    │ │  Resource Monitor  │ │        │ │  Cost Optimizer    │ │        │ │   ClickHouse   │ │
    │ │                    │ │        │ │                    │ │        │ │   (Metrics)    │ │
    │ └────────────────────┘ │        │ └────────────────────┘ │        │ └────────────────┘ │
    │ ┌────────────────────┐ │        │ ┌────────────────────┐ │        │ ┌────────────────┐ │
    │ │   Security Box     │ │        │ │  Rate Limiter      │ │        │ │     Redis      │ │
    │ │                    │ │        │ │                    │ │        │ │    (Cache)     │ │
    │ └────────────────────┘ │        │ └────────────────────┘ │        │ └────────────────┘ │
    └────────────────────────┘        └────────────────────────┘        └────────────────────┘
                │                                 │                                 │
                └─────────────────────────────────┼─────────────────────────────────┘
                                                  │
                                    ┌─────────────▼───────────────────────┐
                                    │     Observability Stack            │
                                    │                                     │
                                    │ ┌─────────────────────────────┐     │
                                    │ │        Jaeger               │     │
                                    │ │   (Distributed Tracing)     │     │
                                    │ └─────────────────────────────┘     │
                                    │ ┌─────────────────────────────┐     │
                                    │ │      Prometheus             │     │
                                    │ │    (Metrics Collection)     │     │
                                    │ └─────────────────────────────┘     │
                                    │ ┌─────────────────────────────┐     │
                                    │ │       Grafana               │     │
                                    │ │   (Visualization)           │     │
                                    │ └─────────────────────────────┘     │
                                    └─────────────────────────────────────┘
```

## Component Deep Dive

### Core Orchestrator (Go)

**Purpose**: Central coordination engine for workflow execution and management.

#### Sub-Components

##### 1. API Server
```go
// REST API for workflow management
type WorkflowAPI struct {
    router     *gin.Engine
    validator  *validator.Validate
    auth       *auth.Manager
    metrics    *prometheus.Registry
}

// GraphQL API for complex queries
type GraphQLServer struct {
    schema     graphql.Schema
    resolver   *resolvers.Root
    middleware []graphql.Middleware
}
```

**Responsibilities**:
- REST and GraphQL API endpoints
- Request validation and authentication
- Rate limiting and throttling
- API documentation generation

##### 2. Workflow Engine
```go
type WorkflowEngine struct {
    dag        *DAG
    executor   *execution.Engine
    state      *state.Manager
    events     *events.Publisher
}

type DAG struct {
    nodes      map[string]*Node
    edges      []*Edge
    topology   [][]string  // Topologically sorted execution order
}
```

**Responsibilities**:
- YAML workflow parsing and validation
- DAG construction and dependency resolution
- Execution planning and optimization
- State transitions and event emission

##### 3. Scheduler
```go
type Scheduler struct {
    queue      *priority.Queue
    workers    *worker.Pool
    ratelimit  *ratelimiter.Manager
    metrics    *metrics.Collector
}
```

**Responsibilities**:
- Task queuing and prioritization
- Worker pool management
- Load balancing across runtime nodes
- Execution scheduling optimization

#### Data Models

```go
type Workflow struct {
    ID          string                 `json:"id" db:"id"`
    Name        string                 `json:"name" db:"name"`
    Version     string                 `json:"version" db:"version"`
    Spec        WorkflowSpec           `json:"spec" db:"spec"`
    Status      WorkflowStatus         `json:"status" db:"status"`
    CreatedAt   time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

type WorkflowSpec struct {
    Agents      []AgentSpec            `json:"agents" yaml:"agents"`
    Triggers    []TriggerSpec          `json:"triggers" yaml:"triggers"`
    Config      WorkflowConfig         `json:"config" yaml:"config"`
}

type AgentSpec struct {
    Name        string                 `json:"name" yaml:"name"`
    Image       string                 `json:"image" yaml:"image"`
    LLM         LLMConfig              `json:"llm" yaml:"llm"`
    Resources   ResourceRequirements   `json:"resources" yaml:"resources"`
    DependsOn   []string               `json:"dependsOn" yaml:"dependsOn"`
    Retry       RetryConfig            `json:"retry" yaml:"retry"`
}
```

### Agent Runtime (Rust)

**Purpose**: Secure, high-performance execution environment for AI agents.

#### Sub-Components

##### 1. WASM Executor
```rust
pub struct WasmExecutor {
    engine: Engine,
    runtime: Runtime,
    store: Store<RuntimeState>,
    limiter: ResourceLimiter,
}

pub struct RuntimeState {
    memory_usage: AtomicU64,
    cpu_usage: AtomicU64,
    io_operations: AtomicU64,
    network_calls: AtomicU64,
}
```

**Responsibilities**:
- WASM module loading and compilation
- Runtime environment setup and teardown
- Resource monitoring and enforcement
- Host function binding for controlled I/O

##### 2. Security Sandbox
```rust
pub struct SecuritySandbox {
    namespace: Namespace,
    capabilities: CapabilitySet,
    network_policy: NetworkPolicy,
    file_system: VirtualFS,
}

pub struct CapabilitySet {
    can_network: bool,
    can_filesystem: bool,
    can_spawn_process: bool,
    allowed_hosts: Vec<String>,
    allowed_paths: Vec<PathBuf>,
}
```

**Responsibilities**:
- Capability-based security model
- Network access control and filtering
- File system isolation and virtualization
- System call interception and filtering

##### 3. Resource Monitor
```rust
pub struct ResourceMonitor {
    limits: ResourceLimits,
    usage: Arc<ResourceUsage>,
    enforcer: ResourceEnforcer,
    metrics: MetricsCollector,
}

pub struct ResourceLimits {
    max_memory: u64,
    max_cpu_time: Duration,
    max_execution_time: Duration,
    max_network_requests: u32,
    max_file_operations: u32,
}
```

**Responsibilities**:
- Real-time resource usage tracking
- Limit enforcement and violation handling
- Performance metrics collection
- Resource optimization recommendations

### LLM Router (Go)

**Purpose**: Intelligent routing and cost optimization for LLM API calls.

#### Sub-Components

##### 1. Provider Router
```go
type ProviderRouter struct {
    providers   map[string]Provider
    selector    *ProviderSelector
    circuitbreaker *breaker.CircuitBreaker
    metrics     *metrics.Collector
}

type Provider interface {
    Name() string
    Models() []string
    Cost(model string, tokens int) float64
    RateLimit() *RateLimit
    Call(ctx context.Context, req *LLMRequest) (*LLMResponse, error)
}
```

**Responsibilities**:
- Multi-provider abstraction layer
- Request routing based on cost/performance criteria
- Circuit breaker pattern for failure handling
- Provider health monitoring

##### 2. Cost Optimizer
```go
type CostOptimizer struct {
    pricing     *PricingEngine
    analytics   *CostAnalytics
    budgets     *BudgetManager
    alerts      *AlertManager
}

type OptimizationStrategy struct {
    PrimaryCriteria   string  // cost, latency, quality
    Fallbacks        []FallbackRule
    BudgetLimits     BudgetConstraints
    QualityThreshold float64
}
```

**Responsibilities**:
- Cost calculation and tracking
- Budget management and alerting
- Provider selection optimization
- Historical cost analysis

##### 3. Rate Limiter
```go
type RateLimiter struct {
    buckets     map[string]*TokenBucket
    redis       *redis.Client
    strategies  map[string]LimitStrategy
}

type LimitStrategy interface {
    Allow(key string) bool
    Remaining(key string) int
    Reset(key string) time.Time
}
```

**Responsibilities**:
- Per-provider rate limiting
- Global rate limiting across instances
- Adaptive rate limiting based on errors
- Fair queuing for multiple workflows

### Message Bus (NATS)

**Purpose**: Event-driven communication and coordination between services.

#### Event Types

```go
type EventType string

const (
    WorkflowCreated     EventType = "workflow.created"
    WorkflowStarted     EventType = "workflow.started"
    WorkflowCompleted   EventType = "workflow.completed"
    WorkflowFailed      EventType = "workflow.failed"
    
    AgentStarted        EventType = "agent.started"
    AgentCompleted      EventType = "agent.completed"
    AgentFailed         EventType = "agent.failed"
    
    ResourceLimitHit    EventType = "resource.limit_hit"
    CostThresholdMet    EventType = "cost.threshold_met"
)

type Event struct {
    ID          string                 `json:"id"`
    Type        EventType              `json:"type"`
    Source      string                 `json:"source"`
    Subject     string                 `json:"subject"`
    Timestamp   time.Time              `json:"timestamp"`
    Data        map[string]interface{} `json:"data"`
    Trace       TraceContext           `json:"trace"`
}
```

#### Stream Configuration

```go
// Workflow execution events
WorkflowStream := &nats.StreamConfig{
    Name:        "WORKFLOWS",
    Subjects:    []string{"workflow.*", "agent.*"},
    Retention:   nats.WorkQueuePolicy,
    MaxAge:      24 * time.Hour,
    Storage:     nats.FileStorage,
    Replicas:    3,
}

// System metrics and monitoring
MetricsStream := &nats.StreamConfig{
    Name:        "METRICS", 
    Subjects:    []string{"metrics.*", "resource.*", "cost.*"},
    Retention:   nats.LimitsPolicy,
    MaxAge:      7 * 24 * time.Hour,
    Storage:     nats.FileStorage,
    Replicas:    3,
}
```

### Storage Layer

#### PostgreSQL (Primary State Store)

**Schema Design**:

```sql
-- Workflows and executions
CREATE TABLE workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    spec JSONB NOT NULL,
    status workflow_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    UNIQUE(name, version)
);

CREATE TABLE workflow_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID REFERENCES workflows(id),
    status execution_status NOT NULL DEFAULT 'pending',
    input JSONB,
    output JSONB,
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    trace_id VARCHAR(255),
    
    INDEX idx_executions_workflow_status (workflow_id, status),
    INDEX idx_executions_started_at (started_at),
    INDEX idx_executions_trace_id (trace_id)
);

CREATE TABLE agent_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_execution_id UUID REFERENCES workflow_executions(id),
    agent_name VARCHAR(255) NOT NULL,
    status execution_status NOT NULL DEFAULT 'pending',
    input JSONB,
    output JSONB,
    error_message TEXT,
    resource_usage JSONB,
    llm_calls JSONB,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    
    INDEX idx_agent_executions_workflow (workflow_execution_id),
    INDEX idx_agent_executions_status (status),
    INDEX idx_agent_executions_agent_name (agent_name)
);
```

#### ClickHouse (Metrics and Analytics)

**Schema Design**:

```sql
-- Execution metrics
CREATE TABLE execution_metrics (
    timestamp DateTime64(3),
    workflow_id String,
    execution_id String,
    agent_name String,
    metric_name String,
    metric_value Float64,
    dimensions Map(String, String)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, workflow_id, execution_id, agent_name, metric_name);

-- Cost tracking
CREATE TABLE cost_metrics (
    timestamp DateTime64(3),
    workflow_id String,
    execution_id String,
    agent_name String,
    provider String,
    model String,
    tokens_used UInt32,
    cost_usd Float64,
    request_count UInt32
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, workflow_id, provider, model);

-- Resource usage
CREATE TABLE resource_metrics (
    timestamp DateTime64(3),
    workflow_id String,
    execution_id String,
    agent_name String,
    cpu_usage_percent Float64,
    memory_usage_bytes UInt64,
    network_bytes_in UInt64,
    network_bytes_out UInt64,
    io_operations UInt32
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, workflow_id, execution_id, agent_name);
```

#### Redis (Caching and Session Storage)

**Usage Patterns**:

```go
// Workflow execution state caching
type ExecutionCache struct {
    redis *redis.Client
}

func (c *ExecutionCache) CacheExecution(exec *WorkflowExecution) error {
    key := fmt.Sprintf("execution:%s", exec.ID)
    data, _ := json.Marshal(exec)
    return c.redis.Set(key, data, 15*time.Minute).Err()
}

// LLM response caching
type LLMCache struct {
    redis *redis.Client
}

func (c *LLMCache) CacheResponse(hash string, response *LLMResponse) error {
    key := fmt.Sprintf("llm:response:%s", hash)
    data, _ := json.Marshal(response)
    return c.redis.Set(key, data, 1*time.Hour).Err()
}
```

### Observability Stack

#### Distributed Tracing (Jaeger)

**Trace Structure**:

```go
type TraceContext struct {
    TraceID    string `json:"trace_id"`
    SpanID     string `json:"span_id"`
    ParentID   string `json:"parent_id,omitempty"`
    Operation  string `json:"operation"`
    Service    string `json:"service"`
    StartTime  time.Time `json:"start_time"`
    Duration   time.Duration `json:"duration"`
    Tags       map[string]interface{} `json:"tags"`
    Logs       []LogEntry `json:"logs"`
}

// Workflow-level trace
WorkflowTrace := &TraceContext{
    Operation: "workflow.execute",
    Service:   "orchestrator",
    Tags: map[string]interface{}{
        "workflow.id":   "workflow-123",
        "workflow.name": "data-pipeline",
        "user.id":       "user-456",
    },
}

// Agent-level spans
AgentSpan := &TraceContext{
    Operation: "agent.execute",
    Service:   "runtime",
    ParentID:  WorkflowTrace.SpanID,
    Tags: map[string]interface{}{
        "agent.name":    "data-processor",
        "agent.image":   "agents/processor:v1.0",
        "resource.cpu":  "200m",
        "resource.memory": "512Mi",
    },
}
```

#### Metrics Collection (Prometheus)

**Custom Metrics**:

```go
var (
    // Workflow metrics
    WorkflowExecutionsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "agentflow_workflow_executions_total",
            Help: "Total number of workflow executions",
        },
        []string{"workflow", "status"},
    )
    
    WorkflowExecutionDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "agentflow_workflow_execution_duration_seconds",
            Help: "Workflow execution duration in seconds",
            Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
        },
        []string{"workflow"},
    )
    
    // Agent metrics
    AgentExecutionsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "agentflow_agent_executions_total",
            Help: "Total number of agent executions",
        },
        []string{"workflow", "agent", "status"},
    )
    
    // LLM metrics
    LLMCallsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "agentflow_llm_calls_total",
            Help: "Total number of LLM API calls",
        },
        []string{"provider", "model", "status"},
    )
    
    LLMCostTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "agentflow_llm_cost_usd_total",
            Help: "Total LLM costs in USD",
        },
        []string{"provider", "model"},
    )
    
    // Resource metrics
    AgentResourceUsage = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "agentflow_agent_resource_usage",
            Help: "Agent resource usage",
        },
        []string{"workflow", "agent", "resource"},
    )
)
```

## Security Architecture

### Defense in Depth

#### 1. Network Layer Security
- **TLS Encryption**: All service communication encrypted with TLS 1.3
- **Network Policies**: Kubernetes NetworkPolicies for traffic segmentation
- **Service Mesh**: Istio for mTLS and traffic management
- **DDoS Protection**: Rate limiting and traffic analysis

#### 2. Application Layer Security
- **Authentication**: JWT tokens with RS256 signing
- **Authorization**: RBAC with fine-grained permissions
- **Input Validation**: Comprehensive request validation and sanitization
- **SQL Injection Prevention**: Parameterized queries and ORM usage

#### 3. Runtime Security
- **WASM Sandboxing**: Complete isolation of agent code execution
- **Capability System**: Minimal permissions for each operation
- **Resource Limits**: CPU, memory, and I/O constraints
- **System Call Filtering**: seccomp-bpf for host protection

#### 4. Data Security
- **Encryption at Rest**: Database and file system encryption
- **Encryption in Transit**: TLS for all network communication
- **Key Management**: HashiCorp Vault for secret management
- **Data Classification**: Automatic PII detection and handling

### Threat Model

#### Threats and Mitigations

| Threat | Impact | Likelihood | Mitigation |
|--------|--------|------------|------------|
| Malicious agent code | High | Medium | WASM sandboxing, capability restrictions |
| API abuse/DDoS | Medium | High | Rate limiting, authentication, monitoring |
| Data exfiltration | High | Low | Network policies, audit logging, encryption |
| Privilege escalation | High | Low | RBAC, least privilege, container security |
| Supply chain attacks | Medium | Medium | Image scanning, signed artifacts, SBOM |

## Scalability and Performance

### Horizontal Scaling

#### Auto-scaling Strategies

```yaml
# Orchestrator HPA
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: orchestrator-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: orchestrator
  minReplicas: 3
  maxReplicas: 50
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  - type: Pods
    pods:
      metric:
        name: workflow_queue_depth
      target:
        type: AverageValue
        averageValue: "10"
```

#### Load Balancing

```go
type LoadBalancer interface {
    SelectNode(workflow *Workflow) (*RuntimeNode, error)
}

type WeightedRoundRobin struct {
    nodes   []*RuntimeNode
    weights []int
    current int
}

type LeastConnections struct {
    nodes []*RuntimeNode
}

type ResourceAware struct {
    nodes []*RuntimeNode
    metrics *MetricsCollector
}
```

### Performance Optimization

#### 1. Workflow Execution Optimization
- **Parallel Execution**: Independent agents run concurrently
- **Dependency Optimization**: Minimal blocking through smart scheduling
- **Resource Pooling**: Shared WASM instances for similar agents
- **Caching**: LLM response caching and workflow state caching

#### 2. Database Optimization
- **Read Replicas**: Separate read and write workloads
- **Connection Pooling**: Efficient database connection management
- **Query Optimization**: Indexed queries and materialized views
- **Partitioning**: Time-based partitioning for metrics tables

#### 3. Network Optimization
- **Compression**: gRPC compression for inter-service communication
- **Connection Reuse**: HTTP/2 and persistent connections
- **CDN Integration**: Static asset delivery optimization
- **Regional Deployment**: Geo-distributed deployments

## Deployment Patterns

### Development Environment

```yaml
# docker-compose.dev.yml
version: '3.8'
services:
  orchestrator:
    build: ./core/orchestrator
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://agentflow:password@postgres:5432/agentflow
      - NATS_URL=nats://nats:4222
      - REDIS_URL=redis://redis:6379
    depends_on:
      - postgres
      - nats
      - redis

  runtime:
    build: ./runtime
    environment:
      - NATS_URL=nats://nats:4222
      - ORCHESTRATOR_URL=http://orchestrator:8080
    depends_on:
      - nats

  llm-router:
    build: ./core/llm-router
    environment:
      - REDIS_URL=redis://redis:6379
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    depends_on:
      - redis

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=agentflow
      - POSTGRES_USER=agentflow
      - POSTGRES_PASSWORD=password
    volumes:
      - postgres_data:/var/lib/postgresql/data

  nats:
    image: nats:2.10
    command: 
      - "--jetstream"
      - "--store_dir=/data"
    volumes:
      - nats_data:/data

  redis:
    image: redis:7
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  nats_data:
  redis_data:
```

### Production Kubernetes

```yaml
# Orchestrator Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: orchestrator
  namespace: agentflow
spec:
  replicas: 3
  selector:
    matchLabels:
      app: orchestrator
  template:
    metadata:
      labels:
        app: orchestrator
    spec:
      containers:
      - name: orchestrator
        image: agentflow/orchestrator:v1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: postgres-credentials
              key: url
        - name: NATS_URL
          value: "nats://nats:4222"
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "1Gi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Multi-Region Deployment

```yaml
# Global Load Balancer (Google Cloud)
apiVersion: networking.gke.io/v1
kind: ManagedCertificate
metadata:
  name: agentflow-cert
spec:
  domains:
    - agentflow.example.com
    - api.agentflow.example.com

---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: agentflow-global-ingress
  annotations:
    kubernetes.io/ingress.global-static-ip-name: "agentflow-ip"
    networking.gke.io/managed-certificates: "agentflow-cert"
    kubernetes.io/ingress.class: "gce"
spec:
  rules:
  - host: api.agentflow.example.com
    http:
      paths:
      - path: /*
        pathType: ImplementationSpecific
        backend:
          service:
            name: orchestrator-service
            port:
              number: 80
```

This comprehensive architecture documentation provides the foundation for building a scalable, secure, and maintainable AgentFlow platform that can grow from development to enterprise scale.