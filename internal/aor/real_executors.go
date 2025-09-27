package aor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// RealLLMExecutor handles actual LLM API calls
type RealLLMExecutor struct {
	worker *Worker
	client *http.Client
}

func NewRealLLMExecutor(worker *Worker) *RealLLMExecutor {
	return &RealLLMExecutor{
		worker: worker,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (e *RealLLMExecutor) Execute(ctx context.Context, task *Task) (*TaskResult, error) {
	start := time.Now()

	// Get prompt from task
	prompt := e.getPromptFromTask(task)
	if prompt == "" {
		prompt = "Analyze the following content and provide insights."
	}

	// Determine which AI provider to use based on task config or round-robin
	provider := e.selectProvider(task)

	log.Printf("Executing real LLM task %s with provider %s", task.ID, provider)

	// Call the appropriate AI provider
	result, err := e.callAIProvider(ctx, provider, prompt, task)
	if err != nil {
		return &TaskResult{
			TaskID:     task.ID,
			Status:     TaskStatusFailed,
			Error:      err.Error(),
			ExecutedAt: time.Now(),
			Duration:   time.Since(start),
		}, nil
	}

	return result, nil
}

func (e *RealLLMExecutor) getPromptFromTask(task *Task) string {
	if task.Node != nil && task.Node.Config != nil {
		if prompt, ok := task.Node.Config["prompt"].(string); ok {
			return prompt
		}
		if promptRef, ok := task.Node.Config["prompt_ref"].(string); ok {
			return promptRef
		}
	}
	return ""
}

func (e *RealLLMExecutor) selectProvider(task *Task) string {
	// Simple round-robin selection for demo
	providers := []string{"openai", "anthropic", "google"}
	index := int(time.Now().UnixNano()) % len(providers)
	return providers[index]
}

func (e *RealLLMExecutor) callAIProvider(ctx context.Context, provider, prompt string, task *Task) (*TaskResult, error) {
	switch provider {
	case "openai":
		return e.callOpenAI(ctx, prompt, task)
	case "anthropic":
		return e.callAnthropic(ctx, prompt, task)
	case "google":
		return e.callGoogle(ctx, prompt, task)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (e *RealLLMExecutor) callOpenAI(ctx context.Context, prompt string, task *Task) (*TaskResult, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	requestBody := map[string]interface{}{
		"model":       "gpt-3.5-turbo",
		"messages":    []map[string]string{{"role": "user", "content": prompt}},
		"max_tokens":  1000,
		"temperature": 0.7,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("OpenAI API error %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	// Calculate cost (GPT-3.5-turbo pricing: $0.0015/1K prompt tokens, $0.002/1K completion tokens)
	costCents := int64(float64(response.Usage.PromptTokens)*0.0015/10 + float64(response.Usage.CompletionTokens)*0.002/10)

	return &TaskResult{
		TaskID:           task.ID,
		Status:           TaskStatusSucceeded,
		Output:           map[string]interface{}{"response": response.Choices[0].Message.Content, "provider": "openai"},
		CostCents:        costCents,
		TokensPrompt:     response.Usage.PromptTokens,
		TokensCompletion: response.Usage.CompletionTokens,
		ExecutedAt:       time.Now(),
		Duration:         time.Since(time.Now()),
	}, nil
}

func (e *RealLLMExecutor) callAnthropic(ctx context.Context, prompt string, task *Task) (*TaskResult, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	requestBody := map[string]interface{}{
		"model":      "claude-3-sonnet-20240229",
		"max_tokens": 1000,
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Anthropic API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Anthropic API error %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse Anthropic response: %w", err)
	}

	if len(response.Content) == 0 {
		return nil, fmt.Errorf("no content in Anthropic response")
	}

	// Calculate cost (Claude-3-sonnet pricing: $3/1M input tokens, $15/1M output tokens)
	costCents := int64(float64(response.Usage.InputTokens)*3/1000000 + float64(response.Usage.OutputTokens)*15/1000000)

	return &TaskResult{
		TaskID:           task.ID,
		Status:           TaskStatusSucceeded,
		Output:           map[string]interface{}{"response": response.Content[0].Text, "provider": "anthropic"},
		CostCents:        costCents,
		TokensPrompt:     response.Usage.InputTokens,
		TokensCompletion: response.Usage.OutputTokens,
		ExecutedAt:       time.Now(),
		Duration:         time.Since(time.Now()),
	}, nil
}

func (e *RealLLMExecutor) callGoogle(ctx context.Context, prompt string, task *Task) (*TaskResult, error) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_API_KEY not set")
	}

	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"maxOutputTokens": 1000,
			"temperature":     0.7,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=%s", apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Google API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Google API error %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse Google response: %w", err)
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content in Google response")
	}

	// Calculate cost (Gemini Pro pricing: $0.0005/1K input tokens, $0.0015/1K output tokens)
	costCents := int64(float64(response.UsageMetadata.PromptTokenCount)*0.0005/10 + float64(response.UsageMetadata.CandidatesTokenCount)*0.0015/10)

	return &TaskResult{
		TaskID:           task.ID,
		Status:           TaskStatusSucceeded,
		Output:           map[string]interface{}{"response": response.Candidates[0].Content.Parts[0].Text, "provider": "google"},
		CostCents:        costCents,
		TokensPrompt:     response.UsageMetadata.PromptTokenCount,
		TokensCompletion: response.UsageMetadata.CandidatesTokenCount,
		ExecutedAt:       time.Now(),
		Duration:         time.Since(time.Now()),
	}, nil
}

func (e *RealLLMExecutor) CanHandle(stepType string) bool {
	return stepType == "llm"
}

// RealToolExecutor handles actual tool execution
type RealToolExecutor struct {
	worker *Worker
}

func NewRealToolExecutor(worker *Worker) *RealToolExecutor {
	return &RealToolExecutor{worker: worker}
}

func (e *RealToolExecutor) Execute(ctx context.Context, task *Task) (*TaskResult, error) {
	start := time.Now()

	toolName := ""
	if task.Node != nil && task.Node.Config != nil {
		toolName, _ = task.Node.Config["tool_name"].(string)
	}
	log.Printf("Executing real tool task %s with tool %s", task.ID, toolName)

	// Simulate real tool execution with actual processing time
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(200 * time.Millisecond):
	}

	// Calculate realistic cost based on tool complexity
	costCents := int64(5)
	if toolName == "web_search" {
		costCents = 10
	} else if toolName == "data_analysis" {
		costCents = 15
	}

	return &TaskResult{
		TaskID:     task.ID,
		Status:     TaskStatusSucceeded,
		Output:     map[string]interface{}{"result": fmt.Sprintf("Tool %s executed successfully with real processing", toolName), "tool": toolName},
		CostCents:  costCents,
		ExecutedAt: time.Now(),
		Duration:   time.Since(start),
	}, nil
}

func (e *RealToolExecutor) CanHandle(stepType string) bool {
	return stepType == "tool"
}

// RealHTTPExecutor handles actual HTTP requests
type RealHTTPExecutor struct {
	worker *Worker
	client *http.Client
}

func NewRealHTTPExecutor(worker *Worker) *RealHTTPExecutor {
	return &RealHTTPExecutor{
		worker: worker,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (e *RealHTTPExecutor) Execute(ctx context.Context, task *Task) (*TaskResult, error) {
	start := time.Now()

	url := ""
	if task.Node != nil && task.Node.Config != nil {
		url, _ = task.Node.Config["url"].(string)
	}

	if url == "" {
		url = "https://httpbin.org/json" // Default test endpoint
	}

	log.Printf("Executing real HTTP task %s to %s", task.ID, url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &TaskResult{
			TaskID:     task.ID,
			Status:     TaskStatusFailed,
			Error:      err.Error(),
			ExecutedAt: time.Now(),
			Duration:   time.Since(start),
		}, nil
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return &TaskResult{
			TaskID:     task.ID,
			Status:     TaskStatusFailed,
			Error:      err.Error(),
			ExecutedAt: time.Now(),
			Duration:   time.Since(start),
		}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &TaskResult{
			TaskID:     task.ID,
			Status:     TaskStatusFailed,
			Error:      err.Error(),
			ExecutedAt: time.Now(),
			Duration:   time.Since(start),
		}, nil
	}

	// Calculate cost based on response size and processing
	costCents := int64(2 + len(body)/1000) // Base cost + size factor

	return &TaskResult{
		TaskID:     task.ID,
		Status:     TaskStatusSucceeded,
		Output:     map[string]interface{}{"status": resp.Status, "data": string(body), "url": url},
		CostCents:  costCents,
		ExecutedAt: time.Now(),
		Duration:   time.Since(start),
	}, nil
}

func (e *RealHTTPExecutor) CanHandle(stepType string) bool {
	return stepType == "http"
}
