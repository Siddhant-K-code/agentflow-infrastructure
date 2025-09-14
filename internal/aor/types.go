package aor

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// WorkflowSpec defines a DAG workflow
type WorkflowSpec struct {
	ID       uuid.UUID `json:"id" db:"id"`
	OrgID    uuid.UUID `json:"org_id" db:"org_id"`
	Name     string    `json:"name" db:"name"`
	Version  int       `json:"version" db:"version"`
	DAG      DAG       `json:"dag" db:"dag"`
	Metadata Metadata  `json:"metadata" db:"metadata"`
}

// DAG represents a directed acyclic graph
type DAG struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Node represents a step in the workflow
type Node struct {
	ID     string   `json:"id"`
	Type   NodeType `json:"type"`
	Config NodeConfig `json:"config"`
	Policy Policy   `json:"policy,omitempty"`
}

type NodeType string

const (
	NodeTypeLLM      NodeType = "llm"
	NodeTypeTool     NodeType = "tool"
	NodeTypeFunction NodeType = "function"
	NodeTypeSwitch   NodeType = "switch"
	NodeTypeMap      NodeType = "map"
	NodeTypeReduce   NodeType = "reduce"
)

// NodeConfig holds type-specific configuration
type NodeConfig struct {
	// LLM nodes
	PromptRef string            `json:"prompt_ref,omitempty"`
	Inputs    map[string]string `json:"inputs,omitempty"`
	
	// Tool nodes
	ToolName string                 `json:"tool_name,omitempty"`
	ToolArgs map[string]interface{} `json:"tool_args,omitempty"`
	
	// Function nodes
	FunctionName string                 `json:"function_name,omitempty"`
	FunctionArgs map[string]interface{} `json:"function_args,omitempty"`
	
	// Switch nodes
	SwitchOn   string            `json:"switch_on,omitempty"`
	Cases      map[string]string `json:"cases,omitempty"`
	DefaultCase string           `json:"default_case,omitempty"`
	
	// Map/Reduce nodes
	IterateOver string `json:"iterate_over,omitempty"`
	SubDAG      *DAG   `json:"sub_dag,omitempty"`
}

// Edge represents a dependency between nodes
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Policy defines execution constraints
type Policy struct {
	Quality    QualityTier `json:"quality,omitempty"`
	SLAMillis  int         `json:"sla_ms,omitempty"`
	MaxRetries int         `json:"max_retries,omitempty"`
	Optional   bool        `json:"optional,omitempty"`
}

type QualityTier string

const (
	QualityGold   QualityTier = "Gold"
	QualitySilver QualityTier = "Silver"
	QualityBronze QualityTier = "Bronze"
)

// WorkflowRun represents an execution instance
type WorkflowRun struct {
	ID               uuid.UUID     `json:"id" db:"id"`
	WorkflowSpecID   uuid.UUID     `json:"workflow_spec_id" db:"workflow_spec_id"`
	Status           RunStatus     `json:"status" db:"status"`
	StartedAt        *time.Time    `json:"started_at" db:"started_at"`
	EndedAt          *time.Time    `json:"ended_at" db:"ended_at"`
	CostCents        int64         `json:"cost_cents" db:"cost_cents"`
	Metadata         Metadata      `json:"metadata" db:"metadata"`
	CreatedAt        time.Time     `json:"created_at" db:"created_at"`
}

type RunStatus string

const (
	RunStatusQueued         RunStatus = "queued"
	RunStatusRunning        RunStatus = "running"
	RunStatusSucceeded      RunStatus = "succeeded"
	RunStatusFailed         RunStatus = "failed"
	RunStatusCanceled       RunStatus = "canceled"
	RunStatusPartialSuccess RunStatus = "partial-success"
)

// StepRun represents a node execution
type StepRun struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	WorkflowRunID    uuid.UUID  `json:"workflow_run_id" db:"workflow_run_id"`
	NodeID           string     `json:"node_id" db:"node_id"`
	Attempt          int        `json:"attempt" db:"attempt"`
	Status           StepStatus `json:"status" db:"status"`
	WorkerID         string     `json:"worker_id" db:"worker_id"`
	StartedAt        *time.Time `json:"started_at" db:"started_at"`
	EndedAt          *time.Time `json:"ended_at" db:"ended_at"`
	InputRef         string     `json:"input_ref" db:"input_ref"`
	OutputRef        string     `json:"output_ref" db:"output_ref"`
	Error            string     `json:"error" db:"error"`
	CostCents        int64      `json:"cost_cents" db:"cost_cents"`
	TokensPrompt     int        `json:"tokens_prompt" db:"tokens_prompt"`
	TokensCompletion int        `json:"tokens_completion" db:"tokens_completion"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

type StepStatus string

const (
	StepStatusQueued    StepStatus = "queued"
	StepStatusRunning   StepStatus = "running"
	StepStatusSucceeded StepStatus = "succeeded"
	StepStatusFailed    StepStatus = "failed"
	StepStatusCanceled  StepStatus = "canceled"
)

// Metadata is a flexible JSON field
type Metadata map[string]interface{}

// RunRequest represents a workflow execution request
type RunRequest struct {
	WorkflowName    string                 `json:"workflow_name"`
	WorkflowVersion int                    `json:"workflow_version"`
	Inputs          map[string]interface{} `json:"inputs"`
	Tags            map[string]string      `json:"tags,omitempty"`
	BudgetCents     int64                  `json:"budget_cents,omitempty"`
}

// Task represents work to be done by a worker
type Task struct {
	ID            uuid.UUID              `json:"id"`
	RunID         uuid.UUID              `json:"run_id"`
	NodeID        string                 `json:"node_id"`
	Attempt       int                    `json:"attempt"`
	Node          Node                   `json:"node"`
	Inputs        map[string]interface{} `json:"inputs"`
	Context       context.Context        `json:"-"`
	DeadlineAt    time.Time              `json:"deadline_at"`
}

// TaskResult represents the result of task execution
type TaskResult struct {
	TaskID           uuid.UUID              `json:"task_id"`
	Status           StepStatus             `json:"status"`
	Output           map[string]interface{} `json:"output,omitempty"`
	Error            string                 `json:"error,omitempty"`
	CostCents        int64                  `json:"cost_cents"`
	TokensPrompt     int                    `json:"tokens_prompt"`
	TokensCompletion int                    `json:"tokens_completion"`
	Artifacts        []string               `json:"artifacts,omitempty"`
}