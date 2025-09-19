package aor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"time"

	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/db"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

type Scheduler struct {
	db    *db.PostgresDB
	redis *redis.Client
	nats  *nats.Conn
	js    nats.JetStreamContext
}

func NewScheduler(pgDB *db.PostgresDB, redisClient *redis.Client, natsConn *nats.Conn, js nats.JetStreamContext) *Scheduler {
	return &Scheduler{
		db:    pgDB,
		redis: redisClient,
		nats:  natsConn,
		js:    js,
	}
}

func (s *Scheduler) ScheduleWorkflow(ctx context.Context, run *WorkflowRun) error {
	log.Printf("Scheduling workflow run %s", run.ID)

	// Get workflow spec
	spec, err := s.getWorkflowSpec(ctx, run.WorkflowSpecID)
	if err != nil {
		return fmt.Errorf("failed to get workflow spec: %w", err)
	}

	// Create step runs for initial nodes
	for _, step := range spec.DAG.Steps {
		// Check if this step has dependencies
		if s.hasDependencies(step.ID, spec.DAG.Edges) {
			continue // Skip steps with dependencies for now
		}

		stepRun := &StepRun{
			ID:            uuid.New().String(),
			WorkflowRunID: run.ID,
			NodeID:        step.ID,
			StepID:        step.ID,
			Attempt:       1,
			Status:        StepStatusQueued,
			CreatedAt:     time.Now(),
		}

		if err := s.saveStepRun(ctx, stepRun); err != nil {
			return fmt.Errorf("failed to save step run: %w", err)
		}

		// Create and enqueue task
		taskID, _ := uuid.Parse(stepRun.ID)
		if taskID == uuid.Nil {
			taskID = uuid.New()
		}

		node := &Node{
			ID:     step.ID,
			Type:   step.Type,
			Config: step.Config,
		}

		task := &Task{
			ID:         taskID,
			RunID:      run.ID,
			StepID:     step.ID,
			NodeID:     step.ID,
			Type:       step.Type,
			Attempt:    1,
			Node:       node,
			Inputs:     s.resolveInputs(ctx, run, node),
			Priority:   1,
			CreatedAt:  time.Now(),
			DeadlineAt: &[]time.Time{time.Now().Add(30 * time.Minute)}[0],
		}

		if err := s.enqueueTask(ctx, task); err != nil {
			return fmt.Errorf("failed to enqueue task: %w", err)
		}
	}

	return nil
}

func (s *Scheduler) ProcessTaskResult(ctx context.Context, result *TaskResult) error {
	log.Printf("Processing task result for task %s", result.TaskID)

	// Update step run
	if err := s.updateStepRun(ctx, result); err != nil {
		return fmt.Errorf("failed to update step run: %w", err)
	}

	// Check if workflow is complete
	if err := s.checkWorkflowCompletion(ctx, result); err != nil {
		return fmt.Errorf("failed to check workflow completion: %w", err)
	}

	return nil
}

func (s *Scheduler) hasDependencies(stepID string, edges []Edge) bool {
	for _, edge := range edges {
		if edge.To == stepID {
			return true
		}
	}
	return false
}

func (s *Scheduler) getWorkflowSpec(ctx context.Context, specID uuid.UUID) (*WorkflowSpec, error) {
	// Mock implementation
	return &WorkflowSpec{
		ID:   specID,
		Name: "mock-workflow",
		DAG: DAG{
			Steps: []Step{
				{
					ID:   "step1",
					Type: "llm",
					Name: "Mock Step",
					Config: map[string]interface{}{
						"prompt_ref": "mock-prompt",
					},
				},
			},
			Edges: []Edge{},
		},
	}, nil
}

func (s *Scheduler) saveStepRun(ctx context.Context, stepRun *StepRun) error {
	log.Printf("Saving step run %s", stepRun.ID)
	// Mock implementation
	return nil
}

func (s *Scheduler) resolveInputs(ctx context.Context, run *WorkflowRun, node *Node) map[string]interface{} {
	// Mock implementation
	return map[string]interface{}{
		"input": "mock input data",
	}
}

func (s *Scheduler) enqueueTask(ctx context.Context, task *Task) error {
	log.Printf("Enqueuing task %s", task.ID)

	// Serialize task
	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// Publish to NATS
	if _, err := s.js.Publish("agentflow.tasks", taskData); err != nil {
		return fmt.Errorf("failed to publish task: %w", err)
	}

	return nil
}

func (s *Scheduler) updateStepRun(ctx context.Context, result *TaskResult) error {
	log.Printf("Updating step run for task %s", result.TaskID)
	// Mock implementation
	return nil
}

func (s *Scheduler) checkWorkflowCompletion(ctx context.Context, result *TaskResult) error {
	log.Printf("Checking workflow completion for task %s", result.TaskID)
	// Mock implementation
	return nil
}

func (s *Scheduler) RetryFailedStep(ctx context.Context, stepRunID string) error {
	log.Printf("Retrying failed step %s", stepRunID)

	// Get step run
	stepRun, err := s.getStepRun(ctx, stepRunID)
	if err != nil {
		return fmt.Errorf("failed to get step run: %w", err)
	}

	// Create retry step run
	retryStepRun := &StepRun{
		ID:            uuid.New().String(),
		WorkflowRunID: stepRun.WorkflowRunID,
		NodeID:        stepRun.NodeID,
		StepID:        stepRun.StepID,
		Attempt:       stepRun.Attempt + 1,
		Status:        StepStatusQueued,
		CreatedAt:     time.Now(),
	}

	if err := s.saveStepRun(ctx, retryStepRun); err != nil {
		return fmt.Errorf("failed to save retry step run: %w", err)
	}

	// Create and enqueue retry task
	taskID, _ := uuid.Parse(retryStepRun.ID)
	if taskID == uuid.Nil {
		taskID = uuid.New()
	}

	node := &Node{
		ID:     stepRun.NodeID,
		Type:   "llm", // Mock type
		Config: map[string]interface{}{},
	}

	task := &Task{
		ID:         taskID,
		RunID:      stepRun.WorkflowRunID,
		StepID:     stepRun.StepID,
		NodeID:     stepRun.NodeID,
		Type:       "llm",
		Attempt:    retryStepRun.Attempt,
		Node:       node,
		Inputs:     map[string]interface{}{},
		Priority:   2, // Higher priority for retries
		CreatedAt:  time.Now(),
		DeadlineAt: &[]time.Time{time.Now().Add(30 * time.Minute)}[0],
	}

	if err := s.enqueueTask(ctx, task); err != nil {
		return fmt.Errorf("failed to enqueue retry task: %w", err)
	}

	return nil
}

func (s *Scheduler) getStepRun(ctx context.Context, stepRunID string) (*StepRun, error) {
	// Mock implementation
	runID := uuid.New()
	return &StepRun{
		ID:            stepRunID,
		WorkflowRunID: runID,
		NodeID:        "mock-node",
		StepID:        "mock-step",
		Attempt:       1,
		Status:        StepStatusFailed,
		CreatedAt:     time.Now(),
	}, nil
}

func (s *Scheduler) GetWorkflowStatus(ctx context.Context, runID uuid.UUID) (*WorkflowRun, error) {
	log.Printf("Getting workflow status for run %s", runID)

	// Mock implementation
	return &WorkflowRun{
		ID:             runID,
		WorkflowSpecID: uuid.New(),
		Status:         WorkflowStatusRunning,
		CreatedAt:      time.Now(),
		Steps:          []StepRun{},
	}, nil
}
