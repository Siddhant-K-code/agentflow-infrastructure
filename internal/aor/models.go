package aor

import (
	"encoding/json"
	"time"

	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/common"
	"github.com/google/uuid"
)

// WorkflowSpec represents a versioned DAG specification
type WorkflowSpec struct {
	ID        uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrgID     uuid.UUID       `json:"org_id" gorm:"type:uuid;not null"`
	Name      string          `json:"name" gorm:"not null"`
	Version   int             `json:"version" gorm:"not null"`
	DAG       json.RawMessage `json:"dag" gorm:"type:jsonb;not null"`
	CreatedAt time.Time       `json:"created_at" gorm:"default:now()"`
}

// WorkflowRun represents an execution instance of a workflow
type WorkflowRun struct {
	ID               uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WorkflowSpecID   uuid.UUID       `json:"workflow_spec_id" gorm:"type:uuid;not null"`
	Status           common.Status   `json:"status" gorm:"not null"`
	StartedAt        *time.Time      `json:"started_at" gorm:"default:now()"`
	EndedAt          *time.Time      `json:"ended_at"`
	CostCents        int64           `json:"cost_cents" gorm:"default:0"`
	Metadata         json.RawMessage `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	BudgetCents      *int64          `json:"budget_cents"`
	Tags             json.RawMessage `json:"tags" gorm:"type:jsonb;default:'[]'"`
	
	// Relations
	WorkflowSpec     *WorkflowSpec   `json:"workflow_spec,omitempty"`
	Steps            []StepRun       `json:"steps,omitempty"`
}

// StepRun represents the execution of a single node in the workflow
type StepRun struct {
	ID               uuid.UUID     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WorkflowRunID    uuid.UUID     `json:"workflow_run_id" gorm:"type:uuid;not null"`
	NodeID           string        `json:"node_id" gorm:"not null"`
	Attempt          int           `json:"attempt" gorm:"default:1"`
	Status           common.Status `json:"status" gorm:"not null"`
	WorkerID         string        `json:"worker_id"`
	StartedAt        *time.Time    `json:"started_at" gorm:"default:now()"`
	EndedAt          *time.Time    `json:"ended_at"`
	InputRef         string        `json:"input_ref"`   // S3 key
	OutputRef        string        `json:"output_ref"`  // S3 key
	Error            string        `json:"error"`
	CostCents        int64         `json:"cost_cents" gorm:"default:0"`
	TokensPrompt     int           `json:"tokens_prompt" gorm:"default:0"`
	TokensCompletion int           `json:"tokens_completion" gorm:"default:0"`
}

// DAGSpec represents the structure of a workflow DAG
type DAGSpec struct {
	Nodes []NodeSpec `json:"nodes"`
	Edges []EdgeSpec `json:"edges,omitempty"`
}

// NodeSpec represents a single node in the workflow
type NodeSpec struct {
	ID       string                 `json:"id"`
	Type     NodeType               `json:"type"`
	Config   map[string]interface{} `json:"config,omitempty"`
	Policy   *PolicySpec            `json:"policy,omitempty"`
	Inputs   map[string]InputRef    `json:"inputs,omitempty"`
	Retries  *RetryConfig           `json:"retries,omitempty"`
}

// EdgeSpec represents a dependency between nodes
type EdgeSpec struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// NodeType represents the type of workflow node
type NodeType string

const (
	NodeTypeTool     NodeType = "tool"
	NodeTypeFunction NodeType = "function"
	NodeTypeLLM      NodeType = "llm"
	NodeTypeSwitch   NodeType = "switch"
	NodeTypeMap      NodeType = "map"
	NodeTypeReduce   NodeType = "reduce"
	NodeTypeSubDAG   NodeType = "subdag"
)

// InputRef represents a reference to input data
type InputRef struct {
	From  string `json:"from,omitempty"`  // e.g., "step_id.output"
	Value interface{} `json:"value,omitempty"` // literal value
}

// PolicySpec represents execution policies for a step
type PolicySpec struct {
	Quality   common.QualityTier `json:"quality,omitempty"`
	SLAMillis int                `json:"sla_ms,omitempty"`
	Optional  bool               `json:"optional,omitempty"`
	Timeout   time.Duration      `json:"timeout,omitempty"`
}

// RetryConfig specifies retry behavior
type RetryConfig struct {
	MaxAttempts int           `json:"max_attempts" default:"3"`
	Backoff     time.Duration `json:"backoff" default:"1s"`
	MaxBackoff  time.Duration `json:"max_backoff" default:"30s"`
}

// RunRequest represents a request to start a workflow run
type RunRequest struct {
	WorkflowName    string                 `json:"workflow_name"`
	WorkflowVersion *int                   `json:"workflow_version,omitempty"` // defaults to latest
	Inputs          map[string]interface{} `json:"inputs,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	BudgetCents     *int64                 `json:"budget_cents,omitempty"`
	Priority        int                    `json:"priority,omitempty"`
}

// RunResponse represents the response from starting a workflow run
type RunResponse struct {
	ID     uuid.UUID     `json:"id"`
	Status common.Status `json:"status"`
}

// CancelRequest represents a request to cancel a workflow run
type CancelRequest struct {
	RunID  uuid.UUID `json:"run_id"`
	Reason string    `json:"reason,omitempty"`
}

// SignalRequest represents an external signal to a running workflow
type SignalRequest struct {
	RunID   uuid.UUID              `json:"run_id"`
	Signal  string                 `json:"signal"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}