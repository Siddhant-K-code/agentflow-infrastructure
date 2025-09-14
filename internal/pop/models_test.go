package pop

import (
	"testing"

	"github.com/google/uuid"
)

func TestPromptTemplate(t *testing.T) {
	template := PromptTemplate{
		ID:       uuid.New(),
		OrgID:    uuid.New(),
		Name:     "test_prompt",
		Version:  1,
		Template: "Hello {{name}}!",
	}

	if template.Name != "test_prompt" {
		t.Errorf("Expected name 'test_prompt', got %s", template.Name)
	}

	if template.Version != 1 {
		t.Errorf("Expected version 1, got %d", template.Version)
	}

	if template.Template != "Hello {{name}}!" {
		t.Errorf("Expected template 'Hello {{name}}!', got %s", template.Template)
	}
}

func TestPromptReference(t *testing.T) {
	ref := PromptReference{
		Name:    "test_prompt",
		Version: nil, // latest
	}

	if ref.Name != "test_prompt" {
		t.Errorf("Expected name 'test_prompt', got %s", ref.Name)
	}

	if ref.Version != nil {
		t.Errorf("Expected version to be nil (latest), got %v", ref.Version)
	}
}

func TestEvaluationCase(t *testing.T) {
	testCase := EvaluationCase{
		Input: map[string]interface{}{
			"name": "World",
		},
		Expected: "Hello World!",
		Scoring: ScoringConfig{
			Type:      ScoringExact,
			Threshold: 1.0,
		},
	}

	if testCase.Input["name"] != "World" {
		t.Errorf("Expected input name 'World', got %v", testCase.Input["name"])
	}

	if testCase.Expected != "Hello World!" {
		t.Errorf("Expected output 'Hello World!', got %v", testCase.Expected)
	}

	if testCase.Scoring.Type != ScoringExact {
		t.Errorf("Expected scoring type 'exact', got %s", testCase.Scoring.Type)
	}
}

func TestPromptDeployment(t *testing.T) {
	deployment := PromptDeployment{
		ID:            uuid.New(),
		OrgID:         uuid.New(),
		PromptName:    "test_prompt",
		StableVersion: 2,
		CanaryVersion: func() *int { v := 3; return &v }(),
		CanaryRatio:   0.1,
	}

	if deployment.PromptName != "test_prompt" {
		t.Errorf("Expected prompt name 'test_prompt', got %s", deployment.PromptName)
	}

	if deployment.StableVersion != 2 {
		t.Errorf("Expected stable version 2, got %d", deployment.StableVersion)
	}

	if deployment.CanaryVersion == nil || *deployment.CanaryVersion != 3 {
		t.Errorf("Expected canary version 3, got %v", deployment.CanaryVersion)
	}

	if deployment.CanaryRatio != 0.1 {
		t.Errorf("Expected canary ratio 0.1, got %f", deployment.CanaryRatio)
	}
}