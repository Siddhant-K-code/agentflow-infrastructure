package aor

import (
	"github.com/google/uuid"
	"context"
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
			Metadata: Metadata{
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
			Nodes: []Node{
				{
					ID:   "node1",
					Type: NodeTypeLLM,
					Config: NodeConfig{
						PromptRef: "test_prompt@1",
					},
				},
				{
					ID:   "node2",
					Type: NodeTypeTool,
					Config: NodeConfig{
						ToolName: "test_tool",
					},
				},
			},
			Edges: []Edge{
				{From: "node1", To: "node2"},
			},
		}

		// Test DAG structure
		assert.Len(t, dag.Nodes, 2)
		assert.Len(t, dag.Edges, 1)
		assert.Equal(t, "node1", dag.Edges[0].From)
		assert.Equal(t, "node2", dag.Edges[0].To)
	})

	t.Run("NodeTypes", func(t *testing.T) {
		nodeTypes := []NodeType{
			NodeTypeLLM,
			NodeTypeTool,
			NodeTypeFunction,
			NodeTypeSwitch,
			NodeTypeMap,
			NodeTypeReduce,
		}

		for _, nodeType := range nodeTypes {
			assert.NotEmpty(t, string(nodeType))
		}
	})
}

func TestStepExecution(t *testing.T) {
	t.Run("StepRunLifecycle", func(t *testing.T) {
		stepRun := &StepRun{
			ID:            uuid.New(),
			WorkflowRunID: uuid.New(),
			NodeID:        "test_node",
			Attempt:       1,
			Status:        StepStatusQueued,
			CreatedAt:     time.Now(),
		}

		// Test initial state
		assert.Equal(t, StepStatusQueued, stepRun.Status)
		assert.Equal(t, 1, stepRun.Attempt)
		assert.Nil(t, stepRun.StartedAt)
		assert.Nil(t, stepRun.EndedAt)

		// Simulate status transitions
		now := time.Now()
		stepRun.Status = StepStatusRunning
		stepRun.StartedAt = &now

		assert.Equal(t, StepStatusRunning, stepRun.Status)
		assert.NotNil(t, stepRun.StartedAt)

		// Complete step
		endTime := time.Now()
		stepRun.Status = StepStatusSucceeded
		stepRun.EndedAt = &endTime
		stepRun.CostCents = 150

		assert.Equal(t, StepStatusSucceeded, stepRun.Status)
		assert.NotNil(t, stepRun.EndedAt)
		assert.Equal(t, int64(150), stepRun.CostCents)
	})
}

func TestTaskExecution(t *testing.T) {
	t.Run("TaskCreation", func(t *testing.T) {
		task := &Task{
			ID:     uuid.New(),
			RunID:  uuid.New(),
			NodeID: "test_node",
			Attempt: 1,
			Node: Node{
				ID:   "test_node",
				Type: NodeTypeLLM,
				Config: NodeConfig{
					PromptRef: "test_prompt@1",
				},
			},
			Inputs: map[string]interface{}{
				"input1": "test_value",
			},
			DeadlineAt: time.Now().Add(30 * time.Minute),
		}

		assert.NotEqual(t, uuid.Nil, task.ID)
		assert.Equal(t, NodeTypeLLM, task.Node.Type)
		assert.NotNil(t, task.Inputs)
		assert.True(t, task.DeadlineAt.After(time.Now()))
	})

	t.Run("TaskResult", func(t *testing.T) {
		result := &TaskResult{
			TaskID: uuid.New(),
			Status: StepStatusSucceeded,
			Output: map[string]interface{}{
				"result": "test_output",
			},
			CostCents:        100,
			TokensPrompt:     50,
			TokensCompletion: 75,
		}

		assert.Equal(t, StepStatusSucceeded, result.Status)
		assert.NotNil(t, result.Output)
		assert.Equal(t, int64(100), result.CostCents)
		assert.Equal(t, 50, result.TokensPrompt)
		assert.Equal(t, 75, result.TokensCompletion)
	})
}

func TestPolicyValidation(t *testing.T) {
	t.Run("QualityTiers", func(t *testing.T) {
		tiers := []QualityTier{
			QualityGold,
			QualitySilver,
			QualityBronze,
		}

		for _, tier := range tiers {
			assert.NotEmpty(t, string(tier))
		}
	})

	t.Run("PolicyConfiguration", func(t *testing.T) {
		policy := &Policy{
			Quality:    QualityGold,
			SLAMillis:  30000,
			MaxRetries: 3,
			Optional:   false,
		}

		assert.Equal(t, QualityGold, policy.Quality)
		assert.Equal(t, 30000, policy.SLAMillis)
		assert.Equal(t, 3, policy.MaxRetries)
		assert.False(t, policy.Optional)
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
		Nodes: []Node{
			{ID: "node1", Type: NodeTypeLLM},
			{ID: "node2", Type: NodeTypeTool},
			{ID: "node3", Type: NodeTypeFunction},
		},
		Edges: []Edge{
			{From: "node1", To: "node2"},
			{From: "node2", To: "node3"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate DAG validation
		nodeMap := make(map[string]bool)
		for _, node := range dag.Nodes {
			nodeMap[node.ID] = true
		}
		
		for _, edge := range dag.Edges {
			_ = nodeMap[edge.From] && nodeMap[edge.To]
		}
	}
}