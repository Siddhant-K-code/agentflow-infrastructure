package pop

import (
	"time"

	"github.com/google/uuid"
)

// PromptTemplate represents a versioned prompt template
type PromptTemplate struct {
	ID       uuid.UUID `json:"id" db:"id"`
	OrgID    uuid.UUID `json:"org_id" db:"org_id"`
	Name     string    `json:"name" db:"name"`
	Version  int       `json:"version" db:"version"`
	Template string    `json:"template" db:"template"`
	Schema   Schema    `json:"schema" db:"schema"`
	Metadata Metadata  `json:"metadata" db:"metadata"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Schema defines the input schema for a prompt template
type Schema struct {
	Type       string                 `json:"type"`
	Properties map[string]Property    `json:"properties"`
	Required   []string               `json:"required"`
	Additional map[string]interface{} `json:"additionalProperties,omitempty"`
}

type Property struct {
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
}

// PromptSuite represents an evaluation suite
type PromptSuite struct {
	ID        uuid.UUID   `json:"id" db:"id"`
	OrgID     uuid.UUID   `json:"org_id" db:"org_id"`
	Name      string      `json:"name" db:"name"`
	Cases     []TestCase  `json:"cases" db:"cases"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
}

// TestCase represents a single test case in an evaluation suite
type TestCase struct {
	ID       string                 `json:"id"`
	Input    map[string]interface{} `json:"input"`
	Expected Expected               `json:"expected"`
	Scoring  ScoringConfig          `json:"scoring"`
}

type Expected struct {
	Output   interface{}            `json:"output,omitempty"`
	Contains []string               `json:"contains,omitempty"`
	Schema   map[string]interface{} `json:"schema,omitempty"`
	Metrics  map[string]float64     `json:"metrics,omitempty"`
}

type ScoringConfig struct {
	Type     ScoringType            `json:"type"`
	Config   map[string]interface{} `json:"config,omitempty"`
	Weight   float64                `json:"weight,omitempty"`
}

type ScoringType string

const (
	ScoringExact      ScoringType = "exact"
	ScoringContains   ScoringType = "contains"
	ScoringRegex      ScoringType = "regex"
	ScoringSchema     ScoringType = "schema"
	ScoringLLMJudge   ScoringType = "llm_judge"
	ScoringEmbedding  ScoringType = "embedding"
)

// PromptDeployment represents a deployment configuration
type PromptDeployment struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	OrgID         uuid.UUID  `json:"org_id" db:"org_id"`
	PromptName    string     `json:"prompt_name" db:"prompt_name"`
	StableVersion int        `json:"stable_version" db:"stable_version"`
	CanaryVersion *int       `json:"canary_version" db:"canary_version"`
	CanaryRatio   float64    `json:"canary_ratio" db:"canary_ratio"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// EvaluationRun represents a single evaluation execution
type EvaluationRun struct {
	ID           uuid.UUID            `json:"id"`
	PromptID     uuid.UUID            `json:"prompt_id"`
	SuiteID      uuid.UUID            `json:"suite_id"`
	Status       EvaluationStatus     `json:"status"`
	Results      []EvaluationResult   `json:"results"`
	Summary      EvaluationSummary    `json:"summary"`
	StartedAt    time.Time            `json:"started_at"`
	CompletedAt  *time.Time           `json:"completed_at,omitempty"`
}

type EvaluationStatus string

const (
	EvaluationStatusQueued    EvaluationStatus = "queued"
	EvaluationStatusRunning   EvaluationStatus = "running"
	EvaluationStatusCompleted EvaluationStatus = "completed"
	EvaluationStatusFailed    EvaluationStatus = "failed"
)

type EvaluationResult struct {
	CaseID    string                 `json:"case_id"`
	Input     map[string]interface{} `json:"input"`
	Output    interface{}            `json:"output"`
	Expected  Expected               `json:"expected"`
	Score     float64                `json:"score"`
	Passed    bool                   `json:"passed"`
	Error     string                 `json:"error,omitempty"`
	Latency   time.Duration          `json:"latency"`
	CostCents int64                  `json:"cost_cents"`
}

type EvaluationSummary struct {
	TotalCases    int     `json:"total_cases"`
	PassedCases   int     `json:"passed_cases"`
	FailedCases   int     `json:"failed_cases"`
	AverageScore  float64 `json:"average_score"`
	TotalCost     int64   `json:"total_cost_cents"`
	AverageLatency time.Duration `json:"average_latency"`
}

// PromptRequest represents a request to resolve a prompt
type PromptRequest struct {
	Name     string                 `json:"name"`
	Version  *int                   `json:"version,omitempty"` // nil for latest deployment
	Inputs   map[string]interface{} `json:"inputs"`
	Context  map[string]interface{} `json:"context,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// PromptResponse represents a resolved prompt
type PromptResponse struct {
	ID           uuid.UUID              `json:"id"`
	Name         string                 `json:"name"`
	Version      int                    `json:"version"`
	RenderedText string                 `json:"rendered_text"`
	Metadata     map[string]interface{} `json:"metadata"`
	TokenCount   int                    `json:"token_count"`
	IsCanary     bool                   `json:"is_canary"`
}

// Metadata is a flexible JSON field
type Metadata map[string]interface{}

// CreatePromptRequest represents a request to create a new prompt version
type CreatePromptRequest struct {
	Name     string   `json:"name"`
	Template string   `json:"template"`
	Schema   Schema   `json:"schema"`
	Metadata Metadata `json:"metadata,omitempty"`
}

// EvaluateRequest represents a request to evaluate a prompt
type EvaluateRequest struct {
	PromptName    string `json:"prompt_name"`
	PromptVersion int    `json:"prompt_version"`
	SuiteName     string `json:"suite_name"`
	Parallel      int    `json:"parallel,omitempty"`
}

// DeploymentRequest represents a request to update deployment
type DeploymentRequest struct {
	PromptName    string  `json:"prompt_name"`
	StableVersion int     `json:"stable_version"`
	CanaryVersion *int    `json:"canary_version,omitempty"`
	CanaryRatio   float64 `json:"canary_ratio,omitempty"`
}