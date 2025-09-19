package aos

import (
	"time"

	"github.com/google/uuid"
)

// TraceEvent represents a single event in a workflow execution trace
type TraceEvent struct {
	OrgID            uuid.UUID              `json:"org_id" ch:"org_id"`
	RunID            uuid.UUID              `json:"run_id" ch:"run_id"`
	StepID           uuid.UUID              `json:"step_id" ch:"step_id"`
	Timestamp        time.Time              `json:"timestamp" ch:"ts"`
	EventType        string                 `json:"event_type" ch:"event_type"`
	Payload          map[string]interface{} `json:"payload" ch:"payload"`
	CostCents        int64                  `json:"cost_cents" ch:"cost_cents"`
	TokensPrompt     int32                  `json:"tokens_prompt" ch:"tokens_prompt"`
	TokensCompletion int32                  `json:"tokens_completion" ch:"tokens_completion"`
	Provider         string                 `json:"provider" ch:"provider"`
	Model            string                 `json:"model" ch:"model"`
	QualityTier      string                 `json:"quality_tier" ch:"quality_tier"`
	LatencyMs        int32                  `json:"latency_ms" ch:"latency_ms"`
}

// EventType constants
const (
	EventTypeStarted     = "started"
	EventTypeCompleted   = "completed"
	EventTypeRetry       = "retry"
	EventTypeLog         = "log"
	EventTypeToolCall    = "tool_call"
	EventTypeModelIO     = "model_io"
	EventTypeError       = "error"
	EventTypeCanceled    = "canceled"
	EventTypeHeartbeat   = "heartbeat"
)

// TraceQuery represents a query for trace data
type TraceQuery struct {
	OrgID     uuid.UUID  `json:"org_id"`
	RunID     *uuid.UUID `json:"run_id,omitempty"`
	StepID    *uuid.UUID `json:"step_id,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	EventType *string    `json:"event_type,omitempty"`
	Provider  *string    `json:"provider,omitempty"`
	Model     *string    `json:"model,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

// TraceResponse represents the response to a trace query
type TraceResponse struct {
	Events     []TraceEvent `json:"events"`
	TotalCount int64        `json:"total_count"`
	Summary    TraceSummary `json:"summary"`
}

// TraceSummary provides aggregate information about a trace
type TraceSummary struct {
	TotalEvents      int64         `json:"total_events"`
	TotalCost        int64         `json:"total_cost_cents"`
	TotalTokens      int64         `json:"total_tokens"`
	AverageLatency   time.Duration `json:"average_latency"`
	SuccessRate      float64       `json:"success_rate"`
	ErrorCount       int64         `json:"error_count"`
	ProviderBreakdown map[string]int64 `json:"provider_breakdown"`
	ModelBreakdown   map[string]int64 `json:"model_breakdown"`
}

// ReplayRequest represents a request to replay a workflow run
type ReplayRequest struct {
	RunID       uuid.UUID              `json:"run_id"`
	Mode        ReplayMode             `json:"mode"`
	Overrides   map[string]interface{} `json:"overrides,omitempty"`
	StepFilter  []string               `json:"step_filter,omitempty"`
	DryRun      bool                   `json:"dry_run,omitempty"`
}

type ReplayMode string

const (
	ReplayModeShadow ReplayMode = "shadow" // Run alongside original without affecting state
	ReplayModeLive   ReplayMode = "live"   // Full replay with state changes
	ReplayModeDebug  ReplayMode = "debug"  // Step-by-step debugging mode
)

// ReplayResponse represents the result of a replay operation
type ReplayResponse struct {
	ReplayRunID uuid.UUID     `json:"replay_run_id"`
	Status      ReplayStatus  `json:"status"`
	Differences []Difference  `json:"differences,omitempty"`
	Summary     ReplaySummary `json:"summary"`
}

type ReplayStatus string

const (
	ReplayStatusQueued    ReplayStatus = "queued"
	ReplayStatusRunning   ReplayStatus = "running"
	ReplayStatusCompleted ReplayStatus = "completed"
	ReplayStatusFailed    ReplayStatus = "failed"
)

// Difference represents a difference between original and replay execution
type Difference struct {
	StepID      string      `json:"step_id"`
	Field       string      `json:"field"`
	Original    interface{} `json:"original"`
	Replay      interface{} `json:"replay"`
	DiffType    DiffType    `json:"diff_type"`
	Significance string     `json:"significance"`
}

type DiffType string

const (
	DiffTypeOutput   DiffType = "output"
	DiffTypeLatency  DiffType = "latency"
	DiffTypeCost     DiffType = "cost"
	DiffTypeTokens   DiffType = "tokens"
	DiffTypeError    DiffType = "error"
	DiffTypeMetadata DiffType = "metadata"
)

// ReplaySummary provides aggregate information about a replay
type ReplaySummary struct {
	TotalSteps       int     `json:"total_steps"`
	MatchingSteps    int     `json:"matching_steps"`
	DifferentSteps   int     `json:"different_steps"`
	FailedSteps      int     `json:"failed_steps"`
	SimilarityScore  float64 `json:"similarity_score"`
	CostDifference   int64   `json:"cost_difference_cents"`
	LatencyDifference time.Duration `json:"latency_difference"`
}

// CostAnalysisRequest represents a request for cost analysis
type CostAnalysisRequest struct {
	OrgID     uuid.UUID  `json:"org_id"`
	StartTime time.Time  `json:"start_time"`
	EndTime   time.Time  `json:"end_time"`
	GroupBy   []string   `json:"group_by"` // workflow, prompt_version, provider, model, quality_tier
	Filters   map[string]interface{} `json:"filters,omitempty"`
}

// CostAnalysisResponse represents cost analysis results
type CostAnalysisResponse struct {
	TotalCost    int64              `json:"total_cost_cents"`
	Breakdown    []CostBreakdown    `json:"breakdown"`
	Trends       []CostTrend        `json:"trends"`
	Projections  []CostProjection   `json:"projections"`
	Savings      CostSavings        `json:"savings"`
}

type CostBreakdown struct {
	Dimensions map[string]string `json:"dimensions"`
	Cost       int64             `json:"cost_cents"`
	Percentage float64           `json:"percentage"`
	Count      int64             `json:"count"`
}

type CostTrend struct {
	Timestamp time.Time `json:"timestamp"`
	Cost      int64     `json:"cost_cents"`
	Count     int64     `json:"count"`
}

type CostProjection struct {
	Period      string  `json:"period"` // daily, weekly, monthly
	Projected   int64   `json:"projected_cost_cents"`
	Confidence  float64 `json:"confidence"`
	Trend       string  `json:"trend"` // increasing, decreasing, stable
}

type CostSavings struct {
	CachingEnabled    int64 `json:"caching_savings_cents"`
	ProviderRouting   int64 `json:"routing_savings_cents"`
	QualityOptimization int64 `json:"quality_savings_cents"`
	TotalSavings      int64 `json:"total_savings_cents"`
}

// QualityDriftRequest represents a request for quality drift analysis
type QualityDriftRequest struct {
	OrgID        uuid.UUID `json:"org_id"`
	PromptName   string    `json:"prompt_name"`
	WindowDays   int       `json:"window_days"`
	BaselineVersion *int   `json:"baseline_version,omitempty"`
}

// QualityDriftResponse represents quality drift analysis results
type QualityDriftResponse struct {
	DriftScore    float64           `json:"drift_score"`
	Trend         string            `json:"trend"`
	Metrics       []QualityMetric   `json:"metrics"`
	Alerts        []QualityAlert    `json:"alerts"`
	Recommendations []string        `json:"recommendations"`
}

type QualityMetric struct {
	Name      string    `json:"name"`
	Current   float64   `json:"current"`
	Baseline  float64   `json:"baseline"`
	Change    float64   `json:"change"`
	Timestamp time.Time `json:"timestamp"`
}

type QualityAlert struct {
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Metric    string    `json:"metric"`
	Threshold float64   `json:"threshold"`
	Actual    float64   `json:"actual"`
	Timestamp time.Time `json:"timestamp"`
}

// MetricsQuery represents a query for aggregated metrics
type MetricsQuery struct {
	OrgID     uuid.UUID              `json:"org_id"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"end_time"`
	Metrics   []string               `json:"metrics"`
	GroupBy   []string               `json:"group_by"`
	Filters   map[string]interface{} `json:"filters,omitempty"`
	Interval  string                 `json:"interval,omitempty"` // 1m, 5m, 1h, 1d
}

// MetricsResponse represents aggregated metrics results
type MetricsResponse struct {
	Series []MetricSeries `json:"series"`
}

type MetricSeries struct {
	Name       string                 `json:"name"`
	Labels     map[string]string      `json:"labels"`
	DataPoints []MetricDataPoint      `json:"data_points"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type MetricDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}