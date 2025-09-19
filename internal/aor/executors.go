package aor

import (
	"context"
	"fmt"
	"log"
	"time"

)

// LLMExecutor handles LLM-based tasks
type LLMExecutor struct {
	worker *Worker
}

func NewLLMExecutor(worker *Worker) *LLMExecutor {
	return &LLMExecutor{worker: worker}
}

func (e *LLMExecutor) Execute(ctx context.Context, task *Task) (*TaskResult, error) {
	start := time.Now()
	
	promptRef := ""
	if task.Node != nil && task.Node.Config != nil {
		promptRef, _ = task.Node.Config["prompt_ref"].(string)
	}
	log.Printf("Executing LLM task %s with prompt %s", task.ID, promptRef)
	
	// Simulate processing time
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(100 * time.Millisecond):
	}
	
	// Mock successful execution
	return &TaskResult{
		TaskID:           task.ID,
		Status:           TaskStatusSucceeded,
		Output:           map[string]interface{}{"response": "Mock LLM response"},
		CostCents:        15,
		TokensPrompt:     100,
		TokensCompletion: 50,
		ExecutedAt:       time.Now(),
		Duration:         time.Since(start),
	}, nil
}

func (e *LLMExecutor) CanHandle(stepType string) bool {
	return stepType == "llm"
}

// ToolExecutor handles tool-based tasks
type ToolExecutor struct {
	worker *Worker
}

func NewToolExecutor(worker *Worker) *ToolExecutor {
	return &ToolExecutor{worker: worker}
}

func (e *ToolExecutor) Execute(ctx context.Context, task *Task) (*TaskResult, error) {
	start := time.Now()
	
	toolName := ""
	if task.Node != nil && task.Node.Config != nil {
		toolName, _ = task.Node.Config["tool_name"].(string)
	}
	log.Printf("Executing tool task %s with tool %s", task.ID, toolName)
	
	// Mock tool execution
	// In a real implementation, this would:
	// 1. Load tool definition
	// 2. Execute tool with inputs
	// 3. Return structured output
	
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(50 * time.Millisecond):
	}
	
	return &TaskResult{
		TaskID:     task.ID,
		Status:     TaskStatusSucceeded,
		Output:     map[string]interface{}{"result": fmt.Sprintf("Tool %s executed successfully", toolName)},
		CostCents:  5,
		ExecutedAt: time.Now(),
		Duration:   time.Since(start),
	}, nil
}

func (e *ToolExecutor) CanHandle(stepType string) bool {
	return stepType == "tool"
}

// HTTPExecutor handles HTTP-based tasks
type HTTPExecutor struct {
	worker *Worker
}

func NewHTTPExecutor(worker *Worker) *HTTPExecutor {
	return &HTTPExecutor{worker: worker}
}

func (e *HTTPExecutor) Execute(ctx context.Context, task *Task) (*TaskResult, error) {
	start := time.Now()
	
	log.Printf("Executing HTTP task %s", task.ID)
	
	// Mock HTTP request
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(200 * time.Millisecond):
	}
	
	return &TaskResult{
		TaskID:     task.ID,
		Status:     TaskStatusSucceeded,
		Output:     map[string]interface{}{"status": "success", "data": "mock response"},
		ExecutedAt: time.Now(),
		Duration:   time.Since(start),
	}, nil
}

func (e *HTTPExecutor) CanHandle(stepType string) bool {
	return stepType == "http"
}

// ScriptExecutor handles script-based tasks
type ScriptExecutor struct {
	worker *Worker
}

func NewScriptExecutor(worker *Worker) *ScriptExecutor {
	return &ScriptExecutor{worker: worker}
}

func (e *ScriptExecutor) Execute(ctx context.Context, task *Task) (*TaskResult, error) {
	start := time.Now()
	
	log.Printf("Executing script task %s", task.ID)
	
	// Mock script execution
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(300 * time.Millisecond):
	}
	
	return &TaskResult{
		TaskID:     task.ID,
		Status:     TaskStatusSucceeded,
		Output:     map[string]interface{}{"exit_code": 0, "stdout": "Script executed successfully"},
		ExecutedAt: time.Now(),
		Duration:   time.Since(start),
	}, nil
}

func (e *ScriptExecutor) CanHandle(stepType string) bool {
	return stepType == "script"
}