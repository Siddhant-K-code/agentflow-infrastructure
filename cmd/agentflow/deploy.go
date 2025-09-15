package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

		// Deploy workflow to orchestrator
		fmt.Printf("🚀 Deploying workflow: %s\n", workflow.Metadata.Name)
		
		if err := deployToOrchestrator(&workflow); err != nil {
			return fmt.Errorf("deployment failed: %w", err)
		}
		
		fmt.Printf("✅ Workflow '%s' deployed successfully!\n", workflow.Metadata.Name)
		fmt.Printf("📊 Monitor with: agentflow status %s\n", workflow.Metadata.Name)
		fmt.Printf("👀 Live view: agentflow live-view %s\n", workflow.Metadata.Name)
		
		return nil
	},
}

func deployToOrchestrator(workflow *Workflow) error {
	orchestratorURL := getOrchestratorURL()
	
	jsonData, err := json.Marshal(workflow)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow: %w", err)
	}

	resp, err := http.Post(orchestratorURL+"/api/v1/workflows", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to connect to orchestrator at %s: %w", orchestratorURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("orchestrator returned error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse orchestrator response: %w", err)
	}

	fmt.Printf("📋 Workflow ID: %s\n", result["id"])
	return nil
}

func getOrchestratorURL() string {
	if url := os.Getenv("AGENTFLOW_ORCHESTRATOR_URL"); url != "" {
		return url
	}
	return "http://localhost:8080"
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
		// LLM provider is optional for POC
	}
	
	return nil
}

// Workflow configuration structures
type Workflow struct {
	APIVersion string `yaml:"apiVersion" json:"apiVersion"`
	Kind       string `yaml:"kind" json:"kind"`
	Metadata   struct {
		Name      string            `yaml:"name" json:"name"`
		Namespace string            `yaml:"namespace,omitempty" json:"namespace,omitempty"`
		Labels    map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	} `yaml:"metadata" json:"metadata"`
	Spec WorkflowSpec `yaml:"spec" json:"spec"`
}

type WorkflowSpec struct {
	Agents   []Agent   `yaml:"agents" json:"agents"`
	Triggers []Trigger `yaml:"triggers,omitempty" json:"triggers,omitempty"`
	Config   Config    `yaml:"config,omitempty" json:"config,omitempty"`
}

type Agent struct {
	Name      string            `yaml:"name" json:"name"`
	Image     string            `yaml:"image" json:"image"`
	LLM       LLMConfig         `yaml:"llm" json:"llm"`
	DependsOn []string          `yaml:"dependsOn,omitempty" json:"dependsOn,omitempty"`
	Resources Resources         `yaml:"resources,omitempty" json:"resources,omitempty"`
	Env       map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	Timeout   string            `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Retries   int               `yaml:"retries,omitempty" json:"retries,omitempty"`
}

type LLMConfig struct {
	Provider string            `yaml:"provider" json:"provider"`
	Model    string            `yaml:"model" json:"model"`
	Config   map[string]string `yaml:"config,omitempty" json:"config,omitempty"`
}

type Resources struct {
	Memory string `yaml:"memory,omitempty" json:"memory,omitempty"`
	CPU    string `yaml:"cpu,omitempty" json:"cpu,omitempty"`
}

type Trigger struct {
	Schedule string `yaml:"schedule,omitempty" json:"schedule,omitempty"`
	Webhook  string `yaml:"webhook,omitempty" json:"webhook,omitempty"`
	Event    string `yaml:"event,omitempty" json:"event,omitempty"`
}

type Config struct {
	Parallelism int    `yaml:"parallelism,omitempty" json:"parallelism,omitempty"`
	Timeout     string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	RetryPolicy string `yaml:"retryPolicy,omitempty" json:"retryPolicy,omitempty"`
}