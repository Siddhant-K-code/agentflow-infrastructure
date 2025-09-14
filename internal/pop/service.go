package pop

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/agentflow/infrastructure/internal/config"
	"github.com/agentflow/infrastructure/internal/db"
	"github.com/google/uuid"
)

type Service struct {
	cfg      *config.Config
	db       *db.PostgresDB
	renderer *TemplateRenderer
	evaluator *Evaluator
}

func NewService(cfg *config.Config, database *db.PostgresDB) *Service {
	return &Service{
		cfg:       cfg,
		db:        database,
		renderer:  NewTemplateRenderer(),
		evaluator: NewEvaluator(database),
	}
}

// CreatePromptVersion creates a new version of a prompt template
func (s *Service) CreatePromptVersion(ctx context.Context, orgID uuid.UUID, req *CreatePromptRequest) (*PromptTemplate, error) {
	// Get next version number
	nextVersion, err := s.getNextVersion(ctx, orgID, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get next version: %w", err)
	}

	// Validate template
	if err := s.renderer.Validate(req.Template, req.Schema); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}

	// Create prompt template
	prompt := &PromptTemplate{
		ID:        uuid.New(),
		OrgID:     orgID,
		Name:      req.Name,
		Version:   nextVersion,
		Template:  req.Template,
		Schema:    req.Schema,
		Metadata:  req.Metadata,
		CreatedAt: time.Now(),
	}

	if err := s.savePromptTemplate(ctx, prompt); err != nil {
		return nil, fmt.Errorf("failed to save prompt template: %w", err)
	}

	return prompt, nil
}

// GetPromptTemplate retrieves a specific prompt template version
func (s *Service) GetPromptTemplate(ctx context.Context, orgID uuid.UUID, name string, version int) (*PromptTemplate, error) {
	query := `SELECT id, org_id, name, version, template, schema, metadata, created_at
			  FROM prompt_template 
			  WHERE org_id = $1 AND name = $2 AND version = $3`

	var prompt PromptTemplate
	var schemaJSON, metadataJSON []byte

	err := s.db.QueryRowContext(ctx, query, orgID, name, version).Scan(
		&prompt.ID, &prompt.OrgID, &prompt.Name, &prompt.Version,
		&prompt.Template, &schemaJSON, &metadataJSON, &prompt.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompt template: %w", err)
	}

	if err := json.Unmarshal(schemaJSON, &prompt.Schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &prompt.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &prompt, nil
}

// ResolvePrompt resolves a prompt request to a rendered prompt
func (s *Service) ResolvePrompt(ctx context.Context, orgID uuid.UUID, req *PromptRequest) (*PromptResponse, error) {
	// Determine version to use
	version := req.Version
	isCanary := false

	if version == nil {
		// Use deployment configuration
		deployment, err := s.getDeployment(ctx, orgID, req.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get deployment: %w", err)
		}

		// Canary routing
		if deployment.CanaryVersion != nil && deployment.CanaryRatio > 0 {
			if rand.Float64() < deployment.CanaryRatio {
				version = deployment.CanaryVersion
				isCanary = true
			} else {
				version = &deployment.StableVersion
			}
		} else {
			version = &deployment.StableVersion
		}
	}

	// Get prompt template
	prompt, err := s.GetPromptTemplate(ctx, orgID, req.Name, *version)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompt template: %w", err)
	}

	// Validate inputs against schema
	if err := s.validateInputs(req.Inputs, prompt.Schema); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Render template
	renderedText, err := s.renderer.Render(prompt.Template, req.Inputs)
	if err != nil {
		return nil, fmt.Errorf("template rendering failed: %w", err)
	}

	// Estimate token count
	tokenCount := s.estimateTokens(renderedText)

	return &PromptResponse{
		ID:           prompt.ID,
		Name:         prompt.Name,
		Version:      prompt.Version,
		RenderedText: renderedText,
		Metadata:     prompt.Metadata,
		TokenCount:   tokenCount,
		IsCanary:     isCanary,
	}, nil
}

// CreateSuite creates a new evaluation suite
func (s *Service) CreateSuite(ctx context.Context, orgID uuid.UUID, name string, cases []TestCase) (*PromptSuite, error) {
	suite := &PromptSuite{
		ID:        uuid.New(),
		OrgID:     orgID,
		Name:      name,
		Cases:     cases,
		CreatedAt: time.Now(),
	}

	if err := s.saveSuite(ctx, suite); err != nil {
		return nil, fmt.Errorf("failed to save suite: %w", err)
	}

	return suite, nil
}

// EvaluatePrompt runs an evaluation suite against a prompt version
func (s *Service) EvaluatePrompt(ctx context.Context, orgID uuid.UUID, req *EvaluateRequest) (*EvaluationRun, error) {
	// Get prompt template
	prompt, err := s.GetPromptTemplate(ctx, orgID, req.PromptName, req.PromptVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompt template: %w", err)
	}

	// Get evaluation suite
	suite, err := s.getSuite(ctx, orgID, req.SuiteName)
	if err != nil {
		return nil, fmt.Errorf("failed to get suite: %w", err)
	}

	// Run evaluation
	evalRun := &EvaluationRun{
		ID:        uuid.New(),
		PromptID:  prompt.ID,
		SuiteID:   suite.ID,
		Status:    EvaluationStatusQueued,
		StartedAt: time.Now(),
	}

	// Execute evaluation asynchronously
	go func() {
		s.executeEvaluation(context.Background(), evalRun, prompt, suite, req.Parallel)
	}()

	return evalRun, nil
}

// UpdateDeployment updates the deployment configuration for a prompt
func (s *Service) UpdateDeployment(ctx context.Context, orgID uuid.UUID, req *DeploymentRequest) (*PromptDeployment, error) {
	// Validate versions exist
	if _, err := s.GetPromptTemplate(ctx, orgID, req.PromptName, req.StableVersion); err != nil {
		return nil, fmt.Errorf("stable version %d not found: %w", req.StableVersion, err)
	}

	if req.CanaryVersion != nil {
		if _, err := s.GetPromptTemplate(ctx, orgID, req.PromptName, *req.CanaryVersion); err != nil {
			return nil, fmt.Errorf("canary version %d not found: %w", *req.CanaryVersion, err)
		}
	}

	deployment := &PromptDeployment{
		ID:            uuid.New(),
		OrgID:         orgID,
		PromptName:    req.PromptName,
		StableVersion: req.StableVersion,
		CanaryVersion: req.CanaryVersion,
		CanaryRatio:   req.CanaryRatio,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.saveDeployment(ctx, deployment); err != nil {
		return nil, fmt.Errorf("failed to save deployment: %w", err)
	}

	return deployment, nil
}

// Helper methods

func (s *Service) getNextVersion(ctx context.Context, orgID uuid.UUID, name string) (int, error) {
	query := `SELECT COALESCE(MAX(version), 0) + 1 FROM prompt_template WHERE org_id = $1 AND name = $2`
	
	var nextVersion int
	err := s.db.QueryRowContext(ctx, query, orgID, name).Scan(&nextVersion)
	if err != nil {
		return 0, err
	}

	return nextVersion, nil
}

func (s *Service) savePromptTemplate(ctx context.Context, prompt *PromptTemplate) error {
	schemaJSON, err := json.Marshal(prompt.Schema)
	if err != nil {
		return err
	}

	metadataJSON, err := json.Marshal(prompt.Metadata)
	if err != nil {
		return err
	}

	query := `INSERT INTO prompt_template (id, org_id, name, version, template, schema, metadata, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = s.db.ExecContext(ctx, query,
		prompt.ID, prompt.OrgID, prompt.Name, prompt.Version,
		prompt.Template, schemaJSON, metadataJSON, prompt.CreatedAt,
	)
	return err
}

func (s *Service) getDeployment(ctx context.Context, orgID uuid.UUID, promptName string) (*PromptDeployment, error) {
	query := `SELECT id, org_id, prompt_name, stable_version, canary_version, canary_ratio, created_at, updated_at
			  FROM prompt_deployment 
			  WHERE org_id = $1 AND prompt_name = $2
			  ORDER BY updated_at DESC LIMIT 1`

	var deployment PromptDeployment
	err := s.db.QueryRowContext(ctx, query, orgID, promptName).Scan(
		&deployment.ID, &deployment.OrgID, &deployment.PromptName,
		&deployment.StableVersion, &deployment.CanaryVersion, &deployment.CanaryRatio,
		&deployment.CreatedAt, &deployment.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &deployment, nil
}

func (s *Service) saveSuite(ctx context.Context, suite *PromptSuite) error {
	casesJSON, err := json.Marshal(suite.Cases)
	if err != nil {
		return err
	}

	query := `INSERT INTO prompt_suite (id, org_id, name, cases, created_at)
			  VALUES ($1, $2, $3, $4, $5)`

	_, err = s.db.ExecContext(ctx, query,
		suite.ID, suite.OrgID, suite.Name, casesJSON, suite.CreatedAt,
	)
	return err
}

func (s *Service) getSuite(ctx context.Context, orgID uuid.UUID, name string) (*PromptSuite, error) {
	query := `SELECT id, org_id, name, cases, created_at FROM prompt_suite WHERE org_id = $1 AND name = $2`

	var suite PromptSuite
	var casesJSON []byte

	err := s.db.QueryRowContext(ctx, query, orgID, name).Scan(
		&suite.ID, &suite.OrgID, &suite.Name, &casesJSON, &suite.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(casesJSON, &suite.Cases); err != nil {
		return nil, err
	}

	return &suite, nil
}

func (s *Service) saveDeployment(ctx context.Context, deployment *PromptDeployment) error {
	query := `INSERT INTO prompt_deployment (id, org_id, prompt_name, stable_version, canary_version, canary_ratio, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			  ON CONFLICT (org_id, prompt_name) 
			  DO UPDATE SET stable_version = $4, canary_version = $5, canary_ratio = $6, updated_at = $8`

	_, err := s.db.ExecContext(ctx, query,
		deployment.ID, deployment.OrgID, deployment.PromptName,
		deployment.StableVersion, deployment.CanaryVersion, deployment.CanaryRatio,
		deployment.CreatedAt, deployment.UpdatedAt,
	)
	return err
}

func (s *Service) validateInputs(inputs map[string]interface{}, schema Schema) error {
	// Basic schema validation - in production would use a proper JSON schema validator
	for _, required := range schema.Required {
		if _, exists := inputs[required]; !exists {
			return fmt.Errorf("required field %s is missing", required)
		}
	}
	return nil
}

func (s *Service) estimateTokens(text string) int {
	// Simple token estimation - in production would use proper tokenizer
	return len(text) / 4 // Rough approximation
}

func (s *Service) executeEvaluation(ctx context.Context, evalRun *EvaluationRun, prompt *PromptTemplate, suite *PromptSuite, parallel int) {
	// This would be implemented to actually run the evaluation
	// For now, just mark as completed
	evalRun.Status = EvaluationStatusCompleted
	now := time.Now()
	evalRun.CompletedAt = &now
}