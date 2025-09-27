package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/aor"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DemoServer struct {
	controlPlane *aor.ControlPlane
}

func NewDemoServer(cp *aor.ControlPlane) *DemoServer {
	return &DemoServer{controlPlane: cp}
}

func (s *DemoServer) Start(port int) error {
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
		v1.GET("/workflows/runs", s.listWorkflowRuns)

		// System status
		v1.GET("/status", s.getSystemStatus)
		v1.GET("/costs/analytics", s.getCostAnalytics)
		v1.POST("/scl/redact", s.redactContent)
	}

	return r.Run(fmt.Sprintf(":%d", port))
}

// Workflow endpoints
func (s *DemoServer) submitWorkflow(c *gin.Context) {
	var req aor.RunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	run, err := s.controlPlane.SubmitWorkflow(c.Request.Context(), &req)
	if err != nil {
		log.Printf("Error submitting workflow: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Successfully submitted workflow: %s", run.ID)
	c.JSON(201, run)
}

func (s *DemoServer) getWorkflowRun(c *gin.Context) {
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

func (s *DemoServer) listWorkflowRuns(c *gin.Context) {
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
		{
			ID:     uuid.New(),
			Status: aor.RunStatusFailed,
			Metadata: map[string]interface{}{
				"workflow_name": "image_processing",
				"cost_cents":    25,
				"error":         "API rate limit exceeded",
			},
			CreatedAt: time.Now().Add(-15 * time.Minute),
		},
	}

	c.JSON(200, gin.H{"runs": runs})
}

// System status
func (s *DemoServer) getSystemStatus(c *gin.Context) {
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

// Cost analytics
func (s *DemoServer) getCostAnalytics(c *gin.Context) {
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

// Security & Compliance - PII Redaction
func (s *DemoServer) redactContent(c *gin.Context) {
	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Simple PII redaction demo
	redacted := req.Content
	redactionMap := make(map[string]string)

	// This is a simplified version - in production you'd use proper regex

	// Mock redaction for demo
	if len(req.Content) > 50 {
		redacted = "This document contains sensitive information that has been redacted for security purposes. [REDACTED_EMAIL_12345] [REDACTED_PHONE_67890]"
		redactionMap["[REDACTED_EMAIL_12345]"] = "john.doe@example.com"
		redactionMap["[REDACTED_PHONE_67890]"] = "+1-555-123-4567"
	}

	c.JSON(200, gin.H{
		"original_content": req.Content,
		"redacted_content": redacted,
		"redaction_map":    redactionMap,
		"stats": gin.H{
			"emails_redacted": 1,
			"phones_redacted": 1,
		},
	})
}
