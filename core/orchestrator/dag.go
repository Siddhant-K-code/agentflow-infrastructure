package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// DAGExecutor manages directed acyclic graph execution of workflows
type DAGExecutor struct {
	config *Config
}

func NewDAGExecutor(config *Config) (*DAGExecutor, error) {
	return &DAGExecutor{
		config: config,
	}, nil
}

func (de *DAGExecutor) Shutdown() {
	// TODO: Implement graceful shutdown
}

// ExecuteWorkflow starts execution of a workflow
func (de *DAGExecutor) ExecuteWorkflow(workflow *Workflow) (*WorkflowExecution, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	
	execution := &WorkflowExecution{
		ID:        uuid.New().String(),
		Status:    "running",
		Agents:    make([]AgentExecution, 0),
		StartTime: time.Now(),
		Workflow:  workflow,
		ctx:       ctx,
		cancel:    cancel,
	}

	// Start execution in background
	go de.executeWorkflowAsync(execution)

	return execution, nil
}

func (de *DAGExecutor) executeWorkflowAsync(execution *WorkflowExecution) {
	defer execution.cancel()

	log.Printf("🚀 Starting workflow execution: %s", execution.ID)

	// Build dependency graph
	graph := buildDependencyGraph(execution.Workflow.Spec.Agents)
	
	// Execute agents in dependency order
	if err := de.executeDependencyGraph(execution, graph); err != nil {
		log.Printf("❌ Workflow %s failed: %v", execution.ID, err)
		execution.mutex.Lock()
		execution.Status = "failed"
		execution.Error = err.Error()
		now := time.Now()
		execution.EndTime = &now
		execution.mutex.Unlock()
		return
	}

	log.Printf("✅ Workflow %s completed successfully", execution.ID)
	execution.mutex.Lock()
	execution.Status = "completed"
	now := time.Now()
	execution.EndTime = &now
	execution.mutex.Unlock()
}

func (de *DAGExecutor) executeDependencyGraph(execution *WorkflowExecution, graph *DependencyGraph) error {
	completed := make(map[string]bool)
	outputs := make(map[string]map[string]interface{})

	for len(completed) < len(graph.Nodes) {
		// Find nodes ready to execute (all dependencies completed)
		ready := make([]*GraphNode, 0)
		for _, node := range graph.Nodes {
			if completed[node.Agent.Name] {
				continue
			}
			
			allDepsCompleted := true
			for _, dep := range node.Dependencies {
				if !completed[dep] {
					allDepsCompleted = false
					break
				}
			}
			
			if allDepsCompleted {
				ready = append(ready, node)
			}
		}

		if len(ready) == 0 {
			return fmt.Errorf("circular dependency detected or no ready nodes")
		}

		// Execute ready nodes in parallel
		var wg sync.WaitGroup
		errorChan := make(chan error, len(ready))

		for _, node := range ready {
			wg.Add(1)
			go func(n *GraphNode) {
				defer wg.Done()
				
				// Collect input from dependencies
				input := make(map[string]interface{})
				for _, dep := range n.Dependencies {
					if output, exists := outputs[dep]; exists {
						input[dep] = output
					}
				}

				// Execute agent
				output, err := de.executeAgent(execution, n.Agent, input)
				if err != nil {
					errorChan <- fmt.Errorf("agent %s failed: %w", n.Agent.Name, err)
					return
				}

				outputs[n.Agent.Name] = output
				completed[n.Agent.Name] = true
			}(node)
		}

		wg.Wait()
		close(errorChan)

		// Check for errors
		for err := range errorChan {
			return err
		}
	}

	return nil
}

func (de *DAGExecutor) executeAgent(execution *WorkflowExecution, agent Agent, input map[string]interface{}) (map[string]interface{}, error) {
	log.Printf("🤖 Executing agent: %s", agent.Name)

	agentExec := AgentExecution{
		Name:      agent.Name,
		Status:    "running",
		StartTime: time.Now(),
	}

	execution.mutex.Lock()
	execution.Agents = append(execution.Agents, agentExec)
	agentIndex := len(execution.Agents) - 1
	execution.mutex.Unlock()

	// For POC, simulate agent execution
	output, err := de.callMockAgent(agent, input)

	execution.mutex.Lock()
	if err != nil {
		execution.Agents[agentIndex].Status = "failed"
		execution.Agents[agentIndex].Error = err.Error()
	} else {
		execution.Agents[agentIndex].Status = "completed"
		execution.Agents[agentIndex].Output = output
	}
	now := time.Now()
	execution.Agents[agentIndex].EndTime = &now
	execution.mutex.Unlock()

	if err != nil {
		log.Printf("❌ Agent %s failed: %v", agent.Name, err)
		return nil, err
	}

	log.Printf("✅ Agent %s completed", agent.Name)
	return output, nil
}

func (de *DAGExecutor) callMockAgent(agent Agent, input map[string]interface{}) (map[string]interface{}, error) {
	// For POC: simulate agent execution
	time.Sleep(time.Duration(1+len(agent.Name)%3) * time.Second) // Simulate processing

	switch agent.Name {
	case "data-collector", "greeter":
		return map[string]interface{}{
			"message":   fmt.Sprintf("Hello from %s!", agent.Name),
			"timestamp": time.Now().Format(time.RFC3339),
			"data":      []string{"item1", "item2", "item3"},
		}, nil
	case "data-processor", "processor":
		return map[string]interface{}{
			"processed": true,
			"timestamp": time.Now().Format(time.RFC3339),
			"quality":   0.95,
			"input":     input,
		}, nil
	case "data-publisher", "publisher":
		return map[string]interface{}{
			"published": true,
			"timestamp": time.Now().Format(time.RFC3339),
			"endpoint":  "mock://published",
		}, nil
	default:
		return map[string]interface{}{
			"message":   fmt.Sprintf("Executed %s", agent.Name),
			"input":     input,
			"timestamp": time.Now().Format(time.RFC3339),
		}, nil
	}
}

// Dependency graph structures
type DependencyGraph struct {
	Nodes []*GraphNode
}

type GraphNode struct {
	Agent        Agent
	Dependencies []string
}

func buildDependencyGraph(agents []Agent) *DependencyGraph {
	nodes := make([]*GraphNode, 0, len(agents))
	
	for _, agent := range agents {
		node := &GraphNode{
			Agent:        agent,
			Dependencies: agent.DependsOn,
		}
		nodes = append(nodes, node)
	}

	return &DependencyGraph{Nodes: nodes}
}

// WorkflowExecution represents a running workflow instance
type WorkflowExecution struct {
	ID        string               `json:"id"`
	Status    string               `json:"status"`
	Agents    []AgentExecution     `json:"agents"`
	StartTime time.Time            `json:"start_time"`
	EndTime   *time.Time           `json:"end_time,omitempty"`
	Error     string               `json:"error,omitempty"`
	Workflow  *Workflow            `json:"workflow,omitempty"`
	ctx       context.Context      `json:"-"`
	cancel    context.CancelFunc   `json:"-"`
	mutex     sync.RWMutex         `json:"-"`
}

func (we *WorkflowExecution) Cancel() {
	we.cancel()
}

func (we *WorkflowExecution) GetStatus() string {
	we.mutex.RLock()
	defer we.mutex.RUnlock()
	return we.Status
}

func (we *WorkflowExecution) GetAgents() []AgentExecution {
	we.mutex.RLock()
	defer we.mutex.RUnlock()
	return we.Agents
}

// AgentExecution represents a running agent instance
type AgentExecution struct {
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	StartTime time.Time              `json:"start_time"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Output    map[string]interface{} `json:"output,omitempty"`
	Error     string                 `json:"error,omitempty"`
}