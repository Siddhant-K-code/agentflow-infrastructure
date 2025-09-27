package main

import (
	"strconv"
	"time"

	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/aor"
	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/cas"
	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/pop"
	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/scl"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type HTTPServer struct {
	controlPlane *aor.ControlPlane
	popService   *pop.Service
	casService   *cas.Service
	sclService   *scl.Service
}

func NewHTTPServer(cp *aor.ControlPlane, popSvc *pop.Service, casSvc *cas.Service, sclSvc *scl.Service) *HTTPServer {
	return &HTTPServer{
		controlPlane: cp,
		popService:   popSvc,
		casService:   casSvc,
		sclService:   sclService,
	}
}

func (s *HTTPServer) SetupRoutes() *gin.Engine {
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"timestamp": time.Now(),
			"version":   "1.0.0",
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Workflow management
		v1.POST("/workflows/runs", s.submitWorkflow)
		v1.GET("/workflows/runs/:id", s.getWorkflowRun)
		v1.POST("/workflows/runs/:id/cancel", s.cancelWorkflowRun)
		v1.GET("/workflows/runs", s.listWorkflowRuns)

		// Prompt management
		v1.POST("/prompts", s.createPrompt)
		v1.GET("/prompts/:name/versions", s.listPromptVersions)
		v1.GET("/prompts/:name/versions/:version", s.getPromptVersion)
		v1.POST("/prompts/:name/resolve", s.resolvePrompt)

		// Cost management
		v1.GET("/budgets/status", s.getBudgetStatus)
		v1.GET("/costs/analytics", s.getCostAnalytics)
		v1.POST("/budgets", s.createBudget)

		// Security & Compliance
		v1.POST("/scl/redact", s.redactContent)
		v1.POST("/scl/unredact", s.unredactContent)

		// System status
		v1.GET("/status", s.getSystemStatus)
	}

	return r
}

// Workflow endpoints
func (s *HTTPServer) submitWorkflow(c *gin.Context) {
	var req aor.RunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Create a demo organization ID
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	run, err := s.controlPlane.SubmitWorkflow(c.Request.Context(), &req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, run)
}

func (s *HTTPServer) getWorkflowRun(c *gin.Context) {
	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid run ID"})
		return
	}

	run, err := s.controlPlane.GetWorkflowRun(c.Request.Context(), runID)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, run)
}

func (s *HTTPServer) cancelWorkflowRun(c *gin.Context) {
	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid run ID"})
		return
	}

	err = s.controlPlane.CancelWorkflowRun(c.Request.Context(), runID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "cancelled"})
}

func (s *HTTPServer) listWorkflowRuns(c *gin.Context) {
	// Demo implementation - return mock data
	runs := []aor.WorkflowRun{
		{
			ID:     uuid.New(),
			Status: aor.RunStatusCompleted,
			Metadata: map[string]interface{}{
				"workflow_name": "document_analysis",
				"cost_cents":    150,
			},
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:     uuid.New(),
			Status: aor.RunStatusRunning,
			Metadata: map[string]interface{}{
				"workflow_name": "data_processing",
				"cost_cents":    75,
			},
			CreatedAt: time.Now().Add(-30 * time.Minute),
		},
	}

	c.JSON(200, gin.H{"runs": runs})
}

// Prompt endpoints
func (s *HTTPServer) createPrompt(c *gin.Context) {
	var req pop.CreatePromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	prompt, err := s.popService.CreatePromptVersion(c.Request.Context(), orgID, &req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, prompt)
}

func (s *HTTPServer) listPromptVersions(c *gin.Context) {
	// Demo implementation
	versions := []pop.PromptTemplate{
		{
			ID:        uuid.New(),
			Name:      c.Param("name"),
			Version:   1,
			Template:  "Analyze the following document: {{content}}",
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        uuid.New(),
			Name:      c.Param("name"),
			Version:   2,
			Template:  "Please analyze this document and provide insights: {{content}}",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
	}

	c.JSON(200, gin.H{"versions": versions})
}

func (s *HTTPServer) getPromptVersion(c *gin.Context) {
	version, _ := strconv.Atoi(c.Param("version"))

	// Demo implementation
	prompt := pop.PromptTemplate{
		ID:       uuid.New(),
		Name:     c.Param("name"),
		Version:  version,
		Template: "Analyze the following document: {{content}}",
		Schema: pop.Schema{
			Type: "object",
			Properties: map[string]interface{}{
				"content": map[string]interface{}{
					"type": "string",
				},
			},
			Required: []string{"content"},
		},
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}

	c.JSON(200, prompt)
}

func (s *HTTPServer) resolvePrompt(c *gin.Context) {
	var req pop.PromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	response, err := s.popService.ResolvePrompt(c.Request.Context(), orgID, &req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, response)
}

// Cost management endpoints
func (s *HTTPServer) getBudgetStatus(c *gin.Context) {
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	status, err := s.casService.GetBudgetStatus(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, status)
}

func (s *HTTPServer) getCostAnalytics(c *gin.Context) {
	// Demo analytics data
	analytics := gin.H{
		"total_spent_cents": 1250,
		"daily_spending": []gin.H{
			{"date": "2024-01-01", "amount_cents": 150},
			{"date": "2024-01-02", "amount_cents": 200},
			{"date": "2024-01-03", "amount_cents": 175},
		},
		"cost_by_provider": []gin.H{
			{"provider": "openai", "amount_cents": 800},
			{"provider": "anthropic", "amount_cents": 450},
		},
		"optimization_suggestions": []gin.H{
			{
				"type":                   "provider_switch",
				"title":                  "Switch to Anthropic for text analysis",
				"potential_saving_cents": 200,
			},
		},
	}

	c.JSON(200, analytics)
}

func (s *HTTPServer) createBudget(c *gin.Context) {
	var req cas.Budget
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Demo implementation
	req.ID = uuid.New()
	req.CreatedAt = time.Now()

	c.JSON(201, req)
}

// Security & Compliance endpoints
func (s *HTTPServer) redactContent(c *gin.Context) {
	var req struct {
		Content interface{} `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	redactor := scl.NewRedactor()
	redacted, redactionMap, err := redactor.Redact(req.Content)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"redacted_content": redacted,
		"redaction_map":    redactionMap,
		"stats":            redactor.GetRedactionStats(redactionMap),
	})
}

func (s *HTTPServer) unredactContent(c *gin.Context) {
	var req struct {
		Content      string            `json:"content"`
		RedactionMap map[string]string `json:"redaction_map"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	redactor := scl.NewRedactor()
	unredacted := redactor.UnredactString(req.Content, req.RedactionMap)

	c.JSON(200, gin.H{
		"unredacted_content": unredacted,
	})
}

// System status
func (s *HTTPServer) getSystemStatus(c *gin.Context) {
	status := gin.H{
		"status": "operational",
		"services": gin.H{
			"control_plane": "healthy",
			"workers":       "2 active",
			"database":      "connected",
			"cache":         "connected",
			"messaging":     "connected",
		},
		"metrics": gin.H{
			"active_workflows": 3,
			"completed_today":  15,
			"total_cost_cents": 1250,
		},
		"timestamp": time.Now(),
	}

	c.JSON(200, status)
}
