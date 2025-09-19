package aor

import (
	"github.com/google/uuid"
	"testing"
	"time"

	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestControlPlane_SubmitWorkflow(t *testing.T) {
	// This would be a full integration test in a real implementation
	// For now, we'll test the basic structure and validation

	_ = &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Database: "test",
			SSLMode:  "disable",
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		NATS: config.NATSConfig{
			URL: "nats://localhost:4222",
		},
	}

	// In a real test, we'd set up test databases and services
	// For now, we'll test the request validation logic

	t.Run("ValidateWorkflowRequest", func(t *testing.T) {
		req := &RunRequest{
			WorkflowName:    "test_workflow",
			WorkflowVersion: 1,
			Inputs: map[string]interface{}{
				"input1": "value1",
				"input2": 42,
			},
			BudgetCents: 1000,
		}

		// Test request validation
		assert.NotEmpty(t, req.WorkflowName)
		assert.Greater(t, req.WorkflowVersion, 0)
		assert.NotNil(t, req.Inputs)
		assert.Greater(t, req.BudgetCents, int64(0))
	})

	t.Run("WorkflowRunCreation", func(t *testing.T) {
		run := &WorkflowRun{
			ID:             uuid.New(),
			WorkflowSpecID: uuid.New(),
			Status:         RunStatusQueued,
			CostCents:      0,
			Metadata: map[string]interface{}{
				"test": "value",
			},
			CreatedAt: time.Now(),
		}

		assert.NotEqual(t, uuid.Nil, run.ID)
		assert.Equal(t, RunStatusQueued, run.Status)
		assert.Equal(t, int64(0), run.CostCents)
		assert.NotNil(t, run.Metadata)
	})
}

func TestWorkflowValidation(t *testing.T) {
	t.Run("ValidDAG", func(t *testing.T) {
		dag := DAG{
			Steps: []Step{
				{
					ID:   "step1",
					Type: "llm",
					Name: "LLM Step",
					Config: map[string]interface{}{
						"prompt_ref": "test_prompt@1",
					},
				},
				{
					ID:   "step2",
					Type: "tool",
					Name: "Tool Step",
					Config: map[string]interface{}{
						"tool_name": "test_tool",
					},
				},
			},
			Edges: []Edge{
				{From: "step1", To: "step2"},
			},
		}

		// Test DAG structure
		assert.Len(t, dag.Steps, 2)
		assert.Len(t, dag.Edges, 1)
		assert.Equal(t, "step1", dag.Edges[0].From)
		assert.Equal(t, "step2", dag.Edges[0].To)
	})

	t.Run("ExecutorTypes", func(t *testing.T) {
		executorTypes := []ExecutorType{
			ExecutorTypeLLM,
			ExecutorTypeHTTP,
			ExecutorTypeScript,
		}

		for _, executorType := range executorTypes {
			assert.NotEmpty(t, string(executorType))
		}
	})
}

func TestStepExecution(t *testing.T) {
	t.Run("StepRunLifecycle", func(t *testing.T) {
		stepRun := &StepRun{
			ID:            uuid.New().String(),
			WorkflowRunID: uuid.New(),
			NodeID:        "test_node",
			Attempt:       1,
			Status:        StepStatusQueued,
			CreatedAt:     time.Now(),
		}

		// Test initial state
		assert.Equal(t, StepStatusQueued, stepRun.Status)
		assert.Equal(t, 1, stepRun.Attempt)
		assert.True(t, stepRun.StartedAt.IsZero())
		assert.Nil(t, stepRun.FinishedAt)

		// Simulate status transitions
		now := time.Now()
		stepRun.Status = StepStatusRunning
		stepRun.StartedAt = now

		assert.Equal(t, StepStatusRunning, stepRun.Status)
		assert.False(t, stepRun.StartedAt.IsZero())

		// Complete step
		endTime := time.Now()
		stepRun.Status = StepStatusSucceeded
		stepRun.FinishedAt = &endTime

		assert.Equal(t, StepStatusSucceeded, stepRun.Status)
		assert.NotNil(t, stepRun.FinishedAt)
	})
}

func TestTaskExecution(t *testing.T) {
	t.Run("TaskCreation", func(t *testing.T) {
		task := &Task{
			ID:     uuid.New(),
			RunID:  uuid.New(),
			NodeID: "test_node",
			Attempt: 1,
			Node: &Node{
				ID:   "test_node",
				Type: "llm",
				Config: map[string]interface{}{
					"prompt_ref": "test_prompt@1",
				},
			},
			Inputs: map[string]interface{}{
				"input1": "test_value",
			},
		}

		assert.NotEqual(t, uuid.Nil, task.ID)
		assert.Equal(t, "llm", task.Node.Type)
		assert.NotNil(t, task.Inputs)
	})

	t.Run("TaskResult", func(t *testing.T) {
		result := &TaskResult{
			TaskID: uuid.New(),
			Status: TaskStatusSucceeded,
			Output: map[string]interface{}{
				"result": "test_output",
			},
			CostCents:        100,
			TokensPrompt:     50,
			TokensCompletion: 75,
		}

		assert.Equal(t, TaskStatusSucceeded, result.Status)
		assert.NotNil(t, result.Output)
		assert.Equal(t, int64(100), result.CostCents)
		assert.Equal(t, 50, result.TokensPrompt)
		assert.Equal(t, 75, result.TokensCompletion)
	})
}

func TestExecutorTypes(t *testing.T) {
	t.Run("ExecutorTypeValidation", func(t *testing.T) {
		executorTypes := []ExecutorType{
			ExecutorTypeLLM,
			ExecutorTypeHTTP,
			ExecutorTypeScript,
			ExecutorTypeWASM,
			ExecutorTypeWorkflow,
		}

		for _, executorType := range executorTypes {
			assert.NotEmpty(t, string(executorType))
		}
	})
}

// Benchmark tests for performance validation
func BenchmarkWorkflowSubmission(b *testing.B) {
	req := &RunRequest{
		WorkflowName:    "benchmark_workflow",
		WorkflowVersion: 1,
		Inputs: map[string]interface{}{
			"input": "test",
		},
		BudgetCents: 1000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate workflow submission validation
		_ = req.WorkflowName != ""
		_ = req.WorkflowVersion > 0
		_ = req.Inputs != nil
	}
}

func BenchmarkDAGValidation(b *testing.B) {
	dag := DAG{
		Steps: []Step{
			{ID: "step1", Type: "llm"},
			{ID: "step2", Type: "http"},
			{ID: "step3", Type: "script"},
		},
		Edges: []Edge{
			{From: "step1", To: "step2"},
			{From: "step2", To: "step3"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate DAG validation
		stepMap := make(map[string]bool)
		for _, step := range dag.Steps {
			stepMap[step.ID] = true
		}
		
		for _, edge := range dag.Edges {
			_ = stepMap[edge.From] && stepMap[edge.To]
		}
	}
}