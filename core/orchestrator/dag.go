package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// DAGExecutor manages the execution of workflow DAGs
type DAGExecutor struct {
	config     *Config
	executions map[string]*WorkflowExecution
	mutex      sync.RWMutex
}

func NewDAGExecutor(config *Config) (*DAGExecutor, error) {
	return &DAGExecutor{
		config:     config,
		executions: make(map[string]*WorkflowExecution),
	}, nil
}

func (d *DAGExecutor) Shutdown() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	for _, execution := range d.executions {
		execution.Cancel()
	}
}

// WorkflowExecution represents a running workflow
type WorkflowExecution struct {
	ID        string
	Name      string
	Status    WorkflowStatus
	Agents    []*AgentExecution
	StartTime time.Time
	EndTime   *time.Time
	Context   context.Context
	Cancel    context.CancelFunc
	mutex     sync.RWMutex
}

// AgentExecution represents a running agent within a workflow
type AgentExecution struct {
	ID        string
	Name      string
	Status    AgentStatus
	LLMConfig LLMConfig
	StartTime time.Time
	EndTime   *time.Time
	Retries   int
	MaxRetries int
	Error     error
	mutex     sync.RWMutex
}

type WorkflowStatus string

const (
	WorkflowStatusPending   WorkflowStatus = "pending"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
)

type AgentStatus string

const (
	AgentStatusPending   AgentStatus = "pending"
	AgentStatusRunning   AgentStatus = "running"
	AgentStatusCompleted AgentStatus = "completed"
	AgentStatusFailed    AgentStatus = "failed"
	AgentStatusRetrying  AgentStatus = "retrying"
)

type LLMConfig struct {
	Provider string
	Model    string
	Config   map[string]string
}

// ExecuteWorkflow starts a new workflow execution
func (d *DAGExecutor) ExecuteWorkflow(workflow *Workflow) (*WorkflowExecution, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	ctx, cancel := context.WithCancel(context.Background())
	
	execution := &WorkflowExecution{
		ID:        uuid.New().String(),
		Name:      workflow.Metadata.Name,
		Status:    WorkflowStatusPending,
		StartTime: time.Now(),
		Context:   ctx,
		Cancel:    cancel,
	}
	
	// Convert workflow agents to agent executions
	for _, agent := range workflow.Spec.Agents {
		agentExec := &AgentExecution{
			ID:         uuid.New().String(),
			Name:       agent.Name,
			Status:     AgentStatusPending,
			LLMConfig:  agent.LLM,
			MaxRetries: agent.Retries,
		}
		if agentExec.MaxRetries == 0 {
			agentExec.MaxRetries = 3 // Default retries
		}
		execution.Agents = append(execution.Agents, agentExec)
	}
	
	d.executions[execution.ID] = execution
	
	// Start execution in background
	go d.runWorkflow(execution, workflow)
	
	return execution, nil
}

// runWorkflow executes the workflow DAG
func (d *DAGExecutor) runWorkflow(execution *WorkflowExecution, workflow *Workflow) {
	execution.mutex.Lock()
	execution.Status = WorkflowStatusRunning
	execution.mutex.Unlock()
	
	defer func() {
		execution.mutex.Lock()
		if execution.Status == WorkflowStatusRunning {
			execution.Status = WorkflowStatusCompleted
		}
		now := time.Now()
		execution.EndTime = &now
		execution.mutex.Unlock()
	}()
	
	// Build dependency graph
	depGraph := d.buildDependencyGraph(workflow.Spec.Agents)
	
	// Execute agents based on dependency order
	for level := range depGraph {
		agents := depGraph[level]
		
		// Execute agents in this level concurrently
		var wg sync.WaitGroup
		for _, agentName := range agents {
			wg.Add(1)
			go func(name string) {
				defer wg.Done()
				
				agent := d.findAgentExecution(execution, name)
				if agent != nil {
					d.executeAgent(execution.Context, agent)
				}
			}(agentName)
		}
		
		wg.Wait()
		
		// Check if any agent failed and we should stop
		if d.shouldStopExecution(execution) {
			execution.mutex.Lock()
			execution.Status = WorkflowStatusFailed
			execution.mutex.Unlock()
			return
		}
		
		// Check if context was cancelled
		select {
		case <-execution.Context.Done():
			execution.mutex.Lock()
			execution.Status = WorkflowStatusCancelled
			execution.mutex.Unlock()
			return
		default:
		}
	}
}

// buildDependencyGraph creates a levelized dependency graph
func (d *DAGExecutor) buildDependencyGraph(agents []Agent) map[int][]string {
	// Simple implementation - more sophisticated topological sort needed for production
	graph := make(map[int][]string)
	
	// For now, put agents with no dependencies at level 0
	// and agents with dependencies at level 1
	level0 := []string{}
	level1 := []string{}
	
	for _, agent := range agents {
		if len(agent.DependsOn) == 0 {
			level0 = append(level0, agent.Name)
		} else {
			level1 = append(level1, agent.Name)
		}
	}
	
	if len(level0) > 0 {
		graph[0] = level0
	}
	if len(level1) > 0 {
		graph[1] = level1
	}
	
	return graph
}

// executeAgent runs a single agent
func (d *DAGExecutor) executeAgent(ctx context.Context, agent *AgentExecution) {
	agent.mutex.Lock()
	agent.Status = AgentStatusRunning
	agent.StartTime = time.Now()
	agent.mutex.Unlock()
	
	defer func() {
		agent.mutex.Lock()
		if agent.Status == AgentStatusRunning {
			agent.Status = AgentStatusCompleted
		}
		now := time.Now()
		agent.EndTime = &now
		agent.mutex.Unlock()
	}()
	
	// Retry loop
	for attempt := 0; attempt <= agent.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return
		default:
		}
		
		if attempt > 0 {
			agent.mutex.Lock()
			agent.Status = AgentStatusRetrying
			agent.Retries = attempt
			agent.mutex.Unlock()
			
			// Exponential backoff
			backoff := time.Duration(attempt) * time.Second
			time.Sleep(backoff)
		}
		
		// Simulate agent execution
		if err := d.runAgentTask(ctx, agent); err != nil {
			agent.mutex.Lock()
			agent.Error = err
			agent.mutex.Unlock()
			
			if attempt == agent.MaxRetries {
				agent.mutex.Lock()
				agent.Status = AgentStatusFailed
				agent.mutex.Unlock()
				return
			}
			continue
		}
		
		// Success
		return
	}
}

// runAgentTask simulates running an agent task
func (d *DAGExecutor) runAgentTask(ctx context.Context, agent *AgentExecution) error {
	// TODO: Integrate with Rust runtime and LLM router
	
	// Simulate work
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Duration(2+agent.Retries) * time.Second):
		// Simulate occasional failures
		if agent.Retries == 0 && agent.Name == "data-processor" {
			return fmt.Errorf("simulated LLM timeout")
		}
		return nil
	}
}

func (d *DAGExecutor) findAgentExecution(execution *WorkflowExecution, name string) *AgentExecution {
	for _, agent := range execution.Agents {
		if agent.Name == name {
			return agent
		}
	}
	return nil
}

func (d *DAGExecutor) shouldStopExecution(execution *WorkflowExecution) bool {
	for _, agent := range execution.Agents {
		agent.mutex.RLock()
		status := agent.Status
		agent.mutex.RUnlock()
		
		if status == AgentStatusFailed {
			return true
		}
	}
	return false
}

// GetExecution retrieves a workflow execution by ID
func (d *DAGExecutor) GetExecution(id string) *WorkflowExecution {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.executions[id]
}

// ListExecutions returns all workflow executions
func (d *DAGExecutor) ListExecutions() []*WorkflowExecution {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	
	executions := make([]*WorkflowExecution, 0, len(d.executions))
	for _, exec := range d.executions {
		executions = append(executions, exec)
	}
	return executions
}

// Workflow represents the structure from deploy.go
type Workflow struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string            `yaml:"name"`
		Namespace string            `yaml:"namespace,omitempty"`
		Labels    map[string]string `yaml:"labels,omitempty"`
	} `yaml:"metadata"`
	Spec WorkflowSpec `yaml:"spec"`
}

type WorkflowSpec struct {
	Agents   []Agent   `yaml:"agents"`
	Triggers []Trigger `yaml:"triggers,omitempty"`
	Config   Config    `yaml:"config,omitempty"`
}

type Agent struct {
	Name      string            `yaml:"name"`
	Image     string            `yaml:"image"`
	LLM       LLMConfig         `yaml:"llm"`
	DependsOn []string          `yaml:"dependsOn,omitempty"`
	Resources Resources         `yaml:"resources,omitempty"`
	Env       map[string]string `yaml:"env,omitempty"`
	Timeout   string            `yaml:"timeout,omitempty"`
	Retries   int               `yaml:"retries,omitempty"`
}

type Resources struct {
	Memory string `yaml:"memory,omitempty"`
	CPU    string `yaml:"cpu,omitempty"`
}

type Trigger struct {
	Schedule string `yaml:"schedule,omitempty"`
	Webhook  string `yaml:"webhook,omitempty"`
	Event    string `yaml:"event,omitempty"`
}