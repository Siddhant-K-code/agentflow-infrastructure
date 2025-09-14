package scl

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service implements the Secure Context Layer
type Service struct {
	db               *gorm.DB
	piiRedactor      *PIIRedactor
	injectionDetector *InjectionDetector
}

// NewService creates a new SCL service
func NewService(db *gorm.DB) *Service {
	return &Service{
		db:               db,
		piiRedactor:      NewPIIRedactor(),
		injectionDetector: NewInjectionDetector(),
	}
}

// IngestContext ingests and validates context data
func (s *Service) IngestContext(c *gin.Context) {
	var req IngestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // placeholder

	// Validate content against schema if provided
	validation := s.validateContent(req.Content, req.SchemaURI)
	if !validation.Valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "validation failed",
			"details": validation,
		})
		return
	}

	// Apply content filters
	filterResult := s.applyContentFilters(req.Content)
	if !filterResult.Allowed {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "content filtered",
			"reason": filterResult.Reason,
		})
		return
	}

	// Redact PII
	redactedContent, redactions := s.redactPII(req.Content)
	
	// Compute hash
	hash := ComputeContentHash(redactedContent)

	// Create provenance record
	provenance := map[string]interface{}{
		"source": req.Source,
		"ingested_at": "now", // would be actual timestamp
		"validation": validation,
		"filters_applied": []string{"pii", "injection"},
	}

	// Store bundle
	bundle := ContextBundle{
		OrgID:        orgID,
		Hash:         hash,
		SchemaURI:    req.SchemaURI,
		TrustScore:   s.calculateTrustScore(req.Source, validation, filterResult),
		RedactionMap: s.serializeRedactions(redactions),
		Provenance:   s.serializeProvenance(provenance),
	}

	if err := s.db.Create(&bundle).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := IngestResponse{
		BundleID:   bundle.ID,
		TrustScore: bundle.TrustScore,
		Validation: validation,
		Hash:       hash,
	}

	c.JSON(http.StatusCreated, response)
}

// PrepareContext prepares context bundles for consumption
func (s *Service) PrepareContext(c *gin.Context) {
	var req PrepareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retrieve bundles
	var bundles []ContextBundle
	if err := s.db.Where("id IN ?", req.BundleIDs).Find(&bundles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Rank and filter chunks based on query relevance and trust score
	chunks := s.rankAndFilterChunks(bundles, req.Query, req.MaxChunks)
	
	// Generate citations
	citations := s.generateCitations(bundles, chunks)

	response := PrepareResponse{
		Chunks:    chunks,
		Citations: citations,
		Metadata: map[string]interface{}{
			"total_bundles": len(bundles),
			"selected_chunks": len(chunks),
		},
	}

	c.JSON(http.StatusOK, response)
}

// TestPolicy tests a policy against sample data
func (s *Service) TestPolicy(c *gin.Context) {
	var req PolicyTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Evaluate policy (simplified - would use OpenFGA or OPA in practice)
	decision := s.evaluatePolicy(req)
	
	response := PolicyTestResponse{
		Decision:    decision,
		Explanation: "Policy evaluation based on rules",
		Trace:       []PolicyTraceStep{}, // simplified
	}

	c.JSON(http.StatusOK, response)
}

// validateContent validates content against schema
func (s *Service) validateContent(content, schemaURI string) ValidationResult {
	// Simplified validation - would use JSONSchema validation in practice
	result := ValidationResult{
		Valid:      true,
		TrustScore: 0.8,
	}

	if len(content) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "content",
			Message: "content cannot be empty",
			Code:    "EMPTY_CONTENT",
		})
	}

	return result
}

// applyContentFilters applies content filtering
func (s *Service) applyContentFilters(content string) FilterResult {
	// Check for injection attempts
	if s.injectionDetector.Enabled {
		for _, pattern := range s.injectionDetector.Patterns {
			if strings.Contains(strings.ToLower(content), pattern) {
				return FilterResult{
					Allowed:    false,
					Confidence: 0.9,
					Reason:     "potential prompt injection detected",
				}
			}
		}
	}

	return FilterResult{
		Allowed:    true,
		Confidence: 0.95,
	}
}

// redactPII detects and redacts PII from content
func (s *Service) redactPII(content string) (string, []RedactionRange) {
	if !s.piiRedactor.Enabled {
		return content, nil
	}

	redacted := content
	var redactions []RedactionRange

	for piiType, pattern := range s.piiRedactor.Patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringIndex(redacted, -1)
		
		for _, match := range matches {
			token := "[REDACTED_" + strings.ToUpper(piiType) + "]"
			redactions = append(redactions, RedactionRange{
				Start: match[0],
				End:   match[1],
				Type:  piiType,
				Token: token,
			})
			
			// Replace in reverse order to maintain indices
			redacted = redacted[:match[0]] + token + redacted[match[1]:]
		}
	}

	return redacted, redactions
}

// calculateTrustScore computes a trust score for the context
func (s *Service) calculateTrustScore(source ContextSource, validation ValidationResult, filter FilterResult) float64 {
	score := 0.5 // base score

	// Source type affects trust
	switch source.Type {
	case SourceTypeFile:
		score += 0.2
	case SourceTypeAPI:
		score += 0.1
	case SourceTypeUserInput:
		score -= 0.1
	case SourceTypeExternal:
		score -= 0.2
	}

	// Attestation boosts trust
	if source.Attestation != nil {
		score += 0.2
	}

	// Validation affects trust
	if validation.Valid {
		score += 0.1
	} else {
		score -= 0.3
	}

	// Filter confidence affects trust
	score += (filter.Confidence - 0.5) * 0.2

	// Clamp to [0, 1]
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}

// rankAndFilterChunks ranks context chunks by relevance and trust
func (s *Service) rankAndFilterChunks(bundles []ContextBundle, query string, maxChunks int) []ContextChunk {
	var chunks []ContextChunk

	for _, bundle := range bundles {
		// Simplified chunking - would use semantic chunking in practice
		chunk := ContextChunk{
			Content:    "chunk content from bundle " + bundle.ID.String(), // placeholder
			TrustScore: bundle.TrustScore,
			Relevance:  0.8, // would calculate based on query similarity
			Source: ContextSource{
				Type: SourceTypeFile, // would extract from provenance
				URI:  "source_uri",
			},
			Provenance: map[string]interface{}{
				"bundle_id": bundle.ID,
				"hash":      bundle.Hash,
			},
		}
		chunks = append(chunks, chunk)
	}

	// Sort by relevance * trust_score (simplified ranking)
	// Would use more sophisticated ranking in practice

	// Limit to maxChunks
	if maxChunks > 0 && len(chunks) > maxChunks {
		chunks = chunks[:maxChunks]
	}

	return chunks
}

// generateCitations creates citations for context usage
func (s *Service) generateCitations(bundles []ContextBundle, chunks []ContextChunk) []Citation {
	var citations []Citation

	for i, chunk := range chunks {
		citation := Citation{
			ChunkIndex: i,
			Source:     chunk.Source.URI,
			// BundleID and Timestamp would be extracted from provenance
		}
		citations = append(citations, citation)
	}

	return citations
}

// evaluatePolicy evaluates a policy request (simplified)
func (s *Service) evaluatePolicy(req PolicyTestRequest) PolicyDecision {
	// Simplified policy evaluation - would use OpenFGA/OPA in practice
	decision := PolicyDecision{
		Allowed: true,
		Reason:  "allowed by default policy",
	}

	// Basic checks
	if req.Subject == "" {
		decision.Allowed = false
		decision.Reason = "missing subject"
	}

	return decision
}

// Helper methods for serialization
func (s *Service) serializeRedactions(redactions []RedactionRange) []byte {
	data, _ := json.Marshal(redactions)
	return data
}

func (s *Service) serializeProvenance(provenance map[string]interface{}) []byte {
	data, _ := json.Marshal(provenance)
	return data
}