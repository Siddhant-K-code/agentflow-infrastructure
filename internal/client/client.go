package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Client provides HTTP client for AgentFlow API
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	OrgID      string
}

// New creates a new AgentFlow API client
func New(baseURL, orgID string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: time.Second * 30,
		},
		OrgID: orgID,
	}
}

// WorkflowSpec represents a workflow specification
type WorkflowSpec struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Version int       `json:"version"`
	DAG     json.RawMessage `json:"dag"`
}

// WorkflowRun represents a workflow execution
type WorkflowRun struct {
	ID               uuid.UUID `json:"id"`
	WorkflowSpecID   uuid.UUID `json:"workflow_spec_id"`
	Status           string    `json:"status"`
	StartedAt        time.Time `json:"started_at"`
	EndedAt          *time.Time `json:"ended_at,omitempty"`
	CostCents        int64     `json:"cost_cents"`
	BudgetCents      *int64    `json:"budget_cents,omitempty"`
	Tags             []string  `json:"tags"`
}

// CreateWorkflowRequest represents workflow creation request
type CreateWorkflowRequest struct {
	DAG json.RawMessage `json:"dag"`
}

// CreateRunRequest represents run creation request
type CreateRunRequest struct {
	WorkflowName    string   `json:"workflow_name"`
	WorkflowVersion int      `json:"workflow_version"`
	BudgetCents     *int64   `json:"budget_cents,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	Input           map[string]interface{} `json:"input,omitempty"`
}

// SystemStatus represents system health status
type SystemStatus struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Queue    string `json:"queue"`
	Workers  int    `json:"workers"`
}

// request makes an HTTP request to the API
func (c *Client) request(method, path string, body interface{}) (*http.Response, error) {
	url := c.BaseURL + path
	
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.OrgID != "" {
		req.Header.Set("X-Org-ID", c.OrgID)
	}

	return c.HTTPClient.Do(req)
}

// CreateWorkflow creates a new workflow specification
func (c *Client) CreateWorkflow(name string, dag json.RawMessage) (*WorkflowSpec, error) {
	req := CreateWorkflowRequest{DAG: dag}
	resp, err := c.request("POST", fmt.Sprintf("/api/v1/workflows/%s/versions", name), req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var workflow WorkflowSpec
	if err := json.NewDecoder(resp.Body).Decode(&workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// GetWorkflow retrieves a workflow specification
func (c *Client) GetWorkflow(name string, version int) (*WorkflowSpec, error) {
	path := fmt.Sprintf("/api/v1/workflows/%s", name)
	if version > 0 {
		path = fmt.Sprintf("/api/v1/workflows/%s/versions/%d", name, version)
	}

	resp, err := c.request("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var workflow WorkflowSpec
	if err := json.NewDecoder(resp.Body).Decode(&workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// ListWorkflows lists all workflows
func (c *Client) ListWorkflows() ([]WorkflowSpec, error) {
	resp, err := c.request("GET", "/api/v1/workflows", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var workflows []WorkflowSpec
	if err := json.NewDecoder(resp.Body).Decode(&workflows); err != nil {
		return nil, err
	}

	return workflows, nil
}

// CreateRun starts a new workflow run
func (c *Client) CreateRun(req CreateRunRequest) (*WorkflowRun, error) {
	resp, err := c.request("POST", "/api/v1/runs", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var run WorkflowRun
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, err
	}

	return &run, nil
}

// GetRun retrieves a workflow run
func (c *Client) GetRun(runID uuid.UUID) (*WorkflowRun, error) {
	resp, err := c.request("GET", fmt.Sprintf("/api/v1/runs/%s", runID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var run WorkflowRun
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, err
	}

	return &run, nil
}

// ListRuns lists all workflow runs
func (c *Client) ListRuns() ([]WorkflowRun, error) {
	resp, err := c.request("GET", "/api/v1/runs", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var runs []WorkflowRun
	if err := json.NewDecoder(resp.Body).Decode(&runs); err != nil {
		return nil, err
	}

	return runs, nil
}

// CancelRun cancels a workflow run
func (c *Client) CancelRun(runID uuid.UUID) error {
	resp, err := c.request("POST", fmt.Sprintf("/api/v1/runs/%s/cancel", runID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetSystemStatus retrieves system health status
func (c *Client) GetSystemStatus() (*SystemStatus, error) {
	resp, err := c.request("GET", "/health", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var status SystemStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}

	return &status, nil
}