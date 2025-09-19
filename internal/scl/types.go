package scl

import (
	"time"

	"github.com/google/uuid"
)

// ContextBundle represents a processed and validated context bundle
type ContextBundle struct {
	ID           uuid.UUID         `json:"id" db:"id"`
	OrgID        uuid.UUID         `json:"org_id" db:"org_id"`
	Hash         string            `json:"hash" db:"hash"`
	SchemaURI    string            `json:"schema_uri" db:"schema_uri"`
	TrustScore   float64           `json:"trust_score" db:"trust_score"`
	RedactionMap map[string]string `json:"redaction_map" db:"redaction_map"`
	Provenance   Provenance        `json:"provenance" db:"provenance"`
	CreatedAt    time.Time         `json:"created_at" db:"created_at"`
}

// Provenance tracks the source and processing history of context
type Provenance struct {
	Sources      []Source               `json:"sources"`
	Attestations []Attestation          `json:"attestations"`
	Processing   []ProcessingStep       `json:"processing"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type Source struct {
	ID        string                 `json:"id"`
	Type      SourceType             `json:"type"`
	URI       string                 `json:"uri"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type SourceType string

const (
	SourceTypeFile     SourceType = "file"
	SourceTypeAPI      SourceType = "api"
	SourceTypeDatabase SourceType = "database"
	SourceTypeUser     SourceType = "user"
	SourceTypeExternal SourceType = "external"
)

type Attestation struct {
	Type      AttestationType        `json:"type"`
	Signature string                 `json:"signature"`
	PublicKey string                 `json:"public_key"`
	Timestamp time.Time              `json:"timestamp"`
	Claims    map[string]interface{} `json:"claims,omitempty"`
}

type AttestationType string

const (
	AttestationSigstore AttestationType = "sigstore"
	AttestationGPG      AttestationType = "gpg"
	AttestationJWT      AttestationType = "jwt"
	AttestationCustom   AttestationType = "custom"
)

type ProcessingStep struct {
	Stage     ProcessingStage        `json:"stage"`
	Timestamp time.Time              `json:"timestamp"`
	Result    ProcessingResult       `json:"result"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type ProcessingStage string

const (
	StageValidation   ProcessingStage = "validation"
	StageSanitization ProcessingStage = "sanitization"
	StageRedaction    ProcessingStage = "redaction"
	StagePolicy       ProcessingStage = "policy"
	StageEnrichment   ProcessingStage = "enrichment"
)

type ProcessingResult struct {
	Status   ProcessingStatus       `json:"status"`
	Message  string                 `json:"message,omitempty"`
	Warnings []string               `json:"warnings,omitempty"`
	Errors   []string               `json:"errors,omitempty"`
	Metrics  map[string]interface{} `json:"metrics,omitempty"`
}

type ProcessingStatus string

const (
	StatusPassed  ProcessingStatus = "passed"
	StatusWarning ProcessingStatus = "warning"
	StatusFailed  ProcessingStatus = "failed"
	StatusSkipped ProcessingStatus = "skipped"
)

// IngestRequest represents a request to ingest and process context
type IngestRequest struct {
	Content     interface{}            `json:"content"`
	SchemaURI   string                 `json:"schema_uri,omitempty"`
	Sources     []Source               `json:"sources,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	PolicyHints map[string]interface{} `json:"policy_hints,omitempty"`
}

// IngestResponse represents the result of context ingestion
type IngestResponse struct {
	BundleID       uuid.UUID        `json:"bundle_id"`
	Hash           string           `json:"hash"`
	TrustScore     float64          `json:"trust_score"`
	Status         ProcessingStatus `json:"status"`
	Warnings       []string         `json:"warnings,omitempty"`
	Errors         []string         `json:"errors,omitempty"`
	ProcessingTime time.Duration    `json:"processing_time"`
}

// PrepareRequest represents a request to prepare context for use
type PrepareRequest struct {
	BundleIDs []uuid.UUID            `json:"bundle_ids"`
	Query     string                 `json:"query,omitempty"`
	MaxChunks int                    `json:"max_chunks,omitempty"`
	Policy    map[string]interface{} `json:"policy,omitempty"`
}

// PrepareResponse represents prepared context ready for use
type PrepareResponse struct {
	Chunks     []ContextChunk         `json:"chunks"`
	Citations  []Citation             `json:"citations"`
	TrustScore float64                `json:"trust_score"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type ContextChunk struct {
	ID         string                 `json:"id"`
	Content    string                 `json:"content"`
	BundleID   uuid.UUID              `json:"bundle_id"`
	Rank       float64                `json:"rank"`
	TrustScore float64                `json:"trust_score"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type Citation struct {
	ChunkID   string    `json:"chunk_id"`
	BundleID  uuid.UUID `json:"bundle_id"`
	Source    Source    `json:"source"`
	Relevance float64   `json:"relevance"`
}

// PolicyTestRequest represents a request to test policies
type PolicyTestRequest struct {
	Policy  map[string]interface{} `json:"policy"`
	Content interface{}            `json:"content"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// PolicyTestResponse represents the result of policy testing
type PolicyTestResponse struct {
	Allowed    bool                   `json:"allowed"`
	Reason     string                 `json:"reason,omitempty"`
	Violations []PolicyViolation      `json:"violations,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type PolicyViolation struct {
	Rule       string `json:"rule"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	Location   string `json:"location,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

// RedactionConfig configures how PII and sensitive data is handled
type RedactionConfig struct {
	Enabled    bool               `json:"enabled"`
	Patterns   []RedactionPattern `json:"patterns"`
	Reversible bool               `json:"reversible"`
	KMSKeyID   string             `json:"kms_key_id,omitempty"`
}

type RedactionPattern struct {
	Name        string  `json:"name"`
	Pattern     string  `json:"pattern"`
	Replacement string  `json:"replacement"`
	Confidence  float64 `json:"confidence"`
}

// InjectionFilter configures prompt injection detection
type InjectionFilter struct {
	Enabled    bool            `json:"enabled"`
	Heuristics []HeuristicRule `json:"heuristics"`
	Embeddings EmbeddingConfig `json:"embeddings,omitempty"`
	LLMJudge   LLMJudgeConfig  `json:"llm_judge,omitempty"`
}

type HeuristicRule struct {
	Name        string  `json:"name"`
	Pattern     string  `json:"pattern"`
	Confidence  float64 `json:"confidence"`
	Description string  `json:"description"`
}

type EmbeddingConfig struct {
	Model      string  `json:"model"`
	Threshold  float64 `json:"threshold"`
	CorpusPath string  `json:"corpus_path"`
}

type LLMJudgeConfig struct {
	Model     string  `json:"model"`
	Threshold float64 `json:"threshold"`
	Prompt    string  `json:"prompt"`
}
