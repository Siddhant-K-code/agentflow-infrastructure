package aor

import (
	"testing"

	"github.com/google/uuid"
	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/common"
)

func TestWorkflowSpec(t *testing.T) {
	spec := WorkflowSpec{
		ID:      uuid.New(),
		OrgID:   uuid.New(),
		Name:    "test_workflow",
		Version: 1,
		DAG:     []byte(`{"nodes": [], "edges": []}`),
	}

	if spec.Name != "test_workflow" {
		t.Errorf("Expected name 'test_workflow', got %s", spec.Name)
	}

	if spec.Version != 1 {
		t.Errorf("Expected version 1, got %d", spec.Version)
	}
}

func TestWorkflowRun(t *testing.T) {
	run := WorkflowRun{
		ID:             uuid.New(),
		WorkflowSpecID: uuid.New(),
		Status:         common.StatusQueued,
		CostCents:      100,
	}

	if run.Status != common.StatusQueued {
		t.Errorf("Expected status 'queued', got %s", run.Status)
	}

	if run.CostCents != 100 {
		t.Errorf("Expected cost 100, got %d", run.CostCents)
	}
}

func TestStepRun(t *testing.T) {
	step := StepRun{
		ID:               uuid.New(),
		WorkflowRunID:    uuid.New(),
		NodeID:           "test_node",
		Status:           common.StatusRunning,
		Attempt:          1,
		CostCents:        50,
		TokensPrompt:     100,
		TokensCompletion: 200,
	}

	if step.NodeID != "test_node" {
		t.Errorf("Expected node ID 'test_node', got %s", step.NodeID)
	}

	if step.Status != common.StatusRunning {
		t.Errorf("Expected status 'running', got %s", step.Status)
	}

	if step.TokensPrompt != 100 {
		t.Errorf("Expected prompt tokens 100, got %d", step.TokensPrompt)
	}

	if step.TokensCompletion != 200 {
		t.Errorf("Expected completion tokens 200, got %d", step.TokensCompletion)
	}
}

func TestDAGSpec(t *testing.T) {
	dag := DAGSpec{
		Nodes: []NodeSpec{
			{
				ID:   "node1",
				Type: NodeTypeLLM,
				Config: map[string]interface{}{
					"prompt": "test_prompt",
				},
			},
		},
		Edges: []EdgeSpec{
			{
				From: "node1",
				To:   "node2",
			},
		},
	}

	if len(dag.Nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(dag.Nodes))
	}

	if dag.Nodes[0].Type != NodeTypeLLM {
		t.Errorf("Expected node type 'llm', got %s", dag.Nodes[0].Type)
	}

	if len(dag.Edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(dag.Edges))
	}
}