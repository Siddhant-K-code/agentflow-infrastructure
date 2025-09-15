package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [workflow-name]",
	Short: "Show status of workflows",
	Long:  `Display the current status of all workflows or a specific workflow.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			return showWorkflowStatus(args[0])
		}
		return listAllWorkflows()
	},
}

func listAllWorkflows() error {
	orchestratorURL := getOrchestratorURL()
	
	resp, err := http.Get(orchestratorURL + "/api/v1/workflows")
	if err != nil {
		return fmt.Errorf("failed to connect to orchestrator at %s: %w", orchestratorURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("orchestrator returned error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Workflows []WorkflowSummary `json:"workflows"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse orchestrator response: %w", err)
	}

	if len(result.Workflows) == 0 {
		fmt.Println("📭 No workflows found")
		fmt.Println("💡 Deploy one with: agentflow deploy workflow.yaml")
		return nil
	}

	// Display workflows in table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tSTARTED\tID")
	
	for _, workflow := range result.Workflows {
		startTime := ""
		if workflow.StartTime != nil {
			startTime = workflow.StartTime.Format("2006-01-02 15:04:05")
		}
		
		status := workflow.Status
		switch status {
		case "running":
			status = "🔄 " + status
		case "completed":
			status = "✅ " + status
		case "failed":
			status = "❌ " + status
		default:
			status = "⏳ " + status
		}
		
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", 
			workflow.Name, status, startTime, workflow.ID[:8]+"...")
	}
	
	w.Flush()
	return nil
}

func showWorkflowStatus(workflowNameOrID string) error {
	orchestratorURL := getOrchestratorURL()
	
	// First try to find workflow by name or ID
	workflows, err := getWorkflowList()
	if err != nil {
		return err
	}

	var targetWorkflow *WorkflowSummary
	for _, w := range workflows {
		if w.Name == workflowNameOrID || w.ID == workflowNameOrID {
			targetWorkflow = &w
			break
		}
	}

	if targetWorkflow == nil {
		return fmt.Errorf("workflow '%s' not found", workflowNameOrID)
	}

	// Get detailed status
	resp, err := http.Get(orchestratorURL + "/api/v1/workflows/" + targetWorkflow.ID + "/status")
	if err != nil {
		return fmt.Errorf("failed to get workflow status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("orchestrator returned error (status %d): %s", resp.StatusCode, string(body))
	}

	var status WorkflowStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return fmt.Errorf("failed to parse workflow status: %w", err)
	}

	// Display detailed status
	fmt.Printf("📋 Workflow: %s\n", targetWorkflow.Name)
	fmt.Printf("🆔 ID: %s\n", status.ID)
	fmt.Printf("📊 Status: %s\n", getStatusEmoji(status.Status))
	fmt.Printf("⏰ Started: %s\n", status.StartTime.Format("2006-01-02 15:04:05"))
	
	if status.EndTime != nil {
		fmt.Printf("🏁 Finished: %s\n", status.EndTime.Format("2006-01-02 15:04:05"))
		duration := status.EndTime.Sub(status.StartTime)
		fmt.Printf("⏱️  Duration: %s\n", duration.Round(time.Second))
	}

	if status.Error != "" {
		fmt.Printf("❌ Error: %s\n", status.Error)
	}

	// Display agent status
	if len(status.Agents) > 0 {
		fmt.Printf("\n🤖 Agents:\n")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "AGENT\tSTATUS\tDURATION\tERROR")
		
		for _, agent := range status.Agents {
			duration := ""
			if agent.EndTime != nil {
				duration = agent.EndTime.Sub(agent.StartTime).Round(time.Second).String()
			} else if agent.Status == "running" {
				duration = time.Since(agent.StartTime).Round(time.Second).String()
			}
			
			errorMsg := ""
			if agent.Error != "" {
				errorMsg = agent.Error
				if len(errorMsg) > 30 {
					errorMsg = errorMsg[:27] + "..."
				}
			}
			
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				agent.Name, getStatusEmoji(agent.Status), duration, errorMsg)
		}
		w.Flush()
	}

	fmt.Printf("\n💡 Live view: agentflow live-view %s\n", targetWorkflow.Name)
	return nil
}

func getWorkflowList() ([]WorkflowSummary, error) {
	orchestratorURL := getOrchestratorURL()
	
	resp, err := http.Get(orchestratorURL + "/api/v1/workflows")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to orchestrator: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Workflows []WorkflowSummary `json:"workflows"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse workflows: %w", err)
	}

	return result.Workflows, nil
}

func getStatusEmoji(status string) string {
	switch status {
	case "running":
		return "🔄 running"
	case "completed":
		return "✅ completed"
	case "failed":
		return "❌ failed"
	case "pending":
		return "⏳ pending"
	default:
		return "❓ " + status
	}
}

type WorkflowSummary struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	StartTime *time.Time `json:"start_time"`
}

type WorkflowStatus struct {
	ID        string           `json:"id"`
	Status    string           `json:"status"`
	Agents    []AgentStatus    `json:"agents"`
	StartTime time.Time        `json:"start_time"`
	EndTime   *time.Time       `json:"end_time"`
	Error     string           `json:"error"`
}

type AgentStatus struct {
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	StartTime time.Time              `json:"start_time"`
	EndTime   *time.Time             `json:"end_time"`
	Output    map[string]interface{} `json:"output"`
	Error     string                 `json:"error"`
}