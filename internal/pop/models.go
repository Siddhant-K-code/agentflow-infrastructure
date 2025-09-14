package pop

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PromptTemplate represents a versioned prompt template
type PromptTemplate struct {
	ID        uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrgID     uuid.UUID       `json:"org_id" gorm:"type:uuid;not null"`
	Name      string          `json:"name" gorm:"not null"`
	Version   int             `json:"version" gorm:"not null"`
	Template  string          `json:"template" gorm:"not null"`
	Schema    json.RawMessage `json:"schema" gorm:"type:jsonb"`
	Metadata  json.RawMessage `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt time.Time       `json:"created_at" gorm:"default:now()"`
}

// PromptSuite represents an evaluation suite for prompts
type PromptSuite struct {
	ID        uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrgID     uuid.UUID       `json:"org_id" gorm:"type:uuid;not null"`
	Name      string          `json:"name" gorm:"not null"`
	Cases     json.RawMessage `json:"cases" gorm:"type:jsonb;not null"`
	CreatedAt time.Time       `json:"created_at" gorm:"default:now()"`
}

// PromptDeployment represents a deployment configuration for prompts
type PromptDeployment struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrgID         uuid.UUID `json:"org_id" gorm:"type:uuid;not null"`
	PromptName    string    `json:"prompt_name" gorm:"not null"`
	StableVersion int       `json:"stable_version" gorm:"not null"`
	CanaryVersion *int      `json:"canary_version"`
	CanaryRatio   float64   `json:"canary_ratio" gorm:"default:0"`
	CreatedAt     time.Time `json:"created_at" gorm:"default:now()"`
}

// EvaluationCase represents a single test case in an evaluation suite
type EvaluationCase struct {
	Input    map[string]interface{} `json:"input"`
	Expected interface{}            `json:"expected"`
	Scoring  ScoringConfig          `json:"scoring"`
	Tags     []string               `json:"tags,omitempty"`
}

// ScoringConfig defines how to score a prompt evaluation
type ScoringConfig struct {
	Type       ScoringType            `json:"type"`
	Config     map[string]interface{} `json:"config,omitempty"`
	Threshold  float64                `json:"threshold,omitempty"`
}

// ScoringType represents different evaluation methods
type ScoringType string

const (
	ScoringExact      ScoringType = "exact"
	ScoringRegex      ScoringType = "regex"
	ScoringJSONSchema ScoringType = "json_schema"
	ScoringLLMJudge   ScoringType = "llm_judge"
	ScoringEmbedding  ScoringType = "embedding"
)

// PromptReference represents a reference to a specific prompt version
type PromptReference struct {
	Name    string `json:"name"`
	Version *int   `json:"version,omitempty"` // nil means latest
	OrgID   string `json:"org_id,omitempty"`
}

// ParsePromptRef parses a prompt reference string like "prompt_name@version"
func ParsePromptRef(ref string) PromptReference {
	// Implementation would parse strings like "doc_triage@3"
	// For now, simplified
	return PromptReference{Name: ref}
}

// TemplateEngine represents the template engine type
type TemplateEngine string

const (
	EngineHandlebars TemplateEngine = "handlebars"
	EngineMustache   TemplateEngine = "mustache"
	EngineGo         TemplateEngine = "go"
)

// PromptMetadata contains metadata about a prompt template
type PromptMetadata struct {
	Engine      TemplateEngine `json:"engine,omitempty"`
	Description string         `json:"description,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
	TokenHint   int            `json:"token_hint,omitempty"` // estimated tokens
}

// EvaluationResult represents the result of evaluating a prompt
type EvaluationResult struct {
	ID            uuid.UUID `json:"id"`
	PromptID      uuid.UUID `json:"prompt_id"`
	SuiteID       uuid.UUID `json:"suite_id"`
	Score         float64   `json:"score"`
	PassedCases   int       `json:"passed_cases"`
	TotalCases    int       `json:"total_cases"`
	Details       []CaseResult `json:"details"`
	CostCents     int64     `json:"cost_cents"`
	DurationMs    int64     `json:"duration_ms"`
	CreatedAt     time.Time `json:"created_at"`
}

// CaseResult represents the result of a single evaluation case
type CaseResult struct {
	CaseIndex int     `json:"case_index"`
	Passed    bool    `json:"passed"`
	Score     float64 `json:"score"`
	Output    string  `json:"output"`
	Error     string  `json:"error,omitempty"`
}

// CreatePromptRequest represents a request to create a new prompt version
type CreatePromptRequest struct {
	Name     string          `json:"name"`
	Template string          `json:"template"`
	Schema   json.RawMessage `json:"schema,omitempty"`
	Metadata PromptMetadata  `json:"metadata,omitempty"`
}

// EvaluateRequest represents a request to evaluate a prompt
type EvaluateRequest struct {
	PromptID uuid.UUID `json:"prompt_id"`
	SuiteID  uuid.UUID `json:"suite_id"`
	Provider string    `json:"provider,omitempty"`
	Model    string    `json:"model,omitempty"`
}

// DeploymentRequest represents a request to deploy a prompt version
type DeploymentRequest struct {
	PromptName    string  `json:"prompt_name"`
	StableVersion int     `json:"stable_version"`
	CanaryVersion *int    `json:"canary_version,omitempty"`
	CanaryRatio   float64 `json:"canary_ratio,omitempty"`
}