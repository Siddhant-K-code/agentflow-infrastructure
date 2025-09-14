package aos

import (
	"time"

	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/common"
	"github.com/google/uuid"
)

// TraceEvent represents an event in the execution trace
type TraceEvent struct {
	OrgID            uuid.UUID         `json:"org_id"`
	RunID            uuid.UUID         `json:"run_id"`
	StepID           uuid.UUID         `json:"step_id"`
	Timestamp        time.Time         `json:"timestamp"`
	EventType        common.EventType  `json:"event_type"`
	Payload          map[string]interface{} `json:"payload"`
	CostCents        int64             `json:"cost_cents"`
	TokensPrompt     int32             `json:"tokens_prompt"`
	TokensCompletion int32             `json:"tokens_completion"`
	Provider         string            `json:"provider"`
	Model            string            `json:"model"`
	QualityTier      common.QualityTier `json:"quality_tier"`
	LatencyMs        int32             `json:"latency_ms"`
}

// TraceSpan represents a span in the distributed trace
type TraceSpan struct {
	ID       uuid.UUID            `json:"id"`
	ParentID *uuid.UUID           `json:"parent_id,omitempty"`
	RunID    uuid.UUID            `json:"run_id"`
	StepID   uuid.UUID            `json:"step_id"`
	Name     string               `json:"name"`
	StartTime time.Time           `json:"start_time"`
	EndTime   *time.Time          `json:"end_time,omitempty"`
	Duration  *time.Duration      `json:"duration,omitempty"`
	Status    common.Status       `json:"status"`
	Tags      map[string]string   `json:"tags"`
	Events    []TraceEvent        `json:"events"`
}

// RunTrace represents the complete trace for a workflow run
type RunTrace struct {
	RunID     uuid.UUID           `json:"run_id"`
	StartTime time.Time           `json:"start_time"`
	EndTime   *time.Time          `json:"end_time,omitempty"`
	Duration  *time.Duration      `json:"duration,omitempty"`
	Status    common.Status       `json:"status"`
	Spans     []TraceSpan         `json:"spans"`
	Timeline  []TraceEvent        `json:"timeline"`
	Stats     RunStats            `json:"stats"`
}

// RunStats contains aggregate statistics for a run
type RunStats struct {
	TotalSteps        int           `json:"total_steps"`
	SuccessfulSteps   int           `json:"successful_steps"`
	FailedSteps       int           `json:"failed_steps"`
	TotalCostCents    int64         `json:"total_cost_cents"`
	TotalTokens       int64         `json:"total_tokens"`
	AverageLatencyMs  float64       `json:"average_latency_ms"`
	TotalDuration     time.Duration `json:"total_duration"`
	RetryCount        int           `json:"retry_count"`
}

// SemanticDiff represents the difference between two runs
type SemanticDiff struct {
	BaseRunID    uuid.UUID      `json:"base_run_id"`
	CompareRunID uuid.UUID      `json:"compare_run_id"`
	Summary      DiffSummary    `json:"summary"`
	StepDiffs    []StepDiff     `json:"step_diffs"`
	TokenDiffs   []TokenDiff    `json:"token_diffs,omitempty"`
}

// DiffSummary provides high-level diff statistics
type DiffSummary struct {
	ChangedSteps   int     `json:"changed_steps"`
	AddedSteps     int     `json:"added_steps"`
	RemovedSteps   int     `json:"removed_steps"`
	CostDelta      int64   `json:"cost_delta_cents"`
	DurationDelta  int64   `json:"duration_delta_ms"`
	SimilarityScore float64 `json:"similarity_score"`
}

// StepDiff represents differences at the step level
type StepDiff struct {
	StepID      string        `json:"step_id"`
	DiffType    DiffType      `json:"diff_type"`
	Changes     []FieldChange `json:"changes"`
	BaseOutput  string        `json:"base_output,omitempty"`
	CompareOutput string      `json:"compare_output,omitempty"`
}

// TokenDiff represents token-level differences in outputs
type TokenDiff struct {
	Position int      `json:"position"`
	Type     DiffType `json:"type"`
	BaseToken string  `json:"base_token,omitempty"`
	CompareToken string `json:"compare_token,omitempty"`
}

// FieldChange represents a change in a specific field
type FieldChange struct {
	Field    string      `json:"field"`
	BaseValue interface{} `json:"base_value"`
	CompareValue interface{} `json:"compare_value"`
	ChangeType string     `json:"change_type"` // modified, added, removed
}

// DiffType represents the type of difference
type DiffType string

const (
	DiffTypeAdded    DiffType = "added"
	DiffTypeRemoved  DiffType = "removed"
	DiffTypeModified DiffType = "modified"
	DiffTypeEqual    DiffType = "equal"
)

// ReplayRequest represents a request to replay a workflow run
type ReplayRequest struct {
	RunID      uuid.UUID              `json:"run_id"`
	Mode       ReplayMode             `json:"mode"`
	StepFilter []string               `json:"step_filter,omitempty"` // specific steps to replay
	Options    map[string]interface{} `json:"options,omitempty"`
}

// ReplayMode represents different replay modes
type ReplayMode string

const (
	ReplayModeLive   ReplayMode = "live"   // actually execute against live services
	ReplayModeShadow ReplayMode = "shadow" // run in parallel but don't affect production
	ReplayModeDry    ReplayMode = "dry"    // just validate without execution
)

// ReplayResult represents the result of a replay operation
type ReplayResult struct {
	ReplayID    uuid.UUID     `json:"replay_id"`
	OriginalRunID uuid.UUID   `json:"original_run_id"`
	NewRunID    *uuid.UUID    `json:"new_run_id,omitempty"`
	Mode        ReplayMode    `json:"mode"`
	Status      common.Status `json:"status"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     *time.Time    `json:"end_time,omitempty"`
	Diff        *SemanticDiff `json:"diff,omitempty"`
	Issues      []ReplayIssue `json:"issues,omitempty"`
}

// ReplayIssue represents an issue encountered during replay
type ReplayIssue struct {
	StepID   string `json:"step_id"`
	Type     string `json:"type"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // error, warning, info
}

// CostAnalysis provides cost breakdown and insights
type CostAnalysis struct {
	Period      string                 `json:"period"`
	TotalCost   int64                  `json:"total_cost_cents"`
	Breakdown   CostBreakdown          `json:"breakdown"`
	Trends      []CostTrend            `json:"trends"`
	Projections []CostProjection       `json:"projections"`
	Anomalies   []CostAnomaly          `json:"anomalies,omitempty"`
}

// CostBreakdown breaks down costs by various dimensions
type CostBreakdown struct {
	ByWorkflow []WorkflowCost `json:"by_workflow"`
	ByProvider []ProviderCost `json:"by_provider"`
	ByModel    []ModelCost    `json:"by_model"`
	ByQuality  []QualityCost  `json:"by_quality"`
}

// WorkflowCost represents cost for a specific workflow
type WorkflowCost struct {
	WorkflowName string `json:"workflow_name"`
	Version      int    `json:"version"`
	CostCents    int64  `json:"cost_cents"`
	RunCount     int    `json:"run_count"`
	AvgCostCents int64  `json:"avg_cost_cents"`
}

// ProviderCost represents cost for a specific provider
type ProviderCost struct {
	Provider  string `json:"provider"`
	CostCents int64  `json:"cost_cents"`
	Requests  int64  `json:"requests"`
	Tokens    int64  `json:"tokens"`
}

// ModelCost represents cost for a specific model
type ModelCost struct {
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	CostCents    int64  `json:"cost_cents"`
	Requests     int64  `json:"requests"`
	TokensPrompt int64  `json:"tokens_prompt"`
	TokensCompletion int64 `json:"tokens_completion"`
}

// QualityCost represents cost by quality tier
type QualityCost struct {
	QualityTier common.QualityTier `json:"quality_tier"`
	CostCents   int64              `json:"cost_cents"`
	RunCount    int64              `json:"run_count"`
}

// CostTrend represents cost trends over time
type CostTrend struct {
	Date      time.Time `json:"date"`
	CostCents int64     `json:"cost_cents"`
	RunCount  int       `json:"run_count"`
}

// CostProjection represents future cost projections
type CostProjection struct {
	Date           time.Time `json:"date"`
	ProjectedCents int64     `json:"projected_cents"`
	Confidence     float64   `json:"confidence"`
	Methodology    string    `json:"methodology"`
}

// CostAnomaly represents an unusual cost pattern
type CostAnomaly struct {
	Date        time.Time `json:"date"`
	CostCents   int64     `json:"cost_cents"`
	ExpectedCents int64   `json:"expected_cents"`
	Deviation   float64   `json:"deviation"`
	Reason      string    `json:"reason,omitempty"`
}

// QualityMetrics represents quality drift and regression detection
type QualityMetrics struct {
	Period         string              `json:"period"`
	EvalPassRate   float64             `json:"eval_pass_rate"`
	AvgScore       float64             `json:"avg_score"`
	Regressions    []QualityRegression `json:"regressions,omitempty"`
	DriftAlerts    []DriftAlert        `json:"drift_alerts,omitempty"`
}

// QualityRegression represents a detected quality regression
type QualityRegression struct {
	PromptName     string    `json:"prompt_name"`
	FromVersion    int       `json:"from_version"`
	ToVersion      int       `json:"to_version"`
	ScoreDelta     float64   `json:"score_delta"`
	DetectedAt     time.Time `json:"detected_at"`
	AffectedRuns   int       `json:"affected_runs"`
}

// DriftAlert represents a quality drift alert
type DriftAlert struct {
	Type        string    `json:"type"`
	Metric      string    `json:"metric"`
	Threshold   float64   `json:"threshold"`
	ActualValue float64   `json:"actual_value"`
	DetectedAt  time.Time `json:"detected_at"`
	Severity    string    `json:"severity"`
}

// TraceQuery represents a query for trace data
type TraceQuery struct {
	OrgID      uuid.UUID              `json:"org_id"`
	RunIDs     []uuid.UUID            `json:"run_ids,omitempty"`
	WorkflowName string               `json:"workflow_name,omitempty"`
	Status     common.Status          `json:"status,omitempty"`
	StartTime  *time.Time             `json:"start_time,omitempty"`
	EndTime    *time.Time             `json:"end_time,omitempty"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
	Limit      int                    `json:"limit,omitempty"`
}

// GuardrailsReport represents a security and policy report
type GuardrailsReport struct {
	Period           string                `json:"period"`
	PolicyDenials    int                   `json:"policy_denials"`
	InjectionBlocks  int                   `json:"injection_blocks"`
	PIIRedactions    int                   `json:"pii_redactions"`
	TrustViolations  int                   `json:"trust_violations"`
	TopViolations    []GuardrailViolation  `json:"top_violations"`
}

// GuardrailViolation represents a specific guardrail violation
type GuardrailViolation struct {
	Type        string    `json:"type"`
	Count       int       `json:"count"`
	Description string    `json:"description"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
}