package agentflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

// Client provides access to the AgentFlow API
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
	orgID      string
}

// ClientOptions configures the AgentFlow client
type ClientOptions struct {
	BaseURL    string
	Token      string
	OrgID      string
	HTTPClient *http.Client
	Timeout    time.Duration
}

// NewClient creates a new AgentFlow client
func NewClient(opts ClientOptions) *Client {
	if opts.HTTPClient == nil {
		timeout := opts.Timeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}
		opts.HTTPClient = &http.Client{Timeout: timeout}
	}

	if opts.BaseURL == "" {
		opts.BaseURL = "http://localhost:8080"
	}

	return &Client{
		baseURL:    opts.BaseURL,
		httpClient: opts.HTTPClient,
		token:      opts.Token,
		orgID:      opts.OrgID,
	}
}

// Workflows returns a workflow service client
func (c *Client) Workflows() *WorkflowService {
	return &WorkflowService{client: c}
}

// Prompts returns a prompt service client
func (c *Client) Prompts() *PromptService {
	return &PromptService{client: c}
}

// Traces returns a trace service client
func (c *Client) Traces() *TraceService {
	return &TraceService{client: c}
}

// Budgets returns a budget service client
func (c *Client) Budgets() *BudgetService {
	return &BudgetService{client: c}
}

// makeRequest makes an HTTP request to the API
func (c *Client) makeRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	
	if c.orgID != "" {
		req.Header.Set("X-Org-ID", c.orgID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// parseResponse parses an HTTP response into a struct
func (c *Client) parseResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	return nil
}

// WorkflowService provides workflow-related operations
type WorkflowService struct {
	client *Client
}

// Submit submits a workflow for execution
func (ws *WorkflowService) Submit(ctx context.Context, req *SubmitWorkflowRequest) (*WorkflowRun, error) {
	resp, err := ws.client.makeRequest(ctx, "POST", "/api/v1/workflows/runs", req)
	if err != nil {
		return nil, err
	}

	var result WorkflowRun
	if err := ws.client.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Get retrieves a workflow run by ID
func (ws *WorkflowService) Get(ctx context.Context, runID uuid.UUID) (*WorkflowRun, error) {
	path := fmt.Sprintf("/api/v1/workflows/runs/%s", runID)
	resp, err := ws.client.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result WorkflowRun
	if err := ws.client.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// List lists workflow runs with optional filters
func (ws *WorkflowService) List(ctx context.Context, opts *ListWorkflowsOptions) (*ListWorkflowsResponse, error) {
	path := "/api/v1/workflows/runs"
	
	if opts != nil {
		params := url.Values{}
		if opts.Status != "" {
			params.Add("status", opts.Status)
		}
		if opts.Limit > 0 {
			params.Add("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Since != "" {
			params.Add("since", opts.Since)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	resp, err := ws.client.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result ListWorkflowsResponse
	if err := ws.client.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Cancel cancels a workflow run
func (ws *WorkflowService) Cancel(ctx context.Context, runID uuid.UUID) error {
	path := fmt.Sprintf("/api/v1/workflows/runs/%s/cancel", runID)
	_, err := ws.client.makeRequest(ctx, "POST", path, nil)
	return err
}

// PromptService provides prompt-related operations
type PromptService struct {
	client *Client
}

// Create creates a new prompt version
func (ps *PromptService) Create(ctx context.Context, req *CreatePromptRequest) (*PromptTemplate, error) {
	resp, err := ps.client.makeRequest(ctx, "POST", "/api/v1/prompts", req)
	if err != nil {
		return nil, err
	}

	var result PromptTemplate
	if err := ps.client.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Get retrieves a prompt template
func (ps *PromptService) Get(ctx context.Context, name string, version int) (*PromptTemplate, error) {
	path := fmt.Sprintf("/api/v1/prompts/%s/versions/%d", name, version)
	resp, err := ps.client.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result PromptTemplate
	if err := ps.client.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Resolve resolves a prompt with inputs
func (ps *PromptService) Resolve(ctx context.Context, req *ResolvePromptRequest) (*ResolvePromptResponse, error) {
	resp, err := ps.client.makeRequest(ctx, "POST", "/api/v1/prompts/resolve", req)
	if err != nil {
		return nil, err
	}

	var result ResolvePromptResponse
	if err := ps.client.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Deploy deploys a prompt version
func (ps *PromptService) Deploy(ctx context.Context, req *DeployPromptRequest) (*PromptDeployment, error) {
	resp, err := ps.client.makeRequest(ctx, "POST", "/api/v1/prompts/deployments", req)
	if err != nil {
		return nil, err
	}

	var result PromptDeployment
	if err := ps.client.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// TraceService provides trace-related operations
type TraceService struct {
	client *Client
}

// Get retrieves a trace for a workflow run
func (ts *TraceService) Get(ctx context.Context, runID uuid.UUID) (*TraceResponse, error) {
	path := fmt.Sprintf("/api/v1/traces/runs/%s", runID)
	resp, err := ts.client.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result TraceResponse
	if err := ts.client.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Query queries traces with filters
func (ts *TraceService) Query(ctx context.Context, req *TraceQueryRequest) (*TraceResponse, error) {
	resp, err := ts.client.makeRequest(ctx, "POST", "/api/v1/traces/query", req)
	if err != nil {
		return nil, err
	}

	var result TraceResponse
	if err := ts.client.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Replay replays a workflow run
func (ts *TraceService) Replay(ctx context.Context, req *ReplayRequest) (*ReplayResponse, error) {
	resp, err := ts.client.makeRequest(ctx, "POST", "/api/v1/traces/replay", req)
	if err != nil {
		return nil, err
	}

	var result ReplayResponse
	if err := ts.client.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// BudgetService provides budget-related operations
type BudgetService struct {
	client *Client
}

// Create creates a new budget
func (bs *BudgetService) Create(ctx context.Context, req *CreateBudgetRequest) (*Budget, error) {
	resp, err := bs.client.makeRequest(ctx, "POST", "/api/v1/budgets", req)
	if err != nil {
		return nil, err
	}

	var result Budget
	if err := bs.client.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Get retrieves a budget by ID
func (bs *BudgetService) Get(ctx context.Context, budgetID uuid.UUID) (*Budget, error) {
	path := fmt.Sprintf("/api/v1/budgets/%s", budgetID)
	resp, err := bs.client.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result Budget
	if err := bs.client.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetStatus retrieves current budget status
func (bs *BudgetService) GetStatus(ctx context.Context) (*BudgetStatus, error) {
	resp, err := bs.client.makeRequest(ctx, "GET", "/api/v1/budgets/status", nil)
	if err != nil {
		return nil, err
	}

	var result BudgetStatus
	if err := bs.client.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}