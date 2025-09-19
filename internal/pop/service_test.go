package pop

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptTemplateValidation(t *testing.T) {
	t.Run("ValidTemplate", func(t *testing.T) {
		template := &PromptTemplate{
			ID:      uuid.New(),
			OrgID:   uuid.New(),
			Name:    "test_prompt",
			Version: 1,
			Template: "Hello {{name}}, please analyze: {{content}}",
			Schema: Schema{
				Type: "object",
				Properties: map[string]Property{
					"name": {
						Type:        "string",
						Description: "User name",
					},
					"content": {
						Type:        "string",
						Description: "Content to analyze",
					},
				},
				Required: []string{"name", "content"},
			},
			Metadata: Metadata{
				"author": "test-team",
				"tags":   []string{"analysis", "greeting"},
			},
			CreatedAt: time.Now(),
		}

		assert.NotEqual(t, uuid.Nil, template.ID)
		assert.NotEmpty(t, template.Name)
		assert.Greater(t, template.Version, 0)
		assert.Contains(t, template.Template, "{{name}}")
		assert.Contains(t, template.Template, "{{content}}")
		assert.Len(t, template.Schema.Required, 2)
	})

	t.Run("SchemaValidation", func(t *testing.T) {
		schema := Schema{
			Type: "object",
			Properties: map[string]Property{
				"input": {
					Type:        "string",
					Description: "Input text",
				},
				"options": {
					Type:    "string",
					Enum:    []string{"option1", "option2", "option3"},
					Default: "option1",
				},
			},
			Required: []string{"input"},
		}

		assert.Equal(t, "object", schema.Type)
		assert.Len(t, schema.Properties, 2)
		assert.Contains(t, schema.Required, "input")
		assert.Len(t, schema.Properties["options"].Enum, 3)
	})
}

func TestPromptRendering(t *testing.T) {
	renderer := NewTemplateRenderer()

	t.Run("BasicRendering", func(t *testing.T) {
		template := "Hello {{.name}}, your score is {{.score}}"
		data := map[string]interface{}{
			"name":  "Alice",
			"score": 95,
		}

		result, err := renderer.Render(template, data)
		require.NoError(t, err)
		assert.Equal(t, "Hello Alice, your score is 95", result)
	})

	t.Run("TemplateWithFunctions", func(t *testing.T) {
		template := "Hello {{upper .name}}, your message: {{trim .message}}"
		data := map[string]interface{}{
			"name":    "alice",
			"message": "  Hello World  ",
		}

		result, err := renderer.Render(template, data)
		require.NoError(t, err)
		assert.Equal(t, "Hello ALICE, your message: Hello World", result)
	})

	t.Run("TemplateValidation", func(t *testing.T) {
		template := "Hello {{.name}}, your score is {{.score}}"
		schema := Schema{
			Type: "object",
			Properties: map[string]Property{
				"name":  {Type: "string"},
				"score": {Type: "number"},
			},
			Required: []string{"name", "score"},
		}

		err := renderer.Validate(template, schema)
		assert.NoError(t, err)
	})

	t.Run("InvalidTemplate", func(t *testing.T) {
		template := "Hello {{name}, missing closing brace"
		schema := Schema{Type: "object"}

		err := renderer.Validate(template, schema)
		assert.Error(t, err)
	})

	t.Run("SafetyValidation", func(t *testing.T) {
		dangerousTemplate := "{{.}} dangerous direct access"
		schema := Schema{Type: "object"}

		err := renderer.Validate(dangerousTemplate, schema)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsafe pattern")
	})
}

func TestPromptDeployment(t *testing.T) {
	t.Run("DeploymentConfiguration", func(t *testing.T) {
		deployment := &PromptDeployment{
			ID:            uuid.New(),
			OrgID:         uuid.New(),
			PromptName:    "test_prompt",
			StableVersion: 2,
			CanaryVersion: &[]int{3}[0],
			CanaryRatio:   0.1,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		assert.NotEqual(t, uuid.Nil, deployment.ID)
		assert.Equal(t, "test_prompt", deployment.PromptName)
		assert.Equal(t, 2, deployment.StableVersion)
		assert.NotNil(t, deployment.CanaryVersion)
		assert.Equal(t, 3, *deployment.CanaryVersion)
		assert.Equal(t, 0.1, deployment.CanaryRatio)
	})

	t.Run("CanaryRouting", func(t *testing.T) {
		// Test canary routing logic
		deployment := &PromptDeployment{
			StableVersion: 1,
			CanaryVersion: &[]int{2}[0],
			CanaryRatio:   0.2, // 20% canary traffic
		}

		// Simulate routing decisions
		canaryCount := 0
		stableCount := 0
		totalRequests := 1000

		for i := 0; i < totalRequests; i++ {
			// Mock random number generation for testing
			mockRandom := float64(i) / float64(totalRequests)
			if mockRandom < deployment.CanaryRatio {
				canaryCount++
			} else {
				stableCount++
			}
		}

		expectedCanary := int(float64(totalRequests) * deployment.CanaryRatio)
		expectedStable := totalRequests - expectedCanary

		assert.Equal(t, expectedCanary, canaryCount)
		assert.Equal(t, expectedStable, stableCount)
	})
}

func TestEvaluationSuite(t *testing.T) {
	t.Run("TestCaseCreation", func(t *testing.T) {
		testCase := TestCase{
			ID: "test_case_1",
			Input: map[string]interface{}{
				"text": "Sample input text",
			},
			Expected: Expected{
				Contains: []string{"analysis", "summary"},
				Schema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"summary": map[string]interface{}{"type": "string"},
					},
				},
			},
			Scoring: ScoringConfig{
				Type:   ScoringContains,
				Weight: 1.0,
			},
		}

		assert.Equal(t, "test_case_1", testCase.ID)
		assert.NotNil(t, testCase.Input)
		assert.Len(t, testCase.Expected.Contains, 2)
		assert.Equal(t, ScoringContains, testCase.Scoring.Type)
	})

	t.Run("EvaluationSuite", func(t *testing.T) {
		suite := &PromptSuite{
			ID:    uuid.New(),
			OrgID: uuid.New(),
			Name:  "test_suite",
			Cases: []TestCase{
				{
					ID: "case1",
					Input: map[string]interface{}{
						"text": "Test input 1",
					},
					Expected: Expected{
						Output: "Expected output 1",
					},
					Scoring: ScoringConfig{
						Type: ScoringExact,
					},
				},
				{
					ID: "case2",
					Input: map[string]interface{}{
						"text": "Test input 2",
					},
					Expected: Expected{
						Contains: []string{"keyword1", "keyword2"},
					},
					Scoring: ScoringConfig{
						Type: ScoringContains,
					},
				},
			},
			CreatedAt: time.Now(),
		}

		assert.Len(t, suite.Cases, 2)
		assert.Equal(t, "test_suite", suite.Name)
		assert.Equal(t, ScoringExact, suite.Cases[0].Scoring.Type)
		assert.Equal(t, ScoringContains, suite.Cases[1].Scoring.Type)
	})
}

func TestEvaluator(t *testing.T) {
	evaluator := NewEvaluator(nil) // Mock database for testing

	t.Run("ExactScoring", func(t *testing.T) {
		actual := "Hello World"
		expected := "Hello World"

		score, passed, err := evaluator.scoreExact(actual, expected)
		require.NoError(t, err)
		assert.Equal(t, 1.0, score)
		assert.True(t, passed)

		// Test mismatch
		score, passed, err = evaluator.scoreExact(actual, "Different")
		require.NoError(t, err)
		assert.Equal(t, 0.0, score)
		assert.False(t, passed)
	})

	t.Run("ContainsScoring", func(t *testing.T) {
		actual := "This is a test document with important keywords"
		contains := []string{"test", "important", "keywords"}

		score, passed, err := evaluator.scoreContains(actual, contains)
		require.NoError(t, err)
		assert.Equal(t, 1.0, score)
		assert.True(t, passed)

		// Test partial match
		partialContains := []string{"test", "missing", "keywords"}
		score, passed, err = evaluator.scoreContains(actual, partialContains)
		require.NoError(t, err)
		assert.Equal(t, 2.0/3.0, score)
		assert.False(t, passed)
	})

	t.Run("RegexScoring", func(t *testing.T) {
		actual := "Email: user@example.com"
		config := map[string]interface{}{
			"pattern": `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`,
		}

		score, passed, err := evaluator.scoreRegex(actual, config)
		require.NoError(t, err)
		assert.Equal(t, 1.0, score)
		assert.True(t, passed)

		// Test no match
		noMatch := "No email here"
		score, passed, err = evaluator.scoreRegex(noMatch, config)
		require.NoError(t, err)
		assert.Equal(t, 0.0, score)
		assert.False(t, passed)
	})

	t.Run("EmbeddingSimilarity", func(t *testing.T) {
		actual := "The quick brown fox jumps over the lazy dog"
		expected := "A fast brown fox leaps over a sleepy dog"
		config := map[string]interface{}{
			"threshold": 0.7,
		}

		score, _, err := evaluator.scoreEmbedding(actual, expected, config)
		require.NoError(t, err)
		assert.Greater(t, score, 0.0)
		assert.LessOrEqual(t, score, 1.0)
		// The mock implementation should return a reasonable similarity
	})
}

func TestPromptRequest(t *testing.T) {
	t.Run("PromptResolution", func(t *testing.T) {
		request := &PromptRequest{
			Name:    "test_prompt",
			Version: &[]int{1}[0],
			Inputs: map[string]interface{}{
				"name":    "Alice",
				"content": "Test content",
			},
			Context: map[string]interface{}{
				"user_id": "user123",
			},
		}

		assert.Equal(t, "test_prompt", request.Name)
		assert.NotNil(t, request.Version)
		assert.Equal(t, 1, *request.Version)
		assert.NotNil(t, request.Inputs)
		assert.NotNil(t, request.Context)
	})

	t.Run("PromptResponse", func(t *testing.T) {
		response := &PromptResponse{
			ID:           uuid.New(),
			Name:         "test_prompt",
			Version:      1,
			RenderedText: "Hello Alice, please analyze: Test content",
			Metadata: map[string]interface{}{
				"template_id": "template123",
			},
			TokenCount: 15,
			IsCanary:   false,
		}

		assert.NotEqual(t, uuid.Nil, response.ID)
		assert.Equal(t, "test_prompt", response.Name)
		assert.Equal(t, 1, response.Version)
		assert.Contains(t, response.RenderedText, "Alice")
		assert.Equal(t, 15, response.TokenCount)
		assert.False(t, response.IsCanary)
	})
}

// Benchmark tests
func BenchmarkTemplateRendering(b *testing.B) {
	renderer := NewTemplateRenderer()
	template := "Hello {{name}}, your score is {{score}} and status is {{upper status}}"
	data := map[string]interface{}{
		"name":   "Alice",
		"score":  95,
		"status": "active",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := renderer.Render(template, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEvaluationScoring(b *testing.B) {
	evaluator := NewEvaluator(nil)
	actual := "This is a test document with important keywords and analysis"
	contains := []string{"test", "important", "keywords", "analysis"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := evaluator.scoreContains(actual, contains)
		if err != nil {
			b.Fatal(err)
		}
	}
}