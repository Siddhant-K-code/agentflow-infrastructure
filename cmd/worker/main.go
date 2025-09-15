package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
)

type Worker struct {
	ID        string
	ServerURL string
	Client    *http.Client
}

type Task struct {
	ID       uuid.UUID `json:"id"`
	RunID    uuid.UUID `json:"run_id"`
	NodeID   string    `json:"node_id"`
	NodeType string    `json:"node_type"`
	Config   map[string]interface{} `json:"config"`
	Input    map[string]interface{} `json:"input"`
}

type TaskResult struct {
	TaskID  uuid.UUID `json:"task_id"`
	Status  string    `json:"status"`
	Output  map[string]interface{} `json:"output,omitempty"`
	Error   string    `json:"error,omitempty"`
	Metrics map[string]interface{} `json:"metrics,omitempty"`
}

type LLMProvider struct {
	APIKey string
}

func NewLLMProvider() *LLMProvider {
	return &LLMProvider{
		APIKey: os.Getenv("OPENAI_API_KEY"),
	}
}

func (p *LLMProvider) Execute(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	model, _ := config["model"].(string)
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	// Extract prompt and context
	prompt, _ := config["prompt"].(string)
	context, _ := input["context"].(string)
	
	if prompt == "" {
		return nil, fmt.Errorf("no prompt specified")
	}

	// Simple prompt template replacement
	finalPrompt := fmt.Sprintf("%s\n\nContext: %s", prompt, context)

	// Mock LLM response for POC (replace with actual OpenAI API call)
	response := map[string]interface{}{
		"text": fmt.Sprintf("Mock LLM response for prompt: %s (with context: %s)", prompt, context),
		"model": model,
		"tokens_used": 150,
		"final_prompt_length": len(finalPrompt),
	}

	// If we have an actual API key, we could make real calls here
	if p.APIKey != "" {
		log.Printf("Would make real OpenAI API call with model %s", model)
	}

	return response, nil
}

func (w *Worker) heartbeat() error {
	data := map[string]interface{}{
		"worker_id": w.ID,
		"status": "idle",
		"timestamp": time.Now(),
	}

	jsonData, _ := json.Marshal(data)
	resp, err := w.Client.Post(w.ServerURL+"/api/v1/tasks/heartbeat", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (w *Worker) pollForTasks() (*Task, error) {
	// In a real implementation, this would connect to NATS or poll a task queue
	// For POC, we'll just return nil (no tasks)
	return nil, nil
}

func (w *Worker) executeTask(task *Task) TaskResult {
	result := TaskResult{
		TaskID: task.ID,
		Status: "succeeded",
		Metrics: map[string]interface{}{
			"start_time": time.Now(),
		},
	}

	switch task.NodeType {
	case "llm":
		llmProvider := NewLLMProvider()
		output, err := llmProvider.Execute(task.Config, task.Input)
		if err != nil {
			result.Status = "failed"
			result.Error = err.Error()
		} else {
			result.Output = output
		}

	case "function":
		// Mock function execution
		result.Output = map[string]interface{}{
			"result": "Function executed successfully",
		}

	case "tool":
		// Mock tool execution
		result.Output = map[string]interface{}{
			"result": "Tool executed successfully",
		}

	default:
		result.Status = "failed"
		result.Error = fmt.Sprintf("unknown node type: %s", task.NodeType)
	}

	result.Metrics["end_time"] = time.Now()
	return result
}

func (w *Worker) submitResult(result TaskResult) error {
	jsonData, _ := json.Marshal(result)
	resp, err := w.Client.Post(w.ServerURL+"/api/v1/tasks/complete", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (w *Worker) run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	log.Printf("Worker %s started, polling for tasks...", w.ID)

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker shutting down...")
			return

		case <-ticker.C:
			// Send heartbeat
			if err := w.heartbeat(); err != nil {
				log.Printf("Heartbeat failed: %v", err)
				continue
			}

			// Poll for tasks
			task, err := w.pollForTasks()
			if err != nil {
				log.Printf("Failed to poll for tasks: %v", err)
				continue
			}

			if task != nil {
				log.Printf("Executing task %s (node: %s, type: %s)", task.ID, task.NodeID, task.NodeType)
				result := w.executeTask(task)
				
				if err := w.submitResult(result); err != nil {
					log.Printf("Failed to submit result: %v", err)
				} else {
					log.Printf("Task %s completed with status: %s", task.ID, result.Status)
				}
			}
		}
	}
}

var (
	serverURL = flag.String("server", "http://localhost:8080", "AgentFlow server URL")
	workerID  = flag.String("id", "", "Worker ID")
	debug     = flag.Bool("debug", false, "Enable debug mode")
)

func main() {
	flag.Parse()

	if *workerID == "" {
		*workerID = fmt.Sprintf("worker-%s", uuid.New().String()[:8])
	}

	worker := &Worker{
		ID:        *workerID,
		ServerURL: *serverURL,
		Client: &http.Client{
			Timeout: time.Second * 30,
		},
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	// Start worker
	worker.run(ctx)
}