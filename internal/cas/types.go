package cas

import (
	"time"

	"github.com/google/uuid"
)

// Budget represents spending limits and controls
type Budget struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	OrgID       uuid.UUID   `json:"org_id" db:"org_id"`
	ProjectID   *uuid.UUID  `json:"project_id" db:"project_id"`
	PeriodType  PeriodType  `json:"period_type" db:"period_type"`
	LimitCents  int64       `json:"limit_cents" db:"limit_cents"`
	SpentCents  int64       `json:"spent_cents" db:"spent_cents"`
	PeriodStart time.Time   `json:"period_start" db:"period_start"`
	PeriodEnd   time.Time   `json:"period_end" db:"period_end"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
}

type PeriodType string

const (
	PeriodDaily   PeriodType = "daily"
	PeriodWeekly  PeriodType = "weekly"
	PeriodMonthly PeriodType = "monthly"
)

// ProviderConfig represents configuration for a model provider
type ProviderConfig struct {
	ID                    uuid.UUID              `json:"id" db:"id"`
	OrgID                 uuid.UUID              `json:"org_id" db:"org_id"`
	ProviderName          string                 `json:"provider_name" db:"provider_name"`
	ModelName             string                 `json:"model_name" db:"model_name"`
	Config                map[string]interface{} `json:"config" db:"config"`
	CostPerTokenPrompt    float64                `json:"cost_per_token_prompt" db:"cost_per_token_prompt"`
	CostPerTokenCompletion float64               `json:"cost_per_token_completion" db:"cost_per_token_completion"`
	QPSLimit              int                    `json:"qps_limit" db:"qps_limit"`
	Enabled               bool                   `json:"enabled" db:"enabled"`
	CreatedAt             time.Time              `json:"created_at" db:"created_at"`
}

// RoutingRequest represents a request for provider/model selection
type RoutingRequest struct {
	OrgID         uuid.UUID              `json:"org_id"`
	QualityTier   QualityTier            `json:"quality_tier"`
	PromptTokens  int                    `json:"prompt_tokens"`
	MaxTokens     int                    `json:"max_tokens"`
	LatencySLA    time.Duration          `json:"latency_sla,omitempty"`
	BudgetCents   int64                  `json:"budget_cents,omitempty"`
	Constraints   map[string]interface{} `json:"constraints,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
}

type QualityTier string

const (
	QualityGold   QualityTier = "Gold"
	QualitySilver QualityTier = "Silver"
	QualityBronze QualityTier = "Bronze"
)

// RoutingResponse represents the selected provider and configuration
type RoutingResponse struct {
	ProviderName     string                 `json:"provider_name"`
	ModelName        string                 `json:"model_name"`
	Config           map[string]interface{} `json:"config"`
	EstimatedCost    int64                  `json:"estimated_cost_cents"`
	EstimatedLatency time.Duration          `json:"estimated_latency"`
	Confidence       float64                `json:"confidence"`
	Reason           string                 `json:"reason"`
	Alternatives     []Alternative          `json:"alternatives,omitempty"`
}

type Alternative struct {
	ProviderName     string        `json:"provider_name"`
	ModelName        string        `json:"model_name"`
	EstimatedCost    int64         `json:"estimated_cost_cents"`
	EstimatedLatency time.Duration `json:"estimated_latency"`
	QualityScore     float64       `json:"quality_score"`
	Reason           string        `json:"reason"`
}

// CacheRequest represents a request to cache a completion
type CacheRequest struct {
	Key        string                 `json:"key"`
	PromptHash string                 `json:"prompt_hash"`
	InputHash  string                 `json:"input_hash"`
	Response   map[string]interface{} `json:"response"`
	TTL        time.Duration          `json:"ttl"`
	Policy     CachePolicy            `json:"policy"`
}

type CachePolicy struct {
	Enabled       bool                   `json:"enabled"`
	TTL           time.Duration          `json:"ttl"`
	PrivacyLevel  PrivacyLevel           `json:"privacy_level"`
	Conditions    map[string]interface{} `json:"conditions,omitempty"`
}

type PrivacyLevel string

const (
	PrivacyPublic  PrivacyLevel = "public"
	PrivacyOrg     PrivacyLevel = "org"
	PrivacyProject PrivacyLevel = "project"
	PrivacyUser    PrivacyLevel = "user"
)

// CacheResponse represents a cached completion response
type CacheResponse struct {
	Hit       bool                   `json:"hit"`
	Response  map[string]interface{} `json:"response,omitempty"`
	CreatedAt time.Time              `json:"created_at,omitempty"`
	ExpiresAt time.Time              `json:"expires_at,omitempty"`
}

// QuotaStatus represents current quota usage for a provider
type QuotaStatus struct {
	ProviderName    string    `json:"provider_name"`
	ModelName       string    `json:"model_name"`
	CurrentQPS      int       `json:"current_qps"`
	LimitQPS        int       `json:"limit_qps"`
	ConcurrentCalls int       `json:"concurrent_calls"`
	MaxConcurrent   int       `json:"max_concurrent"`
	LastReset       time.Time `json:"last_reset"`
	NextReset       time.Time `json:"next_reset"`
}

// BudgetStatus represents current budget usage
type BudgetStatus struct {
	BudgetID       uuid.UUID `json:"budget_id"`
	LimitCents     int64     `json:"limit_cents"`
	SpentCents     int64     `json:"spent_cents"`
	RemainingCents int64     `json:"remaining_cents"`
	UtilizationPct float64   `json:"utilization_pct"`
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
	Status         BudgetStatusType `json:"status"`
}

type BudgetStatusType string

const (
	BudgetStatusHealthy   BudgetStatusType = "healthy"
	BudgetStatusWarning   BudgetStatusType = "warning"
	BudgetStatusCritical  BudgetStatusType = "critical"
	BudgetStatusExceeded  BudgetStatusType = "exceeded"
)

// OptimizationSuggestion represents a cost optimization recommendation
type OptimizationSuggestion struct {
	Type           OptimizationType       `json:"type"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	PotentialSaving int64                 `json:"potential_saving_cents"`
	Confidence     float64                `json:"confidence"`
	Impact         ImpactLevel            `json:"impact"`
	Actions        []string               `json:"actions"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type OptimizationType string

const (
	OptimizationProviderSwitch OptimizationType = "provider_switch"
	OptimizationModelDowngrade OptimizationType = "model_downgrade"
	OptimizationCaching        OptimizationType = "caching"
	OptimizationBatching       OptimizationType = "batching"
	OptimizationScheduling     OptimizationType = "scheduling"
)

type ImpactLevel string

const (
	ImpactLow    ImpactLevel = "low"
	ImpactMedium ImpactLevel = "medium"
	ImpactHigh   ImpactLevel = "high"
)

// ProviderMetrics represents performance metrics for a provider
type ProviderMetrics struct {
	ProviderName     string        `json:"provider_name"`
	ModelName        string        `json:"model_name"`
	AvgLatency       time.Duration `json:"avg_latency"`
	P95Latency       time.Duration `json:"p95_latency"`
	SuccessRate      float64       `json:"success_rate"`
	AvgCostPerToken  float64       `json:"avg_cost_per_token"`
	QualityScore     float64       `json:"quality_score"`
	ReliabilityScore float64       `json:"reliability_score"`
	LastUpdated      time.Time     `json:"last_updated"`
}

// BatchRequest represents a request to batch multiple operations
type BatchRequest struct {
	Operations []BatchOperation       `json:"operations"`
	Policy     BatchPolicy            `json:"policy"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type BatchOperation struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

type BatchPolicy struct {
	MaxBatchSize   int           `json:"max_batch_size"`
	MaxWaitTime    time.Duration `json:"max_wait_time"`
	CompatibleOnly bool          `json:"compatible_only"`
}

// BatchResponse represents the result of a batch operation
type BatchResponse struct {
	BatchID   string          `json:"batch_id"`
	Results   []BatchResult   `json:"results"`
	Summary   BatchSummary    `json:"summary"`
	CreatedAt time.Time       `json:"created_at"`
}

type BatchResult struct {
	OperationID string                 `json:"operation_id"`
	Status      string                 `json:"status"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

type BatchSummary struct {
	TotalOperations int   `json:"total_operations"`
	SuccessCount    int   `json:"success_count"`
	FailureCount    int   `json:"failure_count"`
	TotalCostCents  int64 `json:"total_cost_cents"`
	TotalSavings    int64 `json:"total_savings_cents"`
}