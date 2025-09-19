package aor

import (
	"github.com/google/uuid"
	"context"
	"time"

)

// WorkflowSpec defines a DAG workflow
type WorkflowSpec struct {
	ID       uuid.UUID `json:"id" db:"id"`
	OrgID    uuid.UUID `json:"org_id" db:"org_id"`
	Name     string    `json:"name" db:"name"`
	Version  int       `json:"version" db:"version"`
	DAG      DAG       `json:"dag" db:"dag"`
	Metadata Metadata  `json:"metadata" db:"metadata"`
	Created  time.Time `json:"created" db:"created"`
	Updated  time.Time `json:"updated" db:"updated"`
}

// DAG represents a directed acyclic graph
type DAG struct {
	Steps []Step `json:"steps"`
	Edges []Edge `json:"edges"`
}

// Step represents a single step in the workflow
type Step struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	Timeout     time.Duration          `json:"timeout"`
	Retries     int                    `json:"retries"`
	Conditions  []Condition            `json:"conditions"`
}

// Edge represents a dependency between steps
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Condition represents a conditional execution rule
type Condition struct {
	Type   string      `json:"type"`
	Config interface{} `json:"config"`
}

// Metadata contains workflow metadata
type Metadata struct {
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Labels      map[string]string `json:"labels"`
	Author      string            `json:"author"`
}

// WorkflowRun represents an execution instance
type WorkflowRun struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	WorkflowID     uuid.UUID              `json:"workflow_id" db:"workflow_id"`
	WorkflowSpecID uuid.UUID              `json:"workflow_spec_id" db:"workflow_spec_id"`
	OrgID          uuid.UUID              `json:"org_id" db:"org_id"`
	Status         WorkflowStatus         `json:"status" db:"status"`
	Input          map[string]interface{} `json:"input" db:"input"`
	Output         map[string]interface{} `json:"output" db:"output"`
	Error          string                 `json:"error,omitempty" db:"error"`
	StartedAt      time.Time              `json:"started_at" db:"started_at"`
	FinishedAt     *time.Time             `json:"finished_at,omitempty" db:"finished_at"`
	EndedAt        *time.Time             `json:"ended_at,omitempty" db:"ended_at"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	CostCents      int64                  `json:"cost_cents" db:"cost_cents"`
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
	Steps          []StepRun              `json:"steps" db:"steps"`
}

// StepRun represents a step execution
type StepRun struct {
	ID            string                 `json:"id"`
	StepID        string                 `json:"step_id"`
	WorkflowRunID uuid.UUID              `json:"workflow_run_id"`
	NodeID        string                 `json:"node_id"`
	Status        StepStatus             `json:"status"`
	Input         map[string]interface{} `json:"input"`
	Output        map[string]interface{} `json:"output"`
	Error         string                 `json:"error,omitempty"`
	StartedAt     time.Time              `json:"started_at"`
	FinishedAt    *time.Time             `json:"finished_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	Attempt       int                    `json:"attempt"`
	Attempts      int                    `json:"attempts"`
}

// Task represents a unit of work
type Task struct {
	ID         uuid.UUID              `json:"id"`
	RunID      uuid.UUID              `json:"run_id"`
	StepID     string                 `json:"step_id"`
	NodeID     string                 `json:"node_id"`
	Type       string                 `json:"type"`
	Input      map[string]interface{} `json:"input"`
	Inputs     map[string]interface{} `json:"inputs"`
	Config     map[string]interface{} `json:"config"`
	Node       *Node                  `json:"node,omitempty"`
	Attempt    int                    `json:"attempt"`
	Priority   int                    `json:"priority"`
	CreatedAt  time.Time              `json:"created_at"`
	ScheduledAt time.Time             `json:"scheduled_at"`
	DeadlineAt *time.Time             `json:"deadline_at,omitempty"`
}

// TaskResult represents the result of task execution
type TaskResult struct {
	TaskID           uuid.UUID              `json:"task_id"`
	Status           TaskStatus             `json:"status"`
	Output           map[string]interface{} `json:"output"`
	Error            string                 `json:"error,omitempty"`
	ExecutedAt       time.Time              `json:"executed_at"`
	Duration         time.Duration          `json:"duration"`
	CostCents        int64                  `json:"cost_cents"`
	TokensPrompt     int                    `json:"tokens_prompt"`
	TokensCompletion int                    `json:"tokens_completion"`
}

// Executor interface for different step types
type Executor interface {
	Execute(ctx context.Context, task *Task) (*TaskResult, error)
	CanHandle(stepType string) bool
}

// WorkflowStatus represents the status of a workflow run
type WorkflowStatus string

const (
	WorkflowStatusPending   WorkflowStatus = "pending"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
)

// Legacy aliases for compatibility
const (
	RunStatusQueued    = WorkflowStatusPending
	RunStatusRunning   = WorkflowStatusRunning
	RunStatusCompleted = WorkflowStatusCompleted
	RunStatusFailed    = WorkflowStatusFailed
	RunStatusCancelled = WorkflowStatusCancelled
)

type RunStatus = WorkflowStatus

// StepStatus represents the status of a step
type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusQueued    StepStatus = "queued"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusSucceeded StepStatus = "succeeded"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusSucceeded TaskStatus = "succeeded"
	TaskStatusFailed    TaskStatus = "failed"
)

// ExecutorType represents different executor types
type ExecutorType string

const (
	ExecutorTypeLLM      ExecutorType = "llm"
	ExecutorTypeHTTP     ExecutorType = "http"
	ExecutorTypeScript   ExecutorType = "script"
	ExecutorTypeWASM     ExecutorType = "wasm"
	ExecutorTypeWorkflow ExecutorType = "workflow"
)

// RunRequest represents a workflow execution request
type RunRequest struct {
	WorkflowName    string                 `json:"workflow_name"`
	WorkflowVersion int                    `json:"workflow_version"`
	Inputs          map[string]interface{} `json:"inputs"`
	Tags            []string               `json:"tags"`
	BudgetCents     int64                  `json:"budget_cents"`
	Priority        int                    `json:"priority"`
}

// Node represents a workflow node (for scheduler compatibility)
type Node struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Config   map[string]interface{} `json:"config"`
	Status   string                 `json:"status"`
	Children []string               `json:"children"`
}