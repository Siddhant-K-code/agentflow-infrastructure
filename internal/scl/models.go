package scl

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ContextBundle represents a validated and sanitized context bundle
type ContextBundle struct {
	ID           uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrgID        uuid.UUID       `json:"org_id" gorm:"type:uuid;not null"`
	Hash         string          `json:"hash" gorm:"not null"`
	SchemaURI    string          `json:"schema_uri"`
	TrustScore   float64         `json:"trust_score" gorm:"default:0.5"`
	RedactionMap json.RawMessage `json:"redaction_map" gorm:"type:jsonb;default:'{}'"`
	Provenance   json.RawMessage `json:"provenance" gorm:"type:jsonb;default:'{}'"`
	CreatedAt    time.Time       `json:"created_at" gorm:"default:now()"`
}

// ContextSource represents the source of context data
type ContextSource struct {
	Type        SourceType                 `json:"type"`
	URI         string                     `json:"uri"`
	Attestation *SourceAttestation         `json:"attestation,omitempty"`
	Metadata    map[string]interface{}     `json:"metadata,omitempty"`
}

// SourceType represents different types of context sources
type SourceType string

const (
	SourceTypeFile       SourceType = "file"
	SourceTypeAPI        SourceType = "api"
	SourceTypeDatabase   SourceType = "database"
	SourceTypeUserInput  SourceType = "user_input"
	SourceTypeExternal   SourceType = "external"
)

// SourceAttestation provides cryptographic proof of source integrity
type SourceAttestation struct {
	Signature   string            `json:"signature"`
	Certificate string            `json:"certificate"`
	Timestamp   time.Time         `json:"timestamp"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ValidationResult represents the result of context validation
type ValidationResult struct {
	Valid       bool              `json:"valid"`
	Errors      []ValidationError `json:"errors,omitempty"`
	Warnings    []string          `json:"warnings,omitempty"`
	TrustScore  float64           `json:"trust_score"`
	RedactedFields []string       `json:"redacted_fields,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ContentFilter represents different types of content filtering
type ContentFilter interface {
	Filter(content string) FilterResult
}

// FilterResult represents the result of content filtering
type FilterResult struct {
	Allowed    bool                   `json:"allowed"`
	Confidence float64                `json:"confidence"`
	Reason     string                 `json:"reason,omitempty"`
	Redactions []RedactionRange       `json:"redactions,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// RedactionRange represents a range of text to be redacted
type RedactionRange struct {
	Start  int    `json:"start"`
	End    int    `json:"end"`
	Type   string `json:"type"`   // PII type (SSN, email, etc.)
	Token  string `json:"token"`  // replacement token
}

// PolicyDecision represents an authorization policy decision
type PolicyDecision struct {
	Allowed    bool              `json:"allowed"`
	Reason     string            `json:"reason,omitempty"`
	Conditions map[string]string `json:"conditions,omitempty"`
}

// PIIRedactor handles PII detection and redaction
type PIIRedactor struct {
	Patterns map[string]string // regex patterns for different PII types
	Enabled  bool
}

// InjectionDetector detects prompt injection attempts
type InjectionDetector struct {
	Patterns   []string // known injection patterns
	Threshold  float64  // confidence threshold
	ModelBased bool     // use LLM for detection
	Enabled    bool     // whether detection is enabled
}

// IngestRequest represents a request to ingest context
type IngestRequest struct {
	Source    ContextSource              `json:"source"`
	Content   string                     `json:"content"`
	SchemaURI string                     `json:"schema_uri,omitempty"`
	Metadata  map[string]interface{}     `json:"metadata,omitempty"`
}

// IngestResponse represents the response from context ingestion
type IngestResponse struct {
	BundleID    uuid.UUID        `json:"bundle_id"`
	TrustScore  float64          `json:"trust_score"`
	Validation  ValidationResult `json:"validation"`
	Hash        string           `json:"hash"`
}

// PrepareRequest represents a request to prepare context for use
type PrepareRequest struct {
	BundleIDs []uuid.UUID                `json:"bundle_ids"`
	Query     string                     `json:"query,omitempty"`
	MaxChunks int                        `json:"max_chunks,omitempty"`
	Filters   map[string]interface{}     `json:"filters,omitempty"`
}

// PrepareResponse represents prepared context ready for consumption
type PrepareResponse struct {
	Chunks     []ContextChunk             `json:"chunks"`
	Citations  []Citation                 `json:"citations"`
	Metadata   map[string]interface{}     `json:"metadata"`
}

// ContextChunk represents a filtered and ranked piece of context
type ContextChunk struct {
	Content     string                 `json:"content"`
	TrustScore  float64                `json:"trust_score"`
	Relevance   float64                `json:"relevance"`
	Source      ContextSource          `json:"source"`
	Provenance  map[string]interface{} `json:"provenance"`
}

// Citation provides attribution for context usage
type Citation struct {
	BundleID   uuid.UUID `json:"bundle_id"`
	ChunkIndex int       `json:"chunk_index"`
	Source     string    `json:"source"`
	Timestamp  time.Time `json:"timestamp"`
}

// PolicyTestRequest represents a request to test policy against sample data
type PolicyTestRequest struct {
	Policy    string                     `json:"policy"`
	Subject   string                     `json:"subject"`   // user/role
	Resource  string                     `json:"resource"`  // what is being accessed
	Action    string                     `json:"action"`    // read/write/execute
	Context   map[string]interface{}     `json:"context,omitempty"`
}

// PolicyTestResponse represents the result of policy testing
type PolicyTestResponse struct {
	Decision    PolicyDecision         `json:"decision"`
	Explanation string                 `json:"explanation"`
	Trace       []PolicyTraceStep      `json:"trace,omitempty"`
}

// PolicyTraceStep represents a step in policy evaluation
type PolicyTraceStep struct {
	Rule      string `json:"rule"`
	Result    bool   `json:"result"`
	Reason    string `json:"reason"`
}

// ComputeContentHash computes a SHA-256 hash of content
func ComputeContentHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// NewPIIRedactor creates a new PII redactor with common patterns
func NewPIIRedactor() *PIIRedactor {
	return &PIIRedactor{
		Patterns: map[string]string{
			"ssn":          `\b\d{3}-\d{2}-\d{4}\b`,
			"email":        `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`,
			"phone":        `\b\d{3}-\d{3}-\d{4}\b`,
			"credit_card":  `\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b`,
			"ip_address":   `\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`,
		},
		Enabled: true,
	}
}

// NewInjectionDetector creates a new injection detector
func NewInjectionDetector() *InjectionDetector {
	return &InjectionDetector{
		Patterns: []string{
			"ignore previous instructions",
			"system prompt",
			"forget everything",
			"new instructions",
			"override",
		},
		Threshold:  0.7,
		ModelBased: false,
		Enabled:    true,
	}
}