package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Manage workflows",
	Long:  "Submit, monitor, and manage workflow executions",
}

var workflowSubmitCmd = &cobra.Command{
	Use:   "submit [workflow-name]",
	Short: "Submit a workflow for execution",
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkflowSubmit,
}

var workflowStatusCmd = &cobra.Command{
	Use:   "status [run-id]",
	Short: "Get workflow run status",
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkflowStatus,
}

var workflowListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workflow runs",
	RunE:  runWorkflowList,
}

var workflowCancelCmd = &cobra.Command{
	Use:   "cancel [run-id]",
	Short: "Cancel a workflow run",
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkflowCancel,
}

var workflowLogsCmd = &cobra.Command{
	Use:   "logs [run-id]",
	Short: "Get workflow run logs",
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkflowLogs,
}

func init() {
	// Submit command flags
	workflowSubmitCmd.Flags().StringP("version", "v", "", "Workflow version (default: latest)")
	workflowSubmitCmd.Flags().StringP("inputs", "i", "{}", "Input parameters as JSON")
	workflowSubmitCmd.Flags().StringP("inputs-file", "f", "", "Input parameters from file")
	workflowSubmitCmd.Flags().Int64P("budget", "b", 0, "Budget limit in cents")
	workflowSubmitCmd.Flags().StringToStringP("tags", "t", nil, "Tags as key=value pairs")
	workflowSubmitCmd.Flags().BoolP("wait", "w", false, "Wait for completion")
	workflowSubmitCmd.Flags().DurationP("timeout", "", 30*time.Minute, "Wait timeout")

	// List command flags
	workflowListCmd.Flags().StringP("status", "s", "", "Filter by status")
	workflowListCmd.Flags().IntP("limit", "l", 20, "Number of results to return")
	workflowListCmd.Flags().StringP("since", "", "24h", "Show runs since duration")

	// Logs command flags
	workflowLogsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
	workflowLogsCmd.Flags().IntP("tail", "t", 100, "Number of recent log lines")

	// Add subcommands
	workflowCmd.AddCommand(workflowSubmitCmd)
	workflowCmd.AddCommand(workflowStatusCmd)
	workflowCmd.AddCommand(workflowListCmd)
	workflowCmd.AddCommand(workflowCancelCmd)
	workflowCmd.AddCommand(workflowLogsCmd)
}

func runWorkflowSubmit(cmd *cobra.Command, args []string) error {
	workflowName := args[0]

	// Parse inputs
	var inputs map[string]interface{}
	inputsStr, _ := cmd.Flags().GetString("inputs")
	inputsFile, _ := cmd.Flags().GetString("inputs-file")

	if inputsFile != "" {
		// Validate file path to prevent directory traversal
		if err := validateFilePath(inputsFile); err != nil {
			return fmt.Errorf("invalid file path: %w", err)
		}
		data, err := os.ReadFile(inputsFile) // #nosec G304 - path validated above
		if err != nil {
			return fmt.Errorf("failed to read inputs file: %w", err)
		}
		if err := json.Unmarshal(data, &inputs); err != nil {
			return fmt.Errorf("failed to parse inputs file: %w", err)
		}
	} else {
		if err := json.Unmarshal([]byte(inputsStr), &inputs); err != nil {
			return fmt.Errorf("failed to parse inputs: %w", err)
		}
	}

	version, _ := cmd.Flags().GetString("version")
	budget, _ := cmd.Flags().GetInt64("budget")
	tags, _ := cmd.Flags().GetStringToString("tags")
	wait, _ := cmd.Flags().GetBool("wait")
	timeout, _ := cmd.Flags().GetDuration("timeout")

	// Create request
	request := map[string]interface{}{
		"workflow_name": workflowName,
		"inputs":        inputs,
		"tags":          tags,
	}

	if version != "" {
		versionInt := 0
		if _, err := fmt.Sscanf(version, "%d", &versionInt); err != nil {
			return fmt.Errorf("invalid version format: %w", err)
		}
		request["workflow_version"] = versionInt
	}

	if budget > 0 {
		request["budget_cents"] = budget
	}

	// Submit workflow (mock implementation)
	runID := "run_" + fmt.Sprintf("%d", time.Now().Unix())

	fmt.Printf("Submitted workflow: %s\n", workflowName)
	fmt.Printf("Run ID: %s\n", runID)

	if wait {
		fmt.Printf("Waiting for completion (timeout: %v)...\n", timeout)
		// Mock waiting
		time.Sleep(2 * time.Second)
		fmt.Printf("Workflow completed successfully\n")
	} else {
		fmt.Printf("Use 'agentctl workflow status %s' to check progress\n", runID)
	}

	return nil
}

func runWorkflowStatus(cmd *cobra.Command, args []string) error {
	runID := args[0]

	// Mock status response
	status := map[string]interface{}{
		"id":         runID,
		"status":     "running",
		"started_at": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
		"progress": map[string]interface{}{
			"completed_steps": 3,
			"total_steps":     5,
			"current_step":    "llm_analysis",
		},
		"cost_cents": 150,
		"metadata": map[string]interface{}{
			"workflow_name": "document_analysis",
			"version":       1,
		},
	}

	output, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format status: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

func runWorkflowList(cmd *cobra.Command, args []string) error {
	statusFilter, _ := cmd.Flags().GetString("status")
	limit, _ := cmd.Flags().GetInt("limit")
	since, _ := cmd.Flags().GetString("since")

	fmt.Printf("Listing workflows (status: %s, limit: %d, since: %s)\n", statusFilter, limit, since)

	// Mock workflow list
	workflows := []map[string]interface{}{
		{
			"id":           "run_1234567890",
			"workflow":     "document_analysis",
			"status":       "completed",
			"started_at":   time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			"completed_at": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			"cost_cents":   250,
		},
		{
			"id":         "run_1234567891",
			"workflow":   "data_extraction",
			"status":     "running",
			"started_at": time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
			"cost_cents": 120,
		},
		{
			"id":         "run_1234567892",
			"workflow":   "content_generation",
			"status":     "failed",
			"started_at": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			"error":      "Budget exceeded",
			"cost_cents": 500,
		},
	}

	// Print table header
	fmt.Printf("%-20s %-20s %-12s %-20s %-10s\n", "RUN ID", "WORKFLOW", "STATUS", "STARTED", "COST")
	fmt.Println("--------------------------------------------------------------------------------")

	// Print workflows
	for _, wf := range workflows {
		if statusFilter != "" && wf["status"] != statusFilter {
			continue
		}

		startedAt, _ := time.Parse(time.RFC3339, wf["started_at"].(string))
		cost := wf["cost_cents"].(int)

		fmt.Printf("%-20s %-20s %-12s %-20s $%.2f\n",
			wf["id"].(string)[:18]+"...",
			wf["workflow"].(string),
			wf["status"].(string),
			startedAt.Format("2006-01-02 15:04"),
			float64(cost)/100,
		)
	}

	return nil
}

func runWorkflowCancel(cmd *cobra.Command, args []string) error {
	runID := args[0]

	fmt.Printf("Cancelling workflow run: %s\n", runID)

	// Mock cancellation
	time.Sleep(1 * time.Second)
	fmt.Printf("Workflow run cancelled successfully\n")

	return nil
}

func runWorkflowLogs(cmd *cobra.Command, args []string) error {
	runID := args[0]
	follow, _ := cmd.Flags().GetBool("follow")
	tail, _ := cmd.Flags().GetInt("tail")

	fmt.Printf("Getting logs for workflow run: %s (tail: %d, follow: %t)\n", runID, tail, follow)

	// Mock logs
	logs := []string{
		"2024-01-15 10:00:00 [INFO] Workflow started: document_analysis",
		"2024-01-15 10:00:01 [INFO] Step 1/5: document_ingestion - started",
		"2024-01-15 10:00:05 [INFO] Step 1/5: document_ingestion - completed (cost: $0.10)",
		"2024-01-15 10:00:06 [INFO] Step 2/5: text_extraction - started",
		"2024-01-15 10:00:12 [INFO] Step 2/5: text_extraction - completed (cost: $0.25)",
		"2024-01-15 10:00:13 [INFO] Step 3/5: llm_analysis - started",
		"2024-01-15 10:00:18 [INFO] Step 3/5: llm_analysis - in progress (tokens: 1500)",
	}

	for _, log := range logs {
		fmt.Println(log)
	}

	if follow {
		fmt.Println("Following logs... (Ctrl+C to exit)")
		// Mock following
		for i := 0; i < 3; i++ {
			time.Sleep(2 * time.Second)
			fmt.Printf("2024-01-15 10:00:%02d [INFO] Step 3/5: llm_analysis - processing chunk %d\n", 20+i*2, i+1)
		}
	}

	return nil
}
