package aor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Scheduler struct {
	cp       *ControlPlane
	mu       sync.RWMutex
	running  bool
	shutdown chan struct{}
}

func NewScheduler(cp *ControlPlane) *Scheduler {
	return &Scheduler{
		cp:       cp,
		shutdown: make(chan struct{}),
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler already running")
	}

	s.running = true

	// Start scheduling loop
	go s.schedulingLoop(ctx)

	log.Println("Scheduler started")
	return nil
}

func (s *Scheduler) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	close(s.shutdown)
	s.running = false

	log.Println("Scheduler shutdown")
	return nil
}

func (s *Scheduler) SubmitRun(ctx context.Context, run *WorkflowRun, spec *WorkflowSpec) error {
	// Create initial step runs for nodes with no dependencies
	readyNodes := s.findReadyNodes(spec.DAG)
	
	for _, nodeID := range readyNodes {
		node := s.findNode(spec.DAG, nodeID)
		if node == nil {
			continue
		}

		stepRun := &StepRun{
			ID:            uuid.New(),
			WorkflowRunID: run.ID,
			NodeID:        nodeID,
			Attempt:       1,
			Status:        StepStatusQueued,
			CreatedAt:     time.Now(),
		}

		if err := s.saveStepRun(ctx, stepRun); err != nil {
			return fmt.Errorf("failed to save step run: %w", err)
		}

		// Create and enqueue task
		task := &Task{
			ID:         stepRun.ID,
			RunID:      run.ID,
			NodeID:     nodeID,
			Attempt:    1,
			Node:       *node,
			Inputs:     s.resolveInputs(ctx, run, *node),
			DeadlineAt: time.Now().Add(30 * time.Minute), // Default deadline
		}

		if err := s.enqueueTask(ctx, task); err != nil {
			return fmt.Errorf("failed to enqueue task: %w", err)
		}
	}

	// Update run status to running
	if err := s.updateRunStatus(ctx, run.ID, RunStatusRunning); err != nil {
		return fmt.Errorf("failed to update run status: %w", err)
	}

	return nil
}

func (s *Scheduler) schedulingLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.shutdown:
			return
		case <-ticker.C:
			s.processCompletedSteps(ctx)
		}
	}
}

func (s *Scheduler) processCompletedSteps(ctx context.Context) {
	// Get completed steps that haven't been processed
	query := `SELECT sr.id, sr.workflow_run_id, sr.node_id, sr.status, wr.workflow_spec_id
			  FROM step_run sr
			  JOIN workflow_run wr ON sr.workflow_run_id = wr.id
			  WHERE sr.status IN ('succeeded', 'failed') 
			  AND NOT EXISTS (
				  SELECT 1 FROM step_run sr2 
				  WHERE sr2.workflow_run_id = sr.workflow_run_id 
				  AND sr2.node_id = sr.node_id 
				  AND sr2.id > sr.id
			  )`

	rows, err := s.cp.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Failed to query completed steps: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var stepID, runID, specID uuid.UUID
		var nodeID string
		var status StepStatus

		if err := rows.Scan(&stepID, &runID, &nodeID, &status, &specID); err != nil {
			log.Printf("Failed to scan step row: %v", err)
			continue
		}

		if err := s.processStepCompletion(ctx, runID, specID, nodeID, status); err != nil {
			log.Printf("Failed to process step completion: %v", err)
		}
	}
}

func (s *Scheduler) processStepCompletion(ctx context.Context, runID, specID uuid.UUID, completedNodeID string, status StepStatus) error {
	// Get workflow spec
	spec, err := s.getWorkflowSpec(ctx, specID)
	if err != nil {
		return fmt.Errorf("failed to get workflow spec: %w", err)
	}

	// Find nodes that are now ready to run
	readyNodes := s.findNodesReadyAfterCompletion(ctx, runID, spec.DAG, completedNodeID)

	// Schedule ready nodes
	for _, nodeID := range readyNodes {
		node := s.findNode(spec.DAG, nodeID)
		if node == nil {
			continue
		}

		// Check if this node already has a queued/running step
		if s.hasActiveStep(ctx, runID, nodeID) {
			continue
		}

		stepRun := &StepRun{
			ID:            uuid.New(),
			WorkflowRunID: runID,
			NodeID:        nodeID,
			Attempt:       1,
			Status:        StepStatusQueued,
			CreatedAt:     time.Now(),
		}

		if err := s.saveStepRun(ctx, stepRun); err != nil {
			log.Printf("Failed to save step run: %v", err)
			continue
		}

		// Get workflow run for inputs
		run, err := s.cp.GetWorkflowRun(ctx, runID)
		if err != nil {
			log.Printf("Failed to get workflow run: %v", err)
			continue
		}

		task := &Task{
			ID:         stepRun.ID,
			RunID:      runID,
			NodeID:     nodeID,
			Attempt:    1,
			Node:       *node,
			Inputs:     s.resolveInputs(ctx, run, *node),
			DeadlineAt: time.Now().Add(30 * time.Minute),
		}

		if err := s.enqueueTask(ctx, task); err != nil {
			log.Printf("Failed to enqueue task: %v", err)
		}
	}

	// Check if workflow is complete
	if s.isWorkflowComplete(ctx, runID, spec.DAG) {
		finalStatus := s.determineWorkflowStatus(ctx, runID)
		if err := s.updateRunStatus(ctx, runID, finalStatus); err != nil {
			log.Printf("Failed to update final workflow status: %v", err)
		}
	}

	return nil
}

func (s *Scheduler) findReadyNodes(dag DAG) []string {
	var ready []string
	
	// Find nodes with no incoming edges
	hasIncoming := make(map[string]bool)
	for _, edge := range dag.Edges {
		hasIncoming[edge.To] = true
	}

	for _, node := range dag.Nodes {
		if !hasIncoming[node.ID] {
			ready = append(ready, node.ID)
		}
	}

	return ready
}

func (s *Scheduler) findNodesReadyAfterCompletion(ctx context.Context, runID uuid.UUID, dag DAG, completedNodeID string) []string {
	var ready []string

	// Find nodes that depend on the completed node
	dependentNodes := make([]string, 0)
	for _, edge := range dag.Edges {
		if edge.From == completedNodeID {
			dependentNodes = append(dependentNodes, edge.To)
		}
	}

	// Check if all dependencies are satisfied for each dependent node
	for _, nodeID := range dependentNodes {
		if s.areAllDependenciesSatisfied(ctx, runID, dag, nodeID) {
			ready = append(ready, nodeID)
		}
	}

	return ready
}

func (s *Scheduler) areAllDependenciesSatisfied(ctx context.Context, runID uuid.UUID, dag DAG, nodeID string) bool {
	// Find all dependencies for this node
	dependencies := make([]string, 0)
	for _, edge := range dag.Edges {
		if edge.To == nodeID {
			dependencies = append(dependencies, edge.From)
		}
	}

	// Check if all dependencies have succeeded
	for _, depNodeID := range dependencies {
		if !s.hasSucceededStep(ctx, runID, depNodeID) {
			return false
		}
	}

	return true
}

func (s *Scheduler) hasSucceededStep(ctx context.Context, runID uuid.UUID, nodeID string) bool {
	query := `SELECT COUNT(*) FROM step_run 
			  WHERE workflow_run_id = $1 AND node_id = $2 AND status = 'succeeded'`
	
	var count int
	err := s.cp.db.QueryRowContext(ctx, query, runID, nodeID).Scan(&count)
	return err == nil && count > 0
}

func (s *Scheduler) hasActiveStep(ctx context.Context, runID uuid.UUID, nodeID string) bool {
	query := `SELECT COUNT(*) FROM step_run 
			  WHERE workflow_run_id = $1 AND node_id = $2 AND status IN ('queued', 'running')`
	
	var count int
	err := s.cp.db.QueryRowContext(ctx, query, runID, nodeID).Scan(&count)
	return err == nil && count > 0
}

func (s *Scheduler) isWorkflowComplete(ctx context.Context, runID uuid.UUID, dag DAG) bool {
	// Check if all nodes have completed (succeeded or failed)
	for _, node := range dag.Nodes {
		if !s.hasCompletedStep(ctx, runID, node.ID) {
			return false
		}
	}
	return true
}

func (s *Scheduler) hasCompletedStep(ctx context.Context, runID uuid.UUID, nodeID string) bool {
	query := `SELECT COUNT(*) FROM step_run 
			  WHERE workflow_run_id = $1 AND node_id = $2 AND status IN ('succeeded', 'failed')`
	
	var count int
	err := s.cp.db.QueryRowContext(ctx, query, runID, nodeID).Scan(&count)
	return err == nil && count > 0
}

func (s *Scheduler) determineWorkflowStatus(ctx context.Context, runID uuid.UUID) RunStatus {
	query := `SELECT status, COUNT(*) FROM step_run 
			  WHERE workflow_run_id = $1 GROUP BY status`
	
	rows, err := s.cp.db.QueryContext(ctx, query, runID)
	if err != nil {
		return RunStatusFailed
	}
	defer rows.Close()

	statusCounts := make(map[StepStatus]int)
	for rows.Next() {
		var status StepStatus
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}
		statusCounts[status] = count
	}

	// Determine overall status
	if statusCounts[StepStatusFailed] > 0 {
		if statusCounts[StepStatusSucceeded] > 0 {
			return RunStatusPartialSuccess
		}
		return RunStatusFailed
	}

	return RunStatusSucceeded
}

func (s *Scheduler) findNode(dag DAG, nodeID string) *Node {
	for _, node := range dag.Nodes {
		if node.ID == nodeID {
			return &node
		}
	}
	return nil
}

func (s *Scheduler) resolveInputs(ctx context.Context, run *WorkflowRun, node Node) map[string]interface{} {
	inputs := make(map[string]interface{})
	
	// Start with workflow inputs
	if workflowInputs, ok := run.Metadata["inputs"].(map[string]interface{}); ok {
		for k, v := range workflowInputs {
			inputs[k] = v
		}
	}

	// Add node-specific inputs
	for k, v := range node.Config.Inputs {
		inputs[k] = v
	}

	return inputs
}

func (s *Scheduler) enqueueTask(ctx context.Context, task *Task) error {
	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	subject := fmt.Sprintf("agentflow.tasks.%s", task.Node.Policy.Quality)
	if subject == "agentflow.tasks." {
		subject = "agentflow.tasks.Bronze" // Default quality
	}

	_, err = s.cp.js.Publish(subject, taskData)
	if err != nil {
		return fmt.Errorf("failed to publish task: %w", err)
	}

	return nil
}

func (s *Scheduler) saveStepRun(ctx context.Context, stepRun *StepRun) error {
	query := `INSERT INTO step_run (id, workflow_run_id, node_id, attempt, status, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6)`
	
	_, err := s.cp.db.ExecContext(ctx, query,
		stepRun.ID, stepRun.WorkflowRunID, stepRun.NodeID, stepRun.Attempt, stepRun.Status, stepRun.CreatedAt,
	)
	return err
}

func (s *Scheduler) updateRunStatus(ctx context.Context, runID uuid.UUID, status RunStatus) error {
	var endedAt *time.Time
	if status != RunStatusRunning && status != RunStatusQueued {
		now := time.Now()
		endedAt = &now
	}

	query := `UPDATE workflow_run SET status = $1, ended_at = $2 WHERE id = $3`
	_, err := s.cp.db.ExecContext(ctx, query, status, endedAt, runID)
	return err
}

func (s *Scheduler) getWorkflowSpec(ctx context.Context, specID uuid.UUID) (*WorkflowSpec, error) {
	query := `SELECT id, org_id, name, version, dag, metadata FROM workflow_spec WHERE id = $1`
	
	var spec WorkflowSpec
	var dagJSON, metadataJSON []byte
	
	err := s.cp.db.QueryRowContext(ctx, query, specID).Scan(
		&spec.ID, &spec.OrgID, &spec.Name, &spec.Version, &dagJSON, &metadataJSON,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(dagJSON, &spec.DAG); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(metadataJSON, &spec.Metadata); err != nil {
		return nil, err
	}

	return &spec, nil
}