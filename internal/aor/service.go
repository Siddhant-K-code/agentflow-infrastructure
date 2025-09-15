package aor

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service implements the Agent Orchestration Runtime
type Service struct {
	db *gorm.DB
}

// NewService creates a new AOR service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// CreateWorkflowSpec creates a new workflow specification
func (s *Service) CreateWorkflowSpec(c *gin.Context) {
	name := c.Param("name")
	
	var req CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get org_id from context (would come from auth middleware)
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // placeholder

	// Get next version number
	var maxVersion int
	s.db.Model(&WorkflowSpec{}).
		Where("org_id = ? AND name = ?", orgID, name).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion)

	// Validate DAG structure
	var dag DAGSpec
	if err := json.Unmarshal(req.DAG, &dag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid DAG structure"})
		return
	}

	spec := WorkflowSpec{
		OrgID:   orgID,
		Name:    name,
		Version: maxVersion + 1,
		DAG:     req.DAG,
	}

	if err := s.db.Create(&spec).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, spec)
}

// GetWorkflowSpec retrieves a specific workflow specification
func (s *Service) GetWorkflowSpec(c *gin.Context) {
	name := c.Param("name")
	versionStr := c.Param("version")
	
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version"})
		return
	}

	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // placeholder

	var spec WorkflowSpec
	if err := s.db.Where("org_id = ? AND name = ? AND version = ?", orgID, name, version).
		First(&spec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, spec)
}

// CreateRun starts a new workflow run
func (s *Service) CreateRun(c *gin.Context) {
	var req RunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // placeholder

	// Find workflow spec
	query := s.db.Where("org_id = ? AND name = ?", orgID, req.WorkflowName)
	if req.WorkflowVersion != nil {
		query = query.Where("version = ?", *req.WorkflowVersion)
	} else {
		query = query.Order("version DESC")
	}

	var spec WorkflowSpec
	if err := query.First(&spec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create run
	run := WorkflowRun{
		WorkflowSpecID: spec.ID,
		Status:         "queued",
		BudgetCents:    req.BudgetCents,
		Metadata:       json.RawMessage("{}"),
		Tags:           json.RawMessage("[]"),
	}

	if len(req.Tags) > 0 {
		tagsJSON, _ := json.Marshal(req.Tags)
		run.Tags = tagsJSON
	}

	if err := s.db.Create(&run).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: Queue the run for execution

	c.JSON(http.StatusCreated, RunResponse{
		ID:     run.ID,
		Status: run.Status,
	})
}

// GetRun retrieves a workflow run status
func (s *Service) GetRun(c *gin.Context) {
	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid run ID"})
		return
	}

	var run WorkflowRun
	if err := s.db.Preload("WorkflowSpec").Preload("Steps").
		Where("id = ?", runID).First(&run).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "run not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, run)
}

// CancelRun cancels a running workflow
func (s *Service) CancelRun(c *gin.Context) {
	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid run ID"})
		return
	}

	var req CancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update run status
	result := s.db.Model(&WorkflowRun{}).
		Where("id = ? AND status IN ('queued', 'running')", runID).
		Update("status", "canceled")

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "run not found or not cancellable"})
		return
	}

	// TODO: Send cancellation signal to workers

	c.JSON(http.StatusOK, gin.H{"status": "canceled"})
}

// GetLatestWorkflowSpec retrieves the latest version of a workflow
func (s *Service) GetLatestWorkflowSpec(c *gin.Context) {
	name := c.Param("name")
	
	// Get org_id from context (would come from auth middleware)
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // placeholder

	var spec WorkflowSpec
	if err := s.db.Where("org_id = ? AND name = ?", orgID, name).
		Order("version DESC").First(&spec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, spec)
}

// ListWorkflows lists all workflows for an organization
func (s *Service) ListWorkflows(c *gin.Context) {
	// Get org_id from context (would come from auth middleware)
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // placeholder

	var specs []WorkflowSpec
	if err := s.db.Where("org_id = ?", orgID).
		Order("name ASC, version DESC").Find(&specs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Group by name and return latest version of each
	workflowMap := make(map[string]WorkflowSpec)
	for _, spec := range specs {
		if existing, exists := workflowMap[spec.Name]; !exists || spec.Version > existing.Version {
			workflowMap[spec.Name] = spec
		}
	}

	var result []WorkflowSpec
	for _, spec := range workflowMap {
		result = append(result, spec)
	}

	c.JSON(http.StatusOK, result)
}

// ListRuns lists all workflow runs
func (s *Service) ListRuns(c *gin.Context) {
	// Get org_id from context (would come from auth middleware)
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // placeholder

	var runs []WorkflowRun
	if err := s.db.Joins("JOIN workflow_spec ON workflow_run.workflow_spec_id = workflow_spec.id").
		Where("workflow_spec.org_id = ?", orgID).
		Order("workflow_run.started_at DESC").
		Find(&runs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, runs)
}

// SendSignal sends an external signal to a running workflow
func (s *Service) SendSignal(c *gin.Context) {
	runID, err := uuid.Parse(c.Param("run_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid run ID"})
		return
	}

	var req SignalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify run exists and is running
	var run WorkflowRun
	if err := s.db.Where("id = ? AND status = 'running'", runID).First(&run).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "run not found or not running"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: Send signal to workflow execution engine

	c.JSON(http.StatusOK, gin.H{"status": "signal sent"})
}

// WorkerHeartbeat handles worker heartbeat requests
func (s *Service) WorkerHeartbeat(c *gin.Context) {
	// TODO: Implement worker heartbeat logic
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// CompleteTask handles task completion from workers
func (s *Service) CompleteTask(c *gin.Context) {
	// TODO: Implement task completion logic
	c.JSON(http.StatusOK, gin.H{"status": "completed"})
}

// CreateWorkflowRequest represents a request to create a workflow
type CreateWorkflowRequest struct {
	DAG      json.RawMessage `json:"dag"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}