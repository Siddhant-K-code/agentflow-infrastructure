package scl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"time"

	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/config"
	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/db"
)

type Service struct {
	cfg       *config.Config
	db        *db.PostgresDB
	validator *Validator
	sanitizer *Sanitizer
	redactor  *Redactor
	policy    *PolicyEngine
}

func NewService(cfg *config.Config, database *db.PostgresDB) *Service {
	return &Service{
		cfg:       cfg,
		db:        database,
		validator: NewValidator(),
		sanitizer: NewSanitizer(),
		redactor:  NewRedactor(),
		policy:    NewPolicyEngine(),
	}
}

// IngestContext processes and validates untrusted context
func (s *Service) IngestContext(ctx context.Context, orgID uuid.UUID, req *IngestRequest) (*IngestResponse, error) {
	start := time.Now()

	response := &IngestResponse{
		BundleID: uuid.New(),
		Status:   StatusPassed,
		Warnings: make([]string, 0),
		Errors:   make([]string, 0),
	}

	// Create context bundle
	bundle := &ContextBundle{
		ID:           response.BundleID,
		OrgID:        orgID,
		TrustScore:   0.5, // Default trust score
		RedactionMap: make(map[string]string),
		Provenance: Provenance{
			Sources:    req.Sources,
			Processing: make([]ProcessingStep, 0),
			Metadata:   req.Metadata,
		},
		CreatedAt: time.Now(),
	}

	// Step 1: Source validation
	if err := s.validateSources(ctx, bundle, req); err != nil {
		response.Errors = append(response.Errors, fmt.Sprintf("Source validation failed: %v", err))
		response.Status = StatusFailed
	}

	// Step 2: Schema validation
	if req.SchemaURI != "" {
		if err := s.validator.ValidateSchema(req.Content, req.SchemaURI); err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Schema validation failed: %v", err))
			response.Status = StatusFailed
		} else {
			bundle.SchemaURI = req.SchemaURI
		}
	}

	// Step 3: Content sanitization
	sanitizedContent, warnings, err := s.sanitizer.Sanitize(req.Content)
	if err != nil {
		response.Errors = append(response.Errors, fmt.Sprintf("Sanitization failed: %v", err))
		response.Status = StatusFailed
	} else {
		response.Warnings = append(response.Warnings, warnings...)
		req.Content = sanitizedContent
	}

	// Step 4: PII redaction
	redactedContent, redactionMap, err := s.redactor.Redact(req.Content)
	if err != nil {
		response.Errors = append(response.Errors, fmt.Sprintf("Redaction failed: %v", err))
		response.Status = StatusFailed
	} else {
		bundle.RedactionMap = redactionMap
		req.Content = redactedContent
	}

	// Step 5: Policy enforcement
	policyResult, err := s.policy.Evaluate(ctx, orgID, req.Content, req.PolicyHints)
	if err != nil {
		response.Errors = append(response.Errors, fmt.Sprintf("Policy evaluation failed: %v", err))
		response.Status = StatusFailed
	} else if !policyResult.Allowed {
		response.Errors = append(response.Errors, fmt.Sprintf("Policy violation: %s", policyResult.Reason))
		response.Status = StatusFailed
	}

	// Step 6: Trust score calculation
	bundle.TrustScore = s.calculateTrustScore(bundle, policyResult)
	response.TrustScore = bundle.TrustScore

	// Step 7: Generate content hash
	contentBytes, _ := json.Marshal(req.Content)
	hash := sha256.Sum256(contentBytes)
	bundle.Hash = hex.EncodeToString(hash[:])
	response.Hash = bundle.Hash

	// Save bundle if processing succeeded
	if response.Status != StatusFailed {
		if err := s.saveContextBundle(ctx, bundle); err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Failed to save bundle: %v", err))
			response.Status = StatusFailed
		}
	}

	response.ProcessingTime = time.Since(start)
	return response, nil
}

// PrepareContext prepares context bundles for use in agent execution
func (s *Service) PrepareContext(ctx context.Context, orgID uuid.UUID, req *PrepareRequest) (*PrepareResponse, error) {
	response := &PrepareResponse{
		Chunks:    make([]ContextChunk, 0),
		Citations: make([]Citation, 0),
		Metadata:  make(map[string]interface{}),
	}

	// Get context bundles
	bundles, err := s.getContextBundles(ctx, orgID, req.BundleIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get context bundles: %w", err)
	}

	// Apply policy filtering
	filteredBundles := make([]*ContextBundle, 0)
	for _, bundle := range bundles {
		allowed, err := s.policy.CheckAccess(ctx, orgID, bundle.ID, req.Policy)
		if err != nil {
			continue // Skip on error
		}
		if allowed {
			filteredBundles = append(filteredBundles, bundle)
		}
	}

	// Rank and select chunks
	chunks := s.rankAndSelectChunks(filteredBundles, req.Query, req.MaxChunks)
	response.Chunks = chunks

	// Generate citations
	response.Citations = s.generateCitations(chunks, filteredBundles)

	// Calculate overall trust score
	response.TrustScore = s.calculateOverallTrustScore(filteredBundles)

	return response, nil
}

// TestPolicy tests policy rules against sample content
func (s *Service) TestPolicy(ctx context.Context, orgID uuid.UUID, req *PolicyTestRequest) (*PolicyTestResponse, error) {
	result, err := s.policy.Evaluate(ctx, orgID, req.Content, req.Policy)
	if err != nil {
		return nil, fmt.Errorf("policy evaluation failed: %w", err)
	}

	response := &PolicyTestResponse{
		Allowed:    result.Allowed,
		Reason:     result.Reason,
		Violations: make([]PolicyViolation, 0),
		Metadata:   result.Metadata,
	}

	// Convert policy violations
	for _, violation := range result.Violations {
		response.Violations = append(response.Violations, PolicyViolation(violation))
	}

	return response, nil
}

// Helper methods

func (s *Service) validateSources(ctx context.Context, bundle *ContextBundle, req *IngestRequest) error {
	for _, source := range req.Sources {
		// Validate source attestations
		if err := s.validateSourceAttestations(source); err != nil {
			return fmt.Errorf("source %s attestation failed: %w", source.ID, err)
		}

		// Add processing step
		bundle.Provenance.Processing = append(bundle.Provenance.Processing, ProcessingStep{
			Stage:     StageValidation,
			Timestamp: time.Now(),
			Result: ProcessingResult{
				Status:  StatusPassed,
				Message: fmt.Sprintf("Source %s validated", source.ID),
			},
		})
	}

	return nil
}

func (s *Service) validateSourceAttestations(source Source) error {
	// Mock attestation validation
	// In production, this would verify signatures, certificates, etc.
	if source.Type == SourceTypeExternal {
		// External sources require stronger validation
		return fmt.Errorf("external source validation not implemented")
	}
	return nil
}

func (s *Service) calculateTrustScore(bundle *ContextBundle, policyResult *PolicyResult) float64 {
	score := 0.5 // Base score

	// Adjust based on source types
	for _, source := range bundle.Provenance.Sources {
		switch source.Type {
		case SourceTypeFile, SourceTypeDatabase:
			score += 0.2
		case SourceTypeAPI:
			score += 0.1
		case SourceTypeUser:
			score += 0.05
		case SourceTypeExternal:
			score -= 0.1
		}
	}

	// Adjust based on policy compliance
	if policyResult.Allowed {
		score += 0.2
	} else {
		score -= 0.3
	}

	// Adjust based on attestations
	if len(bundle.Provenance.Attestations) > 0 {
		score += 0.1
	}

	// Clamp to [0, 1]
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}

func (s *Service) saveContextBundle(ctx context.Context, bundle *ContextBundle) error {
	provenanceJSON, err := json.Marshal(bundle.Provenance)
	if err != nil {
		return err
	}

	redactionMapJSON, err := json.Marshal(bundle.RedactionMap)
	if err != nil {
		return err
	}

	query := `INSERT INTO context_bundle (id, org_id, hash, schema_uri, trust_score, redaction_map, provenance, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = s.db.ExecContext(ctx, query,
		bundle.ID, bundle.OrgID, bundle.Hash, bundle.SchemaURI,
		bundle.TrustScore, redactionMapJSON, provenanceJSON, bundle.CreatedAt,
	)
	return err
}

func (s *Service) getContextBundles(ctx context.Context, orgID uuid.UUID, bundleIDs []uuid.UUID) ([]*ContextBundle, error) {
	if len(bundleIDs) == 0 {
		return []*ContextBundle{}, nil
	}

	// Build query with IN clause
	query := `SELECT id, org_id, hash, schema_uri, trust_score, redaction_map, provenance, created_at
			  FROM context_bundle WHERE org_id = $1 AND id = ANY($2)`

	rows, err := s.db.QueryContext(ctx, query, orgID, bundleIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bundles := make([]*ContextBundle, 0)
	for rows.Next() {
		var bundle ContextBundle
		var redactionMapJSON, provenanceJSON []byte

		err := rows.Scan(
			&bundle.ID, &bundle.OrgID, &bundle.Hash, &bundle.SchemaURI,
			&bundle.TrustScore, &redactionMapJSON, &provenanceJSON, &bundle.CreatedAt,
		)
		if err != nil {
			continue
		}

		if err := json.Unmarshal(redactionMapJSON, &bundle.RedactionMap); err != nil {
			continue
		}

		if err := json.Unmarshal(provenanceJSON, &bundle.Provenance); err != nil {
			continue
		}

		bundles = append(bundles, &bundle)
	}

	return bundles, nil
}

func (s *Service) rankAndSelectChunks(bundles []*ContextBundle, query string, maxChunks int) []ContextChunk {
	chunks := make([]ContextChunk, 0)

	// Mock chunk generation and ranking
	// In production, this would use embeddings, BM25, etc.
	for i, bundle := range bundles {
		if maxChunks > 0 && len(chunks) >= maxChunks {
			break
		}

		chunk := ContextChunk{
			ID:         fmt.Sprintf("chunk_%d", i),
			Content:    fmt.Sprintf("Mock content from bundle %s", bundle.ID),
			BundleID:   bundle.ID,
			Rank:       1.0 - float64(i)*0.1, // Mock ranking
			TrustScore: bundle.TrustScore,
			Metadata:   map[string]interface{}{"bundle_hash": bundle.Hash},
		}

		chunks = append(chunks, chunk)
	}

	return chunks
}

func (s *Service) generateCitations(chunks []ContextChunk, bundles []*ContextBundle) []Citation {
	citations := make([]Citation, 0)

	bundleMap := make(map[uuid.UUID]*ContextBundle)
	for _, bundle := range bundles {
		bundleMap[bundle.ID] = bundle
	}

	for _, chunk := range chunks {
		bundle, exists := bundleMap[chunk.BundleID]
		if !exists {
			continue
		}

		for _, source := range bundle.Provenance.Sources {
			citation := Citation{
				ChunkID:   chunk.ID,
				BundleID:  chunk.BundleID,
				Source:    source,
				Relevance: chunk.Rank,
			}
			citations = append(citations, citation)
		}
	}

	return citations
}

func (s *Service) calculateOverallTrustScore(bundles []*ContextBundle) float64 {
	if len(bundles) == 0 {
		return 0.0
	}

	total := 0.0
	for _, bundle := range bundles {
		total += bundle.TrustScore
	}

	return total / float64(len(bundles))
}
