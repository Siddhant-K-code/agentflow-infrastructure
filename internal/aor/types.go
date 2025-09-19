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
	ID         uuid.UUID            `json:"id" db:"id"`
	WorkflowID uuid.UUID            `json:"workflow_id" db:"workflow_id"`
	OrgID      uuid.UUID            `json:"org_id" db:"org_id"`
	Status     WorkflowStatus       `json:"status" db:"status"`
	Input      map[string]interface{} `json:"input" db:"input"`
	Output     map[string]interface{} `json:"output" db:"output"`
	Error      string               `json:"error,omitempty" db:"error"`
	StartedAt  time.Time            `json:"started_at" db:"started_at"`
	FinishedAt *time.Time           `json:"finished_at,omitempty" db:"finished_at"`
	Steps      []StepRun            `json:"steps" db:"steps"`
}

// StepRun represents a step execution
type StepRun struct {
	ID         string                 `json:"id"`
	StepID     string                 `json:"step_id"`
	Status     StepStatus             `json:"status"`
	Input      map[string]interface{} `json:"input"`
	Output     map[string]interface{} `json:"output"`
	Error      string                 `json:"error,omitempty"`
	StartedAt  time.Time              `json:"started_at"`
	FinishedAt *time.Time             `json:"finished_at,omitempty"`
	Attempts   int                    `json:"attempts"`
}

// Task represents a unit of work
type Task struct {
	ID         uuid.UUID              `json:"id"`
	RunID      uuid.UUID              `json:"run_id"`
	StepID     string                 `json:"step_id"`
	Type       string                 `json:"type"`
	Input      map[string]interface{} `json:"input"`
	Config     map[string]interface{} `json:"config"`
	Priority   int                    `json:"priority"`
	CreatedAt  time.Time              `json:"created_at"`
	ScheduledAt time.Time             `json:"scheduled_at"`
}

// TaskResult represents the result of task execution
type TaskResult struct {
	TaskID     uuid.UUID              `json:"task_id"`
	Status     TaskStatus             `json:"status"`
	Output     map[string]interface{} `json:"output"`
	Error      string                 `json:"error,omitempty"`
	ExecutedAt time.Time              `json:"executed_at"`
	Duration   time.Duration          `json:"duration"`
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

// StepStatus represents the status of a step
type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
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