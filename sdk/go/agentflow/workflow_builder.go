package agentflow

import (
	"fmt"
	"time"
)

// WorkflowBuilder provides a fluent API for building workflows
type WorkflowBuilder struct {
	name    string
	version int
	nodes   []WorkflowNode
	edges   []WorkflowEdge
	errors  []error
}

// WorkflowNode represents a node in the workflow DAG
type WorkflowNode struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
	Policy *NodePolicy            `json:"policy,omitempty"`
}

// WorkflowEdge represents an edge in the workflow DAG
type WorkflowEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// NodePolicy defines execution constraints for a node
type NodePolicy struct {
	Quality    string        `json:"quality,omitempty"`
	SLAMillis  int           `json:"sla_ms,omitempty"`
	MaxRetries int           `json:"max_retries,omitempty"`
	Optional   bool          `json:"optional,omitempty"`
	Timeout    time.Duration `json:"timeout,omitempty"`
}

// NewWorkflow creates a new workflow builder
func NewWorkflow(name string) *WorkflowBuilder {
	return &WorkflowBuilder{
		name:  name,
		nodes: make([]WorkflowNode, 0),
		edges: make([]WorkflowEdge, 0),
	}
}

// Version sets the workflow version
func (wb *WorkflowBuilder) Version(version int) *WorkflowBuilder {
	wb.version = version
	return wb
}

// LLM adds an LLM node to the workflow
func (wb *WorkflowBuilder) LLM(id, promptRef string) *NodeBuilder {
	return wb.addNode(id, "llm", map[string]interface{}{
		"prompt_ref": promptRef,
	})
}

// Tool adds a tool node to the workflow
func (wb *WorkflowBuilder) Tool(id, toolName string, args map[string]interface{}) *NodeBuilder {
	return wb.addNode(id, "tool", map[string]interface{}{
		"tool_name": toolName,
		"tool_args": args,
	})
}

// Function adds a function node to the workflow
func (wb *WorkflowBuilder) Function(id, functionName string, args map[string]interface{}) *NodeBuilder {
	return wb.addNode(id, "function", map[string]interface{}{
		"function_name": functionName,
		"function_args": args,
	})
}

// Switch adds a switch node to the workflow
func (wb *WorkflowBuilder) Switch(id, switchOn string, cases map[string]string, defaultCase string) *NodeBuilder {
	return wb.addNode(id, "switch", map[string]interface{}{
		"switch_on":    switchOn,
		"cases":        cases,
		"default_case": defaultCase,
	})
}

// Map adds a map node to the workflow
func (wb *WorkflowBuilder) Map(id, iterateOver string, subWorkflow *WorkflowBuilder) *NodeBuilder {
	subDAG := map[string]interface{}{
		"nodes": subWorkflow.nodes,
		"edges": subWorkflow.edges,
	}
	
	return wb.addNode(id, "map", map[string]interface{}{
		"iterate_over": iterateOver,
		"sub_dag":      subDAG,
	})
}

// Reduce adds a reduce node to the workflow
func (wb *WorkflowBuilder) Reduce(id, functionName string, args map[string]interface{}) *NodeBuilder {
	return wb.addNode(id, "reduce", map[string]interface{}{
		"function_name": functionName,
		"function_args": args,
	})
}

// DependsOn adds a dependency between nodes
func (wb *WorkflowBuilder) DependsOn(from, to string) *WorkflowBuilder {
	wb.edges = append(wb.edges, WorkflowEdge{
		From: from,
		To:   to,
	})
	return wb
}

// Build builds the workflow specification
func (wb *WorkflowBuilder) Build() (*WorkflowSpec, error) {
	if len(wb.errors) > 0 {
		return nil, fmt.Errorf("workflow has errors: %v", wb.errors)
	}

	if wb.name == "" {
		return nil, fmt.Errorf("workflow name is required")
	}

	if len(wb.nodes) == 0 {
		return nil, fmt.Errorf("workflow must have at least one node")
	}

	// Validate DAG structure
	if err := wb.validateDAG(); err != nil {
		return nil, fmt.Errorf("invalid DAG: %w", err)
	}

	return &WorkflowSpec{
		Name:    wb.name,
		Version: wb.version,
		DAG: WorkflowDAG{
			Nodes: wb.nodes,
			Edges: wb.edges,
		},
	}, nil
}

// addNode adds a node to the workflow
func (wb *WorkflowBuilder) addNode(id, nodeType string, config map[string]interface{}) *NodeBuilder {
	// Check for duplicate IDs
	for _, node := range wb.nodes {
		if node.ID == id {
			wb.errors = append(wb.errors, fmt.Errorf("duplicate node ID: %s", id))
			return &NodeBuilder{wb: wb, nodeIndex: -1}
		}
	}

	node := WorkflowNode{
		ID:     id,
		Type:   nodeType,
		Config: config,
	}

	wb.nodes = append(wb.nodes, node)
	return &NodeBuilder{
		wb:        wb,
		nodeIndex: len(wb.nodes) - 1,
	}
}

// validateDAG validates the DAG structure
func (wb *WorkflowBuilder) validateDAG() error {
	// Check for cycles using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	// Build adjacency list
	adjList := make(map[string][]string)
	for _, edge := range wb.edges {
		adjList[edge.From] = append(adjList[edge.From], edge.To)
	}

	// Check each node
	for _, node := range wb.nodes {
		if !visited[node.ID] {
			if wb.hasCycle(node.ID, visited, recStack, adjList) {
				return fmt.Errorf("cycle detected in workflow DAG")
			}
		}
	}

	// Validate edge references
	nodeIDs := make(map[string]bool)
	for _, node := range wb.nodes {
		nodeIDs[node.ID] = true
	}

	for _, edge := range wb.edges {
		if !nodeIDs[edge.From] {
			return fmt.Errorf("edge references unknown node: %s", edge.From)
		}
		if !nodeIDs[edge.To] {
			return fmt.Errorf("edge references unknown node: %s", edge.To)
		}
	}

	return nil
}

// hasCycle checks for cycles using DFS
func (wb *WorkflowBuilder) hasCycle(nodeID string, visited, recStack map[string]bool, adjList map[string][]string) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	for _, neighbor := range adjList[nodeID] {
		if !visited[neighbor] {
			if wb.hasCycle(neighbor, visited, recStack, adjList) {
				return true
			}
		} else if recStack[neighbor] {
			return true
		}
	}

	recStack[nodeID] = false
	return false
}

// NodeBuilder provides a fluent API for configuring nodes
type NodeBuilder struct {
	wb        *WorkflowBuilder
	nodeIndex int
}

// WithPolicy sets the execution policy for the node
func (nb *NodeBuilder) WithPolicy(policy *NodePolicy) *NodeBuilder {
	if nb.nodeIndex >= 0 && nb.nodeIndex < len(nb.wb.nodes) {
		nb.wb.nodes[nb.nodeIndex].Policy = policy
	}
	return nb
}

// WithQuality sets the quality tier for the node
func (nb *NodeBuilder) WithQuality(quality string) *NodeBuilder {
	if nb.nodeIndex >= 0 && nb.nodeIndex < len(nb.wb.nodes) {
		if nb.wb.nodes[nb.nodeIndex].Policy == nil {
			nb.wb.nodes[nb.nodeIndex].Policy = &NodePolicy{}
		}
		nb.wb.nodes[nb.nodeIndex].Policy.Quality = quality
	}
	return nb
}

// WithSLA sets the SLA for the node
func (nb *NodeBuilder) WithSLA(sla time.Duration) *NodeBuilder {
	if nb.nodeIndex >= 0 && nb.nodeIndex < len(nb.wb.nodes) {
		if nb.wb.nodes[nb.nodeIndex].Policy == nil {
			nb.wb.nodes[nb.nodeIndex].Policy = &NodePolicy{}
		}
		nb.wb.nodes[nb.nodeIndex].Policy.SLAMillis = int(sla.Milliseconds())
	}
	return nb
}

// WithRetries sets the maximum retries for the node
func (nb *NodeBuilder) WithRetries(maxRetries int) *NodeBuilder {
	if nb.nodeIndex >= 0 && nb.nodeIndex < len(nb.wb.nodes) {
		if nb.wb.nodes[nb.nodeIndex].Policy == nil {
			nb.wb.nodes[nb.nodeIndex].Policy = &NodePolicy{}
		}
		nb.wb.nodes[nb.nodeIndex].Policy.MaxRetries = maxRetries
	}
	return nb
}

// Optional marks the node as optional
func (nb *NodeBuilder) Optional() *NodeBuilder {
	if nb.nodeIndex >= 0 && nb.nodeIndex < len(nb.wb.nodes) {
		if nb.wb.nodes[nb.nodeIndex].Policy == nil {
			nb.wb.nodes[nb.nodeIndex].Policy = &NodePolicy{}
		}
		nb.wb.nodes[nb.nodeIndex].Policy.Optional = true
	}
	return nb
}

// WithInputs sets inputs for the node
func (nb *NodeBuilder) WithInputs(inputs map[string]string) *NodeBuilder {
	if nb.nodeIndex >= 0 && nb.nodeIndex < len(nb.wb.nodes) {
		if nb.wb.nodes[nb.nodeIndex].Config == nil {
			nb.wb.nodes[nb.nodeIndex].Config = make(map[string]interface{})
		}
		nb.wb.nodes[nb.nodeIndex].Config["inputs"] = inputs
	}
	return nb
}

// DependsOn adds a dependency from another node to this node
func (nb *NodeBuilder) DependsOn(fromNodeID string) *NodeBuilder {
	if nb.nodeIndex >= 0 && nb.nodeIndex < len(nb.wb.nodes) {
		nodeID := nb.wb.nodes[nb.nodeIndex].ID
		nb.wb.edges = append(nb.wb.edges, WorkflowEdge{
			From: fromNodeID,
			To:   nodeID,
		})
	}
	return nb
}

// End returns to the workflow builder
func (nb *NodeBuilder) End() *WorkflowBuilder {
	return nb.wb
}

// WorkflowSpec represents a complete workflow specification
type WorkflowSpec struct {
	Name     string      `json:"name"`
	Version  int         `json:"version"`
	DAG      WorkflowDAG `json:"dag"`
	Metadata Metadata    `json:"metadata,omitempty"`
}

// WorkflowDAG represents the workflow DAG
type WorkflowDAG struct {
	Nodes []WorkflowNode `json:"nodes"`
	Edges []WorkflowEdge `json:"edges"`
}

// Metadata represents workflow metadata
type Metadata map[string]interface{}

// Example usage functions

// ExampleDocumentAnalysis creates an example document analysis workflow
func ExampleDocumentAnalysis() *WorkflowBuilder {
	return NewWorkflow("document_analysis").
		Version(1).
		Tool("ingest", "s3.fetch", map[string]interface{}{
			"bucket": "documents",
			"key":    "{{document_key}}",
		}).
		Function("chunk", "text_chunker", map[string]interface{}{
			"chunk_size": 1000,
			"overlap":    100,
		}).WithInputs(map[string]string{
			"content": "ingest.output",
		}).DependsOn("ingest").
		LLM("analyze", "document_analyzer@3").
		WithQuality("Gold").
		WithSLA(30 * time.Second).
		WithInputs(map[string]string{
			"chunks": "chunk.output",
		}).DependsOn("chunk").
		Function("summarize", "text_summarizer", map[string]interface{}{
			"max_length": 500,
		}).WithInputs(map[string]string{
			"analysis": "analyze.output",
		}).DependsOn("analyze")
}

// ExampleMapReduce creates an example map-reduce workflow
func ExampleMapReduce() *WorkflowBuilder {
	// Sub-workflow for processing individual items
	itemProcessor := NewWorkflow("process_item").
		LLM("process", "item_processor@1").
		WithQuality("Silver")

	return NewWorkflow("batch_processing").
		Version(1).
		Tool("load_data", "database.query", map[string]interface{}{
			"query": "SELECT * FROM items WHERE status = 'pending'",
		}).
		Map("process_items", "load_data.output", itemProcessor).
		DependsOn("load_data").
		Reduce("aggregate", "result_aggregator", map[string]interface{}{
			"operation": "sum",
		}).WithInputs(map[string]string{
			"results": "process_items.output",
		}).DependsOn("process_items")
}