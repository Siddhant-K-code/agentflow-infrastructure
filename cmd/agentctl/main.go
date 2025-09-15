package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	serverURL string
	orgID     string
	debug     bool
	apiClient *client.Client
)

func getClient() *client.Client {
	if apiClient == nil {
		if orgID == "" {
			orgID = "00000000-0000-0000-0000-000000000001" // Default org for POC
		}
		apiClient = client.New(serverURL, orgID)
	}
	return apiClient
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "agentctl",
		Short: "AgentFlow CLI for managing workflows and runs",
		Long:  "Command line interface for the AgentFlow multi-agent orchestration platform",
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "http://localhost:8080", "AgentFlow server URL")
	rootCmd.PersistentFlags().StringVar(&orgID, "org", "", "Organization ID")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug output")

	// Add subcommands
	rootCmd.AddCommand(
		newWorkflowCmd(),
		newRunCmd(),
		newPromptCmd(),
		newStatusCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newWorkflowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Manage workflows",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "submit [file]",
			Short: "Submit a workflow specification",
			Args:  cobra.ExactArgs(1),
			RunE:  submitWorkflow,
		},
		&cobra.Command{
			Use:   "list",
			Short: "List workflows",
			RunE:  listWorkflows,
		},
		&cobra.Command{
			Use:   "get [name] [version]",
			Short: "Get workflow specification",
			Args:  cobra.RangeArgs(1, 2),
			RunE:  getWorkflow,
		},
	)

	return cmd
}

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Manage workflow runs",
	}

	var (
		workflowName    string
		workflowVersion int
		budgetCents     int64
		tags            []string
	)

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start a workflow run",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startRun(workflowName, workflowVersion, budgetCents, tags)
		},
	}
	startCmd.Flags().StringVar(&workflowName, "workflow", "", "Workflow name")
	startCmd.Flags().IntVar(&workflowVersion, "version", 0, "Workflow version (0 for latest)")
	startCmd.Flags().Int64Var(&budgetCents, "budget", 0, "Budget in cents")
	startCmd.Flags().StringSliceVar(&tags, "tags", nil, "Tags for the run")
	startCmd.MarkFlagRequired("workflow")

	cmd.AddCommand(
		startCmd,
		&cobra.Command{
			Use:   "status [run-id]",
			Short: "Get run status",
			Args:  cobra.ExactArgs(1),
			RunE:  getRunStatus,
		},
		&cobra.Command{
			Use:   "cancel [run-id]",
			Short: "Cancel a run",
			Args:  cobra.ExactArgs(1),
			RunE:  cancelRun,
		},
		&cobra.Command{
			Use:   "list",
			Short: "List runs", 
			RunE:  listRuns,
		},
	)

	return cmd
}

func newPromptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prompt",
		Short: "Manage prompts",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "create [name] [file]",
			Short: "Create a prompt template",
			Args:  cobra.ExactArgs(2),
			RunE:  createPrompt,
		},
		&cobra.Command{
			Use:   "get [name] [version]",
			Short: "Get prompt template",
			Args:  cobra.RangeArgs(1, 2),
			RunE:  getPrompt,
		},
		&cobra.Command{
			Use:   "evaluate [name] [suite]",
			Short: "Evaluate a prompt",
			Args:  cobra.ExactArgs(2),
			RunE:  evaluatePrompt,
		},
		&cobra.Command{
			Use:   "deploy [name] [version]",
			Short: "Deploy a prompt version",
			Args:  cobra.ExactArgs(2),
			RunE:  deployPrompt,
		},
	)

	return cmd
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show system status",
		RunE:  systemStatus,
	}
}

// Command implementations
func submitWorkflow(cmd *cobra.Command, args []string) error {
	filename := args[0]
	
	// Read and parse workflow file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", filename, err)
	}

	// Parse YAML to extract workflow name and convert to JSON
	var workflow map[string]interface{}
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return fmt.Errorf("failed to parse YAML: %v", err)
	}

	name, ok := workflow["name"].(string)
	if !ok {
		return fmt.Errorf("workflow must have a 'name' field")
	}

	// Convert to JSON for API
	jsonData, err := json.Marshal(workflow)
	if err != nil {
		return fmt.Errorf("failed to convert workflow to JSON: %v", err)
	}

	// Submit to API
	spec, err := getClient().CreateWorkflow(name, jsonData)
	if err != nil {
		return fmt.Errorf("failed to submit workflow: %v", err)
	}

	fmt.Printf("Workflow submitted successfully!\n")
	fmt.Printf("Name: %s\n", spec.Name)
	fmt.Printf("Version: %d\n", spec.Version)
	fmt.Printf("ID: %s\n", spec.ID)
	
	return nil
}

func listWorkflows(cmd *cobra.Command, args []string) error {
	workflows, err := getClient().ListWorkflows()
	if err != nil {
		return fmt.Errorf("failed to list workflows: %v", err)
	}

	if len(workflows) == 0 {
		fmt.Println("No workflows found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tID")
	for _, workflow := range workflows {
		fmt.Fprintf(w, "%s\t%d\t%s\n", workflow.Name, workflow.Version, workflow.ID)
	}
	w.Flush()
	
	return nil
}

func getWorkflow(cmd *cobra.Command, args []string) error {
	name := args[0]
	version := 0
	
	if len(args) > 1 {
		var err error
		version, err = strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid version number: %v", err)
		}
	}

	workflow, err := getClient().GetWorkflow(name, version)
	if err != nil {
		return fmt.Errorf("failed to get workflow: %v", err)
	}

	fmt.Printf("Name: %s\n", workflow.Name)
	fmt.Printf("Version: %d\n", workflow.Version)
	fmt.Printf("ID: %s\n", workflow.ID)
	fmt.Printf("DAG:\n%s\n", string(workflow.DAG))
	
	return nil
}

func startRun(workflowName string, version int, budget int64, tags []string) error {
	req := client.CreateRunRequest{
		WorkflowName:    workflowName,
		WorkflowVersion: version,
		Tags:            tags,
	}
	
	if budget > 0 {
		req.BudgetCents = &budget
	}

	run, err := getClient().CreateRun(req)
	if err != nil {
		return fmt.Errorf("failed to start run: %v", err)
	}

	fmt.Printf("Run started successfully!\n")
	fmt.Printf("Run ID: %s\n", run.ID)
	fmt.Printf("Workflow: %s\n", workflowName)
	fmt.Printf("Status: %s\n", run.Status)
	fmt.Printf("Started: %s\n", run.StartedAt.Format(time.RFC3339))
	
	return nil
}

func getRunStatus(cmd *cobra.Command, args []string) error {
	runIDStr := args[0]
	runID, err := uuid.Parse(runIDStr)
	if err != nil {
		return fmt.Errorf("invalid run ID: %v", err)
	}

	run, err := getClient().GetRun(runID)
	if err != nil {
		return fmt.Errorf("failed to get run status: %v", err)
	}

	fmt.Printf("Run ID: %s\n", run.ID)
	fmt.Printf("Status: %s\n", run.Status)
	fmt.Printf("Started: %s\n", run.StartedAt.Format(time.RFC3339))
	if run.EndedAt != nil {
		fmt.Printf("Ended: %s\n", run.EndedAt.Format(time.RFC3339))
		fmt.Printf("Duration: %s\n", run.EndedAt.Sub(run.StartedAt))
	}
	fmt.Printf("Cost: $%.2f\n", float64(run.CostCents)/100)
	if run.BudgetCents != nil {
		fmt.Printf("Budget: $%.2f\n", float64(*run.BudgetCents)/100)
	}
	if len(run.Tags) > 0 {
		fmt.Printf("Tags: %v\n", run.Tags)
	}
	
	return nil
}

func cancelRun(cmd *cobra.Command, args []string) error {
	runIDStr := args[0]
	runID, err := uuid.Parse(runIDStr)
	if err != nil {
		return fmt.Errorf("invalid run ID: %v", err)
	}

	if err := getClient().CancelRun(runID); err != nil {
		return fmt.Errorf("failed to cancel run: %v", err)
	}

	fmt.Printf("Run %s canceled successfully\n", runID)
	return nil
}

func listRuns(cmd *cobra.Command, args []string) error {
	runs, err := getClient().ListRuns()
	if err != nil {
		return fmt.Errorf("failed to list runs: %v", err)
	}

	if len(runs) == 0 {
		fmt.Println("No runs found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATUS\tSTARTED\tCOST")
	for _, run := range runs {
		cost := fmt.Sprintf("$%.2f", float64(run.CostCents)/100)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", 
			run.ID.String()[:8]+"...", 
			run.Status, 
			run.StartedAt.Format("15:04:05"), 
			cost)
	}
	w.Flush()
	
	return nil
}

func createPrompt(cmd *cobra.Command, args []string) error {
	fmt.Printf("Creating prompt: %s from file: %s\n", args[0], args[1])
	fmt.Println("Note: Prompt management via POP API not yet implemented in CLI")
	return nil
}

func getPrompt(cmd *cobra.Command, args []string) error {
	fmt.Printf("Getting prompt: %s\n", args[0])
	fmt.Println("Note: Prompt management via POP API not yet implemented in CLI")
	return nil
}

func evaluatePrompt(cmd *cobra.Command, args []string) error {
	fmt.Printf("Evaluating prompt: %s with suite: %s\n", args[0], args[1])
	fmt.Println("Note: Prompt evaluation via POP API not yet implemented in CLI")
	return nil
}

func deployPrompt(cmd *cobra.Command, args []string) error {
	fmt.Printf("Deploying prompt: %s version: %s\n", args[0], args[1])
	fmt.Println("Note: Prompt deployment via POP API not yet implemented in CLI")
	return nil
}

func systemStatus(cmd *cobra.Command, args []string) error {
	status, err := getClient().GetSystemStatus()
	if err != nil {
		return fmt.Errorf("failed to get system status: %v", err)
	}

	fmt.Printf("AgentFlow System Status\n")
	fmt.Printf("======================\n")
	fmt.Printf("Overall Status: %s\n", status.Status)
	fmt.Printf("Database: %s\n", status.Database)
	fmt.Printf("Queue: %s\n", status.Queue)
	fmt.Printf("Active Workers: %d\n", status.Workers)
	fmt.Printf("Server: %s\n", serverURL)
	
	return nil
}