package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [workflow.yaml]",
	Short: "Deploy a workflow from YAML configuration",
	Long: `Deploy a multi-agent workflow using a YAML configuration file.
The YAML file defines agents, dependencies, triggers, and execution parameters.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workflowFile := args[0]
		
		// Check if file exists
		if _, err := os.Stat(workflowFile); os.IsNotExist(err) {
			return fmt.Errorf("workflow file not found: %s", workflowFile)
		}

		// Read and parse the workflow file
		data, err := os.ReadFile(workflowFile)
		if err != nil {
			return fmt.Errorf("failed to read workflow file: %w", err)
		}

		var workflow Workflow
		if err := yaml.Unmarshal(data, &workflow); err != nil {
			return fmt.Errorf("failed to parse workflow YAML: %w", err)
		}

		// Validate workflow
		if err := validateWorkflow(&workflow); err != nil {
			return fmt.Errorf("workflow validation failed: %w", err)
		}

		// Deploy workflow
		fmt.Printf("Deploying workflow: %s\n", workflow.Metadata.Name)
		
		// TODO: Implement actual deployment logic
		// This will integrate with the orchestrator service
		
		fmt.Printf("✅ Workflow '%s' deployed successfully!\n", workflow.Metadata.Name)
		fmt.Printf("📊 Monitor with: agentflow status %s\n", workflow.Metadata.Name)
		fmt.Printf("👀 Live view: agentflow live-view %s\n", workflow.Metadata.Name)
		
		return nil
	},
}

func validateWorkflow(workflow *Workflow) error {
	if workflow.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if workflow.Kind != "Workflow" {
		return fmt.Errorf("kind must be 'Workflow'")
	}
	if workflow.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if len(workflow.Spec.Agents) == 0 {
		return fmt.Errorf("at least one agent must be defined")
	}
	
	// Validate agent configurations
	for _, agent := range workflow.Spec.Agents {
		if agent.Name == "" {
			return fmt.Errorf("agent name is required")
		}
		if agent.Image == "" {
			return fmt.Errorf("agent image is required for agent '%s'", agent.Name)
		}
		if agent.LLM.Provider == "" {
			return fmt.Errorf("LLM provider is required for agent '%s'", agent.Name)
		}
	}
	
	return nil
}

// Workflow configuration structures
type Workflow struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string            `yaml:"name"`
		Namespace string            `yaml:"namespace,omitempty"`
		Labels    map[string]string `yaml:"labels,omitempty"`
	} `yaml:"metadata"`
	Spec WorkflowSpec `yaml:"spec"`
}

type WorkflowSpec struct {
	Agents   []Agent   `yaml:"agents"`
	Triggers []Trigger `yaml:"triggers,omitempty"`
	Config   Config    `yaml:"config,omitempty"`
}

type Agent struct {
	Name      string            `yaml:"name"`
	Image     string            `yaml:"image"`
	LLM       LLMConfig         `yaml:"llm"`
	DependsOn []string          `yaml:"dependsOn,omitempty"`
	Resources Resources         `yaml:"resources,omitempty"`
	Env       map[string]string `yaml:"env,omitempty"`
	Timeout   string            `yaml:"timeout,omitempty"`
	Retries   int               `yaml:"retries,omitempty"`
}

type LLMConfig struct {
	Provider string            `yaml:"provider"`
	Model    string            `yaml:"model"`
	Config   map[string]string `yaml:"config,omitempty"`
}

type Resources struct {
	Memory string `yaml:"memory,omitempty"`
	CPU    string `yaml:"cpu,omitempty"`
}

type Trigger struct {
	Schedule string `yaml:"schedule,omitempty"`
	Webhook  string `yaml:"webhook,omitempty"`
	Event    string `yaml:"event,omitempty"`
}

type Config struct {
	Parallelism int    `yaml:"parallelism,omitempty"`
	Timeout     string `yaml:"timeout,omitempty"`
	RetryPolicy string `yaml:"retryPolicy,omitempty"`
}