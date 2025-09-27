package demo

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Standalone demo server that doesn't require any external services
type StandaloneDemoServer struct {
	workflowRuns map[string]*WorkflowRun
}

type WorkflowRun struct {
	ID             string                 `json:"id"`
	WorkflowID     string                 `json:"workflow_id"`
	WorkflowSpecID string                 `json:"workflow_spec_id"`
	OrgID          string                 `json:"org_id"`
	Status         string                 `json:"status"`
	Input          interface{}            `json:"input"`
	Output         interface{}            `json:"output"`
	StartedAt      time.Time              `json:"started_at"`
	CreatedAt      time.Time              `json:"created_at"`
	CostCents      int                    `json:"cost_cents"`
	Metadata       map[string]interface{} `json:"metadata"`
	Steps          []*StepRun             `json:"steps"`
}

type StepRun struct {
	ID            string                 `json:"id"`
	WorkflowRunID string                 `json:"workflow_run_id"`
	StepID        string                 `json:"step_id"`
	Status        string                 `json:"status"`
	Input         interface{}            `json:"input"`
	Output        interface{}            `json:"output"`
	StartedAt     time.Time              `json:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at"`
	CostCents     int                    `json:"cost_cents"`
	Metadata      map[string]interface{} `json:"metadata"`
}

func NewStandaloneDemoServer() *StandaloneDemoServer {
	return &StandaloneDemoServer{
		workflowRuns: make(map[string]*WorkflowRun),
	}
}

func (s *StandaloneDemoServer) Start(port int) error {
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

	// Routes
	r.GET("/health", s.healthCheck)
	r.POST("/api/v1/workflows/runs", s.submitWorkflow)
	r.GET("/api/v1/workflows/runs", s.listWorkflowRuns)
	r.GET("/api/v1/workflows/runs/:id", s.getWorkflowRun)
	r.GET("/api/v1/costs/analytics", s.getCostAnalytics)
	r.POST("/api/v1/scl/redact", s.redactContent)
	r.POST("/api/v1/scl/unredact", s.unredactContent)
	r.GET("/api/v1/prompts/:name/versions/:version", s.getPromptTemplate)
	r.GET("/api/v1/prompts/:name/versions", s.listPromptVersions)
	r.GET("/", s.serveDashboard)

	log.Printf("🚀 Standalone AgentFlow Demo Server starting on port %d", port)
	log.Printf("📊 Dashboard: http://localhost:%d", port)
	log.Printf("🔗 API Health: http://localhost:%d/health", port)

	return r.Run(fmt.Sprintf(":%d", port))
}

func (s *StandaloneDemoServer) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0-demo",
		"mode":      "standalone",
	})
}

func (s *StandaloneDemoServer) submitWorkflow(c *gin.Context) {
	var req struct {
		WorkflowName    string                 `json:"workflow_name"`
		WorkflowVersion int                    `json:"workflow_version"`
		Inputs          map[string]interface{} `json:"inputs"`
		BudgetCents     int                    `json:"budget_cents"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Create workflow run
	runID := uuid.New().String()
	now := time.Now()

	// Simulate workflow execution
	status := "running"
	if req.WorkflowName == "quick_test" {
		status = "completed" // Quick completion for demo
	}

	workflowRun := &WorkflowRun{
		ID:             runID,
		WorkflowID:     uuid.New().String(),
		WorkflowSpecID: uuid.New().String(),
		OrgID:          uuid.New().String(),
		Status:         status,
		Input:          req.Inputs,
		Output:         map[string]interface{}{"result": "Workflow executed successfully", "processed_at": now.Format(time.RFC3339)},
		StartedAt:      now,
		CreatedAt:      now,
		CostCents:      req.BudgetCents / 2, // Simulate cost
		Metadata: map[string]interface{}{
			"cost_cents":     req.BudgetCents / 2,
			"workflow_name":  req.WorkflowName,
			"execution_time": "2.5s",
		},
		Steps: []*StepRun{
			{
				ID:            uuid.New().String(),
				WorkflowRunID: runID,
				StepID:        "step1",
				Status:        status,
				Input:         req.Inputs,
				Output:        map[string]interface{}{"processed": true},
				StartedAt:     now,
				CompletedAt:   &now,
				CostCents:     req.BudgetCents / 2,
				Metadata:      map[string]interface{}{"step_type": "llm", "model": "gpt-4"},
			},
		},
	}

	// Store the run
	s.workflowRuns[runID] = workflowRun

	log.Printf("✅ Workflow submitted: %s (status: %s)", runID, status)

	c.JSON(200, gin.H{
		"id":      runID,
		"status":  status,
		"message": "Workflow submitted successfully",
	})
}

func (s *StandaloneDemoServer) listWorkflowRuns(c *gin.Context) {
	runs := make([]*WorkflowRun, 0, len(s.workflowRuns))
	for _, run := range s.workflowRuns {
		runs = append(runs, run)
	}

	c.JSON(200, gin.H{"runs": runs})
}

func (s *StandaloneDemoServer) getWorkflowRun(c *gin.Context) {
	runID := c.Param("id")
	run, exists := s.workflowRuns[runID]
	if !exists {
		c.JSON(404, gin.H{"error": "Workflow run not found"})
		return
	}

	c.JSON(200, run)
}

func (s *StandaloneDemoServer) getCostAnalytics(c *gin.Context) {
	// Mock cost analytics
	totalCost := 0
	for _, run := range s.workflowRuns {
		totalCost += run.CostCents
	}

	c.JSON(200, gin.H{
		"total_cost_cents": totalCost,
		"total_cost_usd":   float64(totalCost) / 100,
		"runs_count":       len(s.workflowRuns),
		"average_cost":     float64(totalCost) / float64(len(s.workflowRuns)),
		"breakdown": gin.H{
			"llm_calls": totalCost * 3 / 4,
			"api_calls": totalCost / 8,
			"storage":   totalCost / 8,
			"compute":   totalCost / 8,
		},
	})
}

func (s *StandaloneDemoServer) redactContent(c *gin.Context) {
	var req struct {
		Content interface{} `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Simple redaction for demo
	redacted := req.Content
	redactionMap := make(map[string]string)

	// Mock redaction for demo
	if contentStr, ok := req.Content.(string); ok && len(contentStr) > 50 {
		redacted = "This document contains sensitive information that has been redacted for security purposes. [REDACTED_EMAIL_12345] [REDACTED_PHONE_67890]"
		redactionMap["[REDACTED_EMAIL_12345]"] = "john.doe@example.com"
		redactionMap["[REDACTED_PHONE_67890]"] = "+1-555-123-4567"
	}

	c.JSON(200, gin.H{
		"redacted_content": redacted,
		"redaction_map":    redactionMap,
		"stats": gin.H{
			"emails_redacted": 1,
			"phones_redacted": 1,
		},
	})
}

func (s *StandaloneDemoServer) unredactContent(c *gin.Context) {
	var req struct {
		Content      interface{}       `json:"content"`
		RedactionMap map[string]string `json:"redaction_map"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Simple unredaction for demo
	unredacted := req.Content
	if contentStr, ok := req.Content.(string); ok {
		for placeholder, _ := range req.RedactionMap {
			_ = placeholder // In a real implementation, you'd replace the placeholders
			unredacted = contentStr
		}
	}

	c.JSON(200, gin.H{
		"unredacted_content": unredacted,
		"stats": gin.H{
			"emails_restored": 1,
			"phones_restored": 1,
		},
	})
}

func (s *StandaloneDemoServer) getPromptTemplate(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")

	// Mock prompt template
	template := gin.H{
		"name":       name,
		"version":    version,
		"content":    "You are a helpful AI assistant. Process the following input: {{input}}",
		"variables":  []string{"input"},
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
	}

	c.JSON(200, template)
}

func (s *StandaloneDemoServer) listPromptVersions(c *gin.Context) {
	name := c.Param("name")

	// Mock prompt versions
	versions := []gin.H{
		{
			"version":    "1",
			"content":    "You are a helpful AI assistant. Process the following input: {{input}}",
			"created_at": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			"is_active":  true,
		},
		{
			"version":    "2",
			"content":    "You are an advanced AI assistant. Analyze and process: {{input}}",
			"created_at": time.Now().Add(-12 * time.Hour).Format(time.RFC3339),
			"is_active":  false,
		},
	}

	c.JSON(200, gin.H{
		"name":     name,
		"versions": versions,
	})
}

func (s *StandaloneDemoServer) serveDashboard(c *gin.Context) {
	// Serve the dashboard HTML
	c.Header("Content-Type", "text/html")
	c.String(200, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AgentFlow Demo Dashboard</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; border-radius: 10px; margin-bottom: 30px; }
        .card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); margin-bottom: 20px; }
        .btn { background: #667eea; color: white; border: none; padding: 10px 20px; border-radius: 5px; cursor: pointer; margin: 5px; }
        .btn:hover { background: #5a6fd8; }
        .status { padding: 5px 10px; border-radius: 15px; font-size: 12px; font-weight: bold; }
        .status.completed { background: #d4edda; color: #155724; }
        .status.running { background: #fff3cd; color: #856404; }
        .status.failed { background: #f8d7da; color: #721c24; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
        .metric { text-align: center; }
        .metric-value { font-size: 2em; font-weight: bold; color: #667eea; }
        .metric-label { color: #666; margin-top: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 AgentFlow Demo Dashboard</h1>
            <p>Enterprise AI Agent Orchestration Platform</p>
        </div>

        <div class="grid">
            <div class="card">
                <h3>📊 Quick Stats</h3>
                <div class="metric">
                    <div class="metric-value" id="totalRuns">0</div>
                    <div class="metric-label">Total Workflows</div>
                </div>
                <div class="metric">
                    <div class="metric-value" id="totalCost">$0.00</div>
                    <div class="metric-label">Total Cost</div>
                </div>
            </div>

            <div class="card">
                <h3>🎯 Submit Workflow</h3>
                <button class="btn" onclick="submitQuickTest()">Quick Test</button>
                <button class="btn" onclick="submitDocumentAnalysis()">Document Analysis</button>
                <button class="btn" onclick="submitDataProcessing()">Data Processing</button>
            </div>

            <div class="card">
                <h3>🔒 PII Redaction Demo</h3>
                <button class="btn" onclick="testRedaction()">Test Redaction</button>
                <button class="btn" onclick="testUnredaction()">Test Unredaction</button>
            </div>

            <div class="card">
                <h3>📝 Prompt Management</h3>
                <button class="btn" onclick="getPromptTemplate()">Get Template</button>
                <button class="btn" onclick="listPromptVersions()">List Versions</button>
            </div>
        </div>

        <div class="card">
            <h3>📋 Recent Workflow Runs</h3>
            <div id="workflowRuns">Loading...</div>
        </div>
    </div>

    <script>
        const API_BASE = window.location.origin;

        async function apiCall(endpoint, method = 'GET', data = null) {
            const options = {
                method,
                headers: { 'Content-Type': 'application/json' }
            };
            if (data) options.body = JSON.stringify(data);

            const response = await fetch(API_BASE + endpoint, options);
            return response.json();
        }

        async function loadStats() {
            try {
                const [runs, costs] = await Promise.all([
                    apiCall('/api/v1/workflows/runs'),
                    apiCall('/api/v1/costs/analytics')
                ]);

                document.getElementById('totalRuns').textContent = runs.runs.length;
                document.getElementById('totalCost').textContent = '$' + costs.total_cost_usd.toFixed(2);

                displayWorkflowRuns(runs.runs);
            } catch (error) {
                console.error('Error loading stats:', error);
            }
        }

        function displayWorkflowRuns(runs) {
            const container = document.getElementById('workflowRuns');
            if (runs.length === 0) {
                container.innerHTML = '<p>No workflow runs yet. Submit a workflow to get started!</p>';
                return;
            }

            const html = runs.map(run =>
                '<div style="border: 1px solid #ddd; padding: 15px; margin: 10px 0; border-radius: 5px;">' +
                    '<div style="display: flex; justify-content: space-between; align-items: center;">' +
                        '<strong>' + (run.metadata?.workflow_name || 'Unknown') + '</strong>' +
                        '<span class="status ' + run.status + '">' + run.status.toUpperCase() + '</span>' +
                    '</div>' +
                    '<div style="color: #666; font-size: 0.9em; margin-top: 5px;">' +
                        'ID: ' + run.id + ' | Cost: $' + (run.cost_cents / 100).toFixed(2) + ' | ' +
                        'Created: ' + new Date(run.created_at).toLocaleString() +
                    '</div>' +
                '</div>'
            ).join('');

            container.innerHTML = html;
        }

        async function submitWorkflow(name, inputs, budget = 1000) {
            try {
                const result = await apiCall('/api/v1/workflows/runs', 'POST', {
                    workflow_name: name,
                    workflow_version: 1,
                    inputs: inputs,
                    budget_cents: budget
                });

                alert('Workflow submitted successfully! ID: ' + result.id);
                loadStats();
            } catch (error) {
                alert('Error submitting workflow: ' + error.message);
            }
        }

        function submitQuickTest() {
            submitWorkflow('quick_test', { message: 'Hello AgentFlow!' }, 500);
        }

        function submitDocumentAnalysis() {
            submitWorkflow('document_analysis', { document: 'sample.pdf', type: 'contract' }, 2000);
        }

        function submitDataProcessing() {
            submitWorkflow('data_processing', { dataset: 'sales_data.csv', operation: 'aggregation' }, 1500);
        }

        async function testRedaction() {
            try {
                const result = await apiCall('/api/v1/scl/redact', 'POST', {
                    content: 'Contact John Doe at john.doe@example.com or call +1-555-123-4567 for more information.'
                });

                alert('Redaction completed! Check console for details.');
                console.log('Redaction result:', result);
            } catch (error) {
                alert('Error testing redaction: ' + error.message);
            }
        }

        async function testUnredaction() {
            try {
                const result = await apiCall('/api/v1/scl/unredact', 'POST', {
                    content: 'Contact [REDACTED_EMAIL] or call [REDACTED_PHONE] for more information.',
                    redaction_map: {
                        '[REDACTED_EMAIL]': 'john.doe@example.com',
                        '[REDACTED_PHONE]': '+1-555-123-4567'
                    }
                });

                alert('Unredaction completed! Check console for details.');
                console.log('Unredaction result:', result);
            } catch (error) {
                alert('Error testing unredaction: ' + error.message);
            }
        }

        async function getPromptTemplate() {
            try {
                const result = await apiCall('/api/v1/prompts/assistant/versions/1');
                alert('Prompt template retrieved! Check console for details.');
                console.log('Prompt template:', result);
            } catch (error) {
                alert('Error getting prompt template: ' + error.message);
            }
        }

        async function listPromptVersions() {
            try {
                const result = await apiCall('/api/v1/prompts/assistant/versions');
                alert('Prompt versions retrieved! Check console for details.');
                console.log('Prompt versions:', result);
            } catch (error) {
                alert('Error listing prompt versions: ' + error.message);
            }
        }

        // Load stats on page load
        loadStats();
        setInterval(loadStats, 5000); // Refresh every 5 seconds
    </script>
</body>
</html>`)
}
