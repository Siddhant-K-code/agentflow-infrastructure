package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [workflow-name]",
	Short: "Show workflow status and execution details",
	Long: `Display the current status of workflows, including agent states,
execution progress, and recent events.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return listAllWorkflows()
		}
		return showWorkflowStatus(args[0])
	},
}

func listAllWorkflows() error {
	fmt.Println("📋 Active Workflows:")
	fmt.Println()
	
	// TODO: Get actual workflow data from orchestrator
	workflows := []WorkflowStatus{
		{
			Name:      "data-processing-pipeline",
			Status:    "Running",
			Agents:    3,
			Started:   time.Now().Add(-2 * time.Hour),
			LastEvent: "Agent 'data-processor' completed successfully",
		},
		{
			Name:      "content-generation",
			Status:    "Paused",
			Agents:    2,
			Started:   time.Now().Add(-30 * time.Minute),
			LastEvent: "Waiting for manual approval",
		},
	}
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tAGENTS\tSTARTED\tLAST EVENT")
	fmt.Fprintln(w, "----\t------\t------\t-------\t----------")
	
	for _, wf := range workflows {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
			wf.Name,
			wf.Status,
			wf.Agents,
			formatDuration(time.Since(wf.Started)),
			wf.LastEvent)
	}
	
	w.Flush()
	return nil
}

func showWorkflowStatus(workflowName string) error {
	fmt.Printf("🔍 Workflow Status: %s\n", workflowName)
	fmt.Println()
	
	// TODO: Get actual workflow data from orchestrator
	workflow := WorkflowStatus{
		Name:    workflowName,
		Status:  "Running",
		Agents:  3,
		Started: time.Now().Add(-2 * time.Hour),
	}
	
	agents := []AgentStatus{
		{
			Name:     "data-collector",
			Status:   "Completed",
			Started:  time.Now().Add(-2 * time.Hour),
			Finished: time.Now().Add(-90 * time.Minute),
			LLM:      "openai/gpt-4",
			Retries:  0,
		},
		{
			Name:    "data-processor",
			Status:  "Running",
			Started: time.Now().Add(-90 * time.Minute),
			LLM:     "anthropic/claude-3-sonnet",
			Retries: 1,
		},
		{
			Name:   "data-publisher",
			Status: "Pending",
			LLM:    "openai/gpt-3.5-turbo",
		},
	}
	
	fmt.Printf("📊 Overall Status: %s\n", workflow.Status)
	fmt.Printf("⏰ Started: %s ago\n", formatDuration(time.Since(workflow.Started)))
	fmt.Printf("🤖 Total Agents: %d\n", workflow.Agents)
	fmt.Println()
	
	fmt.Println("🔧 Agent Details:")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "AGENT\tSTATUS\tLLM\tSTARTED\tDURATION\tRETRIES")
	fmt.Fprintln(w, "-----\t------\t---\t-------\t--------\t-------")
	
	for _, agent := range agents {
		duration := ""
		if !agent.Started.IsZero() {
			if !agent.Finished.IsZero() {
				duration = formatDuration(agent.Finished.Sub(agent.Started))
			} else {
				duration = formatDuration(time.Since(agent.Started))
			}
		}
		
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\n",
			agent.Name,
			agent.Status,
			agent.LLM,
			formatTime(agent.Started),
			duration,
			agent.Retries)
	}
	
	w.Flush()
	
	fmt.Println()
	fmt.Println("📈 Recent Events:")
	events := []string{
		"2m ago: Agent 'data-processor' started execution",
		"5m ago: Agent 'data-collector' completed successfully",
		"15m ago: Workflow started",
	}
	
	for _, event := range events {
		fmt.Printf("  • %s\n", event)
	}
	
	return nil
}

type WorkflowStatus struct {
	Name      string
	Status    string
	Agents    int
	Started   time.Time
	LastEvent string
}

type AgentStatus struct {
	Name     string
	Status   string
	Started  time.Time
	Finished time.Time
	LLM      string
	Retries  int
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return formatDuration(time.Since(t)) + " ago"
}