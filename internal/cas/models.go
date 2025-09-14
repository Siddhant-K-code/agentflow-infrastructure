package cas

import (
	"time"

	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/common"
	"github.com/google/uuid"
)

// BudgetConfig represents budget configuration for an organization
type BudgetConfig struct {
	ID                   uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrgID                uuid.UUID              `json:"org_id" gorm:"type:uuid;not null"`
	Period               common.BudgetPeriod    `json:"period" gorm:"not null"`
	LimitCents           int64                  `json:"limit_cents" gorm:"not null"`
	AlertThresholdRatio  float64                `json:"alert_threshold_ratio" gorm:"default:0.8"`
	CreatedAt            time.Time              `json:"created_at" gorm:"default:now()"`
}

// ProviderConfig represents configuration for a model provider
type ProviderConfig struct {
	ID                          uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrgID                       uuid.UUID              `json:"org_id" gorm:"type:uuid;not null"`
	ProviderName                string                 `json:"provider_name" gorm:"not null"`
	ModelName                   string                 `json:"model_name" gorm:"not null"`
	PricePerPromptTokenCents    int                    `json:"price_per_prompt_token_cents" gorm:"not null"`
	PricePerCompletionTokenCents int                   `json:"price_per_completion_token_cents" gorm:"not null"`
	QualityTier                 common.QualityTier     `json:"quality_tier" gorm:"default:'Silver'"`
	MaxQPS                      int                    `json:"max_qps" gorm:"default:10"`
	Enabled                     bool                   `json:"enabled" gorm:"default:true"`
	CreatedAt                   time.Time              `json:"created_at" gorm:"default:now()"`
}

// RoutingRequest represents a request for model routing
type RoutingRequest struct {
	PromptVersion    string                 `json:"prompt_version,omitempty"`
	EstimatedTokens  int                    `json:"estimated_tokens"`
	QualityTier      common.QualityTier     `json:"quality_tier"`
	SLAMillis        int                    `json:"sla_ms,omitempty"`
	BudgetRemaining  int64                  `json:"budget_remaining_cents"`
	Context          map[string]interface{} `json:"context,omitempty"`
}

// RoutingResponse represents the routing decision
type RoutingResponse struct {
	Provider         string                 `json:"provider"`
	Model            string                 `json:"model"`
	EstimatedCost    int64                  `json:"estimated_cost_cents"`
	ExpectedLatency  int                    `json:"expected_latency_ms"`
	QualityScore     float64                `json:"quality_score"`
	Reasoning        string                 `json:"reasoning"`
	Alternatives     []ProviderAlternative  `json:"alternatives,omitempty"`
}

// ProviderAlternative represents an alternative provider option
type ProviderAlternative struct {
	Provider        string  `json:"provider"`
	Model           string  `json:"model"`
	EstimatedCost   int64   `json:"estimated_cost_cents"`
	QualityScore    float64 `json:"quality_score"`
	Reason          string  `json:"reason"`
}

// CacheKey represents a cache key for request/response caching
type CacheKey struct {
	PromptHash   string                 `json:"prompt_hash"`
	InputHash    string                 `json:"input_hash"`
	ModelConfig  map[string]interface{} `json:"model_config"`
	OrgID        uuid.UUID              `json:"org_id"`
}

// CacheEntry represents a cached response
type CacheEntry struct {
	Key         CacheKey               `json:"key"`
	Response    map[string]interface{} `json:"response"`
	CostCents   int64                  `json:"cost_cents"`
	TokensUsed  int                    `json:"tokens_used"`
	CreatedAt   time.Time              `json:"created_at"`
	ExpiresAt   time.Time              `json:"expires_at"`
	HitCount    int                    `json:"hit_count"`
	Privacy     PrivacyLevel           `json:"privacy"`
}

// PrivacyLevel represents the privacy level for caching
type PrivacyLevel string

const (
	PrivacyPublic       PrivacyLevel = "public"
	PrivacyOrganization PrivacyLevel = "organization"
	PrivacyPrivate      PrivacyLevel = "private"
)

// BudgetUsage represents current budget usage
type BudgetUsage struct {
	OrgID           uuid.UUID              `json:"org_id"`
	Period          common.BudgetPeriod    `json:"period"`
	UsedCents       int64                  `json:"used_cents"`
	LimitCents      int64                  `json:"limit_cents"`
	RemainingCents  int64                  `json:"remaining_cents"`
	UtilizationRate float64                `json:"utilization_rate"`
	ProjectedUsage  int64                  `json:"projected_usage_cents"`
	AlertLevel      AlertLevel             `json:"alert_level"`
}

// AlertLevel represents budget alert levels
type AlertLevel string

const (
	AlertLevelNone     AlertLevel = "none"
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelCritical AlertLevel = "critical"
)

// QuotaConfig represents rate limiting configuration
type QuotaConfig struct {
	ProviderName    string    `json:"provider_name"`
	MaxQPS          int       `json:"max_qps"`
	MaxConcurrent   int       `json:"max_concurrent"`
	BurstLimit      int       `json:"burst_limit"`
	WindowSize      time.Duration `json:"window_size"`
	BackoffStrategy BackoffStrategy `json:"backoff_strategy"`
}

// BackoffStrategy represents different backoff strategies
type BackoffStrategy string

const (
	BackoffExponential BackoffStrategy = "exponential"
	BackoffLinear      BackoffStrategy = "linear"
	BackoffFixed       BackoffStrategy = "fixed"
)

// ProviderStatus represents the current status of a provider
type ProviderStatus struct {
	ProviderName    string    `json:"provider_name"`
	Available       bool      `json:"available"`
	CurrentQPS      int       `json:"current_qps"`
	ErrorRate       float64   `json:"error_rate"`
	AvgLatencyMs    int       `json:"avg_latency_ms"`
	LastHealthCheck time.Time `json:"last_health_check"`
	LastError       string    `json:"last_error,omitempty"`
}

// BatchRequest represents a request for batching multiple calls
type BatchRequest struct {
	Requests    []IndividualRequest    `json:"requests"`
	BatchType   BatchType              `json:"batch_type"`
	MaxBatchSize int                   `json:"max_batch_size"`
	MaxWaitTime time.Duration          `json:"max_wait_time"`
}

// IndividualRequest represents a single request that can be batched
type IndividualRequest struct {
	ID       string                 `json:"id"`
	Type     RequestType            `json:"type"`
	Payload  map[string]interface{} `json:"payload"`
	Priority int                    `json:"priority"`
}

// BatchType represents different types of batching
type BatchType string

const (
	BatchTypeEmbedding   BatchType = "embedding"
	BatchTypeRetrieval   BatchType = "retrieval"
	BatchTypeCompletion  BatchType = "completion"
)

// RequestType represents the type of individual request
type RequestType string

const (
	RequestTypeLLM       RequestType = "llm"
	RequestTypeEmbedding RequestType = "embedding"
	RequestTypeTool      RequestType = "tool"
)

// BatchResponse represents the response from a batch operation
type BatchResponse struct {
	BatchID     string                 `json:"batch_id"`
	Responses   []IndividualResponse   `json:"responses"`
	TotalCost   int64                  `json:"total_cost_cents"`
	BatchedCount int                   `json:"batched_count"`
	Savings     int64                  `json:"savings_cents"`
	Duration    time.Duration          `json:"duration"`
}

// IndividualResponse represents a response to an individual request
type IndividualResponse struct {
	RequestID string                 `json:"request_id"`
	Success   bool                   `json:"success"`
	Response  map[string]interface{} `json:"response,omitempty"`
	Error     string                 `json:"error,omitempty"`
	CostCents int64                  `json:"cost_cents"`
}

// DegradationPolicy represents graceful degradation policies
type DegradationPolicy struct {
	TriggerThreshold float64            `json:"trigger_threshold"` // budget utilization rate
	Actions          []DegradationAction `json:"actions"`
	Priority         int                `json:"priority"`
}

// DegradationAction represents an action to take during degradation
type DegradationAction struct {
	Type        DegradationType        `json:"type"`
	Target      string                 `json:"target"`
	Parameters  map[string]interface{} `json:"parameters"`
	Description string                 `json:"description"`
}

// DegradationType represents different types of degradation
type DegradationType string

const (
	DegradationReduceTemperature DegradationType = "reduce_temperature"
	DegradationShorterContext    DegradationType = "shorter_context"
	DegradationSkipOptional      DegradationType = "skip_optional"
	DegradationCheaperModel      DegradationType = "cheaper_model"
	DegradationThrottleRequests  DegradationType = "throttle_requests"
)

// ProviderMetrics represents performance metrics for a provider
type ProviderMetrics struct {
	ProviderName     string        `json:"provider_name"`
	Model            string        `json:"model"`
	RequestCount     int64         `json:"request_count"`
	SuccessCount     int64         `json:"success_count"`
	ErrorCount       int64         `json:"error_count"`
	TotalCostCents   int64         `json:"total_cost_cents"`
	TotalTokens      int64         `json:"total_tokens"`
	AvgLatencyMs     float64       `json:"avg_latency_ms"`
	P95LatencyMs     float64       `json:"p95_latency_ms"`
	P99LatencyMs     float64       `json:"p99_latency_ms"`
	QualityScore     float64       `json:"quality_score"`
	LastUpdated      time.Time     `json:"last_updated"`
}

// BanditArm represents an arm in the multi-armed bandit for provider selection
type BanditArm struct {
	Provider     string    `json:"provider"`
	Model        string    `json:"model"`
	QualityTier  common.QualityTier `json:"quality_tier"`
	Pulls        int64     `json:"pulls"`
	Rewards      float64   `json:"rewards"`
	Confidence   float64   `json:"confidence"`
	LastSelected time.Time `json:"last_selected"`
}

// SelectionStrategy represents different provider selection strategies
type SelectionStrategy string

const (
	StrategyLowestCost     SelectionStrategy = "lowest_cost"
	StrategyBestQuality    SelectionStrategy = "best_quality"
	StrategyBalanced       SelectionStrategy = "balanced"
	StrategyBandit         SelectionStrategy = "bandit"
	StrategyRoundRobin     SelectionStrategy = "round_robin"
)

// CostOptimizationReport represents recommendations for cost optimization
type CostOptimizationReport struct {
	OrgID           uuid.UUID                  `json:"org_id"`
	Period          string                     `json:"period"`
	CurrentCost     int64                      `json:"current_cost_cents"`
	PotentialSavings int64                     `json:"potential_savings_cents"`
	Recommendations []OptimizationRecommendation `json:"recommendations"`
	GeneratedAt     time.Time                  `json:"generated_at"`
}

// OptimizationRecommendation represents a specific cost optimization recommendation
type OptimizationRecommendation struct {
	Type            RecommendationType `json:"type"`
	Description     string             `json:"description"`
	Impact          string             `json:"impact"`
	SavingsCents    int64              `json:"savings_cents"`
	ImplementationEffort string        `json:"implementation_effort"`
	RiskLevel       string             `json:"risk_level"`
	Details         map[string]interface{} `json:"details"`
}

// RecommendationType represents different types of optimization recommendations
type RecommendationType string

const (
	RecommendationCaching        RecommendationType = "caching"
	RecommendationBatching       RecommendationType = "batching"
	RecommendationModelSelection RecommendationType = "model_selection"
	RecommendationPromptOptimization RecommendationType = "prompt_optimization"
	RecommendationBudgetTuning   RecommendationType = "budget_tuning"
)