package agentflow

import (
	"time"

	"github.com/google/uuid"
)

// Workflow types

type WorkflowRun struct {
	ID           uuid.UUID              `json:"id"`
	WorkflowName string                 `json:"workflow_name"`
	Version      int                    `json:"version"`
	Status       string                 `json:"status"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	EndedAt      *time.Time             `json:"ended_at,omitempty"`
	CostCents    int64                  `json:"cost_cents"`
	Inputs       map[string]interface{} `json:"inputs"`
	Outputs      map[string]interface{} `json:"outputs,omitempty"`
	Error        string                 `json:"error,omitempty"`
	Tags         map[string]string      `json:"tags,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

type SubmitWorkflowRequest struct {
	WorkflowName    string                 `json:"workflow_name"`
	WorkflowVersion *int                   `json:"workflow_version,omitempty"`
	Inputs          map[string]interface{} `json:"inputs"`
	Tags            map[string]string      `json:"tags,omitempty"`
	BudgetCents     int64                  `json:"budget_cents,omitempty"`
}

type ListWorkflowsOptions struct {
	Status string `json:"status,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Since  string `json:"since,omitempty"`
}

type ListWorkflowsResponse struct {
	Runs       []WorkflowRun `json:"runs"`
	TotalCount int64         `json:"total_count"`
	HasMore    bool          `json:"has_more"`
}

// Prompt types

type PromptTemplate struct {
	ID        uuid.UUID              `json:"id"`
	Name      string                 `json:"name"`
	Version   int                    `json:"version"`
	Template  string                 `json:"template"`
	Schema    map[string]interface{} `json:"schema"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

type CreatePromptRequest struct {
	Name     string                 `json:"name"`
	Template string                 `json:"template"`
	Schema   map[string]interface{} `json:"schema,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type ResolvePromptRequest struct {
	Name     string                 `json:"name"`
	Version  *int                   `json:"version,omitempty"`
	Inputs   map[string]interface{} `json:"inputs"`
	Context  map[string]interface{} `json:"context,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type ResolvePromptResponse struct {
	ID           uuid.UUID              `json:"id"`
	Name         string                 `json:"name"`
	Version      int                    `json:"version"`
	RenderedText string                 `json:"rendered_text"`
	Metadata     map[string]interface{} `json:"metadata"`
	TokenCount   int                    `json:"token_count"`
	IsCanary     bool                   `json:"is_canary"`
}

type PromptDeployment struct {
	ID            uuid.UUID `json:"id"`
	PromptName    string    `json:"prompt_name"`
	StableVersion int       `json:"stable_version"`
	CanaryVersion *int      `json:"canary_version,omitempty"`
	CanaryRatio   float64   `json:"canary_ratio"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type DeployPromptRequest struct {
	PromptName    string  `json:"prompt_name"`
	StableVersion int     `json:"stable_version"`
	CanaryVersion *int    `json:"canary_version,omitempty"`
	CanaryRatio   float64 `json:"canary_ratio,omitempty"`
}

// Trace types

type TraceEvent struct {
	ID               uuid.UUID              `json:"id"`
	RunID            uuid.UUID              `json:"run_id"`
	StepID           uuid.UUID              `json:"step_id"`
	Timestamp        time.Time              `json:"timestamp"`
	EventType        string                 `json:"event_type"`
	Payload          map[string]interface{} `json:"payload"`
	CostCents        int64                  `json:"cost_cents"`
	TokensPrompt     int                    `json:"tokens_prompt"`
	TokensCompletion int                    `json:"tokens_completion"`
	Provider         string                 `json:"provider"`
	Model            string                 `json:"model"`
	LatencyMs        int                    `json:"latency_ms"`
}

type TraceResponse struct {
	Events     []TraceEvent `json:"events"`
	TotalCount int64        `json:"total_count"`
	Summary    TraceSummary `json:"summary"`
}

type TraceSummary struct {
	TotalEvents       int64            `json:"total_events"`
	TotalCost         int64            `json:"total_cost_cents"`
	TotalTokens       int64            `json:"total_tokens"`
	AverageLatency    time.Duration    `json:"average_latency"`
	SuccessRate       float64          `json:"success_rate"`
	ErrorCount        int64            `json:"error_count"`
	ProviderBreakdown map[string]int64 `json:"provider_breakdown"`
	ModelBreakdown    map[string]int64 `json:"model_breakdown"`
}

type TraceQueryRequest struct {
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	EventType *string    `json:"event_type,omitempty"`
	Provider  *string    `json:"provider,omitempty"`
	Model     *string    `json:"model,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

type ReplayRequest struct {
	RunID      uuid.UUID              `json:"run_id"`
	Mode       string                 `json:"mode"`
	Overrides  map[string]interface{} `json:"overrides,omitempty"`
	StepFilter []string               `json:"step_filter,omitempty"`
	DryRun     bool                   `json:"dry_run,omitempty"`
}

type ReplayResponse struct {
	ReplayRunID uuid.UUID     `json:"replay_run_id"`
	Status      string        `json:"status"`
	Differences []Difference  `json:"differences,omitempty"`
	Summary     ReplaySummary `json:"summary"`
}

type Difference struct {
	StepID       string      `json:"step_id"`
	Field        string      `json:"field"`
	Original     interface{} `json:"original"`
	Replay       interface{} `json:"replay"`
	DiffType     string      `json:"diff_type"`
	Significance string      `json:"significance"`
}

type ReplaySummary struct {
	TotalSteps        int           `json:"total_steps"`
	MatchingSteps     int           `json:"matching_steps"`
	DifferentSteps    int           `json:"different_steps"`
	FailedSteps       int           `json:"failed_steps"`
	SimilarityScore   float64       `json:"similarity_score"`
	CostDifference    int64         `json:"cost_difference_cents"`
	LatencyDifference time.Duration `json:"latency_difference"`
}

// Budget types

type Budget struct {
	ID          uuid.UUID `json:"id"`
	PeriodType  string    `json:"period_type"`
	LimitCents  int64     `json:"limit_cents"`
	SpentCents  int64     `json:"spent_cents"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateBudgetRequest struct {
	PeriodType  string     `json:"period_type"`
	LimitCents  int64      `json:"limit_cents"`
	ProjectID   *uuid.UUID `json:"project_id,omitempty"`
	Description string     `json:"description,omitempty"`
}

type BudgetStatus struct {
	BudgetID       uuid.UUID `json:"budget_id"`
	LimitCents     int64     `json:"limit_cents"`
	SpentCents     int64     `json:"spent_cents"`
	RemainingCents int64     `json:"remaining_cents"`
	UtilizationPct float64   `json:"utilization_pct"`
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
	Status         string    `json:"status"`
}
