package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Manage prompts and deployments",
	Long:  "Create, version, evaluate, and deploy prompts",
}

var promptCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new prompt version",
	Args:  cobra.ExactArgs(1),
	RunE:  runPromptCreate,
}

var promptListCmd = &cobra.Command{
	Use:   "list",
	Short: "List prompts",
	RunE:  runPromptList,
}

var promptGetCmd = &cobra.Command{
	Use:   "get [name] [version]",
	Short: "Get a specific prompt version",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runPromptGet,
}

var promptEvalCmd = &cobra.Command{
	Use:   "eval [name] [version]",
	Short: "Evaluate a prompt version",
	Args:  cobra.ExactArgs(2),
	RunE:  runPromptEval,
}

var promptDeployCmd = &cobra.Command{
	Use:   "deploy [name]",
	Short: "Deploy a prompt version",
	Args:  cobra.ExactArgs(1),
	RunE:  runPromptDeploy,
}

var promptTestCmd = &cobra.Command{
	Use:   "test [name]",
	Short: "Test a prompt with sample inputs",
	Args:  cobra.ExactArgs(1),
	RunE:  runPromptTest,
}

func init() {
	// Create command flags
	promptCreateCmd.Flags().StringP("template", "t", "", "Prompt template content")
	promptCreateCmd.Flags().StringP("file", "f", "", "Read template from file")
	promptCreateCmd.Flags().StringP("schema", "s", "", "Input schema as JSON")
	promptCreateCmd.Flags().StringP("schema-file", "", "", "Read schema from file")
	promptCreateCmd.Flags().StringToStringP("metadata", "m", nil, "Metadata as key=value pairs")

	// List command flags
	promptListCmd.Flags().IntP("limit", "l", 20, "Number of results to return")
	promptListCmd.Flags().StringP("filter", "f", "", "Filter by name pattern")

	// Get command flags
	promptGetCmd.Flags().StringP("output", "o", "yaml", "Output format (yaml, json)")

	// Eval command flags
	promptEvalCmd.Flags().StringP("suite", "s", "", "Evaluation suite name")
	promptEvalCmd.Flags().IntP("parallel", "p", 1, "Parallel evaluation workers")
	promptEvalCmd.Flags().BoolP("wait", "w", false, "Wait for evaluation completion")

	// Deploy command flags
	promptDeployCmd.Flags().IntP("stable", "", 0, "Stable version to deploy")
	promptDeployCmd.Flags().IntP("canary", "", 0, "Canary version to deploy")
	promptDeployCmd.Flags().Float64P("canary-ratio", "", 0.0, "Canary traffic ratio (0.0-1.0)")

	// Test command flags
	promptTestCmd.Flags().StringP("inputs", "i", "{}", "Test inputs as JSON")
	promptTestCmd.Flags().StringP("inputs-file", "f", "", "Read inputs from file")
	promptTestCmd.Flags().IntP("version", "v", 0, "Prompt version to test (0 for latest)")

	// Add subcommands
	promptCmd.AddCommand(promptCreateCmd)
	promptCmd.AddCommand(promptListCmd)
	promptCmd.AddCommand(promptGetCmd)
	promptCmd.AddCommand(promptEvalCmd)
	promptCmd.AddCommand(promptDeployCmd)
	promptCmd.AddCommand(promptTestCmd)
}

func runPromptCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	
	// Get template content
	template, _ := cmd.Flags().GetString("template")
	templateFile, _ := cmd.Flags().GetString("file")
	
	if templateFile != "" {
		data, err := os.ReadFile(templateFile)
		if err != nil {
			return fmt.Errorf("failed to read template file: %w", err)
		}
		template = string(data)
	}

	if template == "" {
		return fmt.Errorf("template content is required (use --template or --file)")
	}

	// Get schema
	var schema map[string]interface{}
	schemaStr, _ := cmd.Flags().GetString("schema")
	schemaFile, _ := cmd.Flags().GetString("schema-file")
	
	if schemaFile != "" {
		data, err := os.ReadFile(schemaFile)
		if err != nil {
			return fmt.Errorf("failed to read schema file: %w", err)
		}
		if err := json.Unmarshal(data, &schema); err != nil {
			return fmt.Errorf("failed to parse schema file: %w", err)
		}
	} else if schemaStr != "" {
		if err := json.Unmarshal([]byte(schemaStr), &schema); err != nil {
			return fmt.Errorf("failed to parse schema: %w", err)
		}
	}

	metadata, _ := cmd.Flags().GetStringToString("metadata")

	// Mock creation
	version := 1 // Would get next version from API
	promptID := fmt.Sprintf("prompt_%d", time.Now().Unix())

	fmt.Printf("Created prompt: %s\n", name)
	fmt.Printf("Version: %d\n", version)
	fmt.Printf("ID: %s\n", promptID)
	fmt.Printf("Template length: %d characters\n", len(template))
	
	if len(schema) > 0 {
		fmt.Printf("Schema: %d properties\n", len(schema))
	}
	
	if len(metadata) > 0 {
		fmt.Printf("Metadata: %v\n", metadata)
	}

	return nil
}

func runPromptList(cmd *cobra.Command, args []string) error {
	limit, _ := cmd.Flags().GetInt("limit")
	filter, _ := cmd.Flags().GetString("filter")

	fmt.Printf("Listing prompts (limit: %d, filter: %s)\n", limit, filter)

	// Mock prompt list
	prompts := []map[string]interface{}{
		{
			"name":         "document_analyzer",
			"latest_version": 3,
			"created_at":   time.Now().Add(-7 * 24 * time.Hour).Format(time.RFC3339),
			"updated_at":   time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			"deployed":     true,
		},
		{
			"name":         "content_generator",
			"latest_version": 2,
			"created_at":   time.Now().Add(-5 * 24 * time.Hour).Format(time.RFC3339),
			"updated_at":   time.Now().Add(-1 * 24 * time.Hour).Format(time.RFC3339),
			"deployed":     false,
		},
		{
			"name":         "data_extractor",
			"latest_version": 1,
			"created_at":   time.Now().Add(-3 * 24 * time.Hour).Format(time.RFC3339),
			"updated_at":   time.Now().Add(-3 * 24 * time.Hour).Format(time.RFC3339),
			"deployed":     true,
		},
	}

	// Print table header
	fmt.Printf("%-20s %-8s %-12s %-20s %-10s\n", "NAME", "VERSION", "DEPLOYED", "UPDATED", "STATUS")
	fmt.Println("------------------------------------------------------------------------")

	// Print prompts
	for _, prompt := range prompts {
		if filter != "" && !contains(prompt["name"].(string), filter) {
			continue
		}

		updatedAt, _ := time.Parse(time.RFC3339, prompt["updated_at"].(string))
		deployed := "No"
		if prompt["deployed"].(bool) {
			deployed = "Yes"
		}

		fmt.Printf("%-20s %-8d %-12s %-20s %-10s\n",
			prompt["name"].(string),
			prompt["latest_version"].(int),
			deployed,
			updatedAt.Format("2006-01-02 15:04"),
			"Active",
		)
	}

	return nil
}

func runPromptGet(cmd *cobra.Command, args []string) error {
	name := args[0]
	version := "latest"
	if len(args) > 1 {
		version = args[1]
	}
	_ = version

	output, _ := cmd.Flags().GetString("output")

	// Mock prompt data
	prompt := map[string]interface{}{
		"name":     name,
		"version":  1,
		"template": "Analyze the following document and extract key insights:\n\n{{document}}\n\nProvide a summary with:\n1. Main topics\n2. Key findings\n3. Recommendations",
		"schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"document": map[string]interface{}{
					"type":        "string",
					"description": "The document content to analyze",
				},
			},
			"required": []string{"document"},
		},
		"metadata": map[string]interface{}{
			"author":      "data-team",
			"description": "Document analysis prompt for extracting insights",
			"tags":        []string{"analysis", "document", "insights"},
		},
		"created_at": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
	}

	var outputBytes []byte
	var err error

	switch output {
	case "json":
		outputBytes, err = json.MarshalIndent(prompt, "", "  ")
	case "yaml":
		// Mock YAML output
		outputBytes = []byte(fmt.Sprintf(`name: %s
version: %d
template: |
  %s
schema:
  type: object
  properties:
    document:
      type: string
      description: The document content to analyze
  required:
    - document
metadata:
  author: data-team
  description: Document analysis prompt for extracting insights
  tags:
    - analysis
    - document
    - insights
created_at: %s`,
			prompt["name"], prompt["version"], prompt["template"], prompt["created_at"]))
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}

	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Println(string(outputBytes))
	return nil
}

func runPromptEval(cmd *cobra.Command, args []string) error {
	name := args[0]
	version := args[1]
	
	suite, _ := cmd.Flags().GetString("suite")
	parallel, _ := cmd.Flags().GetInt("parallel")
	wait, _ := cmd.Flags().GetBool("wait")

	if suite == "" {
		return fmt.Errorf("evaluation suite is required (use --suite)")
	}

	fmt.Printf("Starting evaluation of prompt: %s@%s\n", name, version)
	fmt.Printf("Suite: %s\n", suite)
	fmt.Printf("Parallel workers: %d\n", parallel)

	evalID := fmt.Sprintf("eval_%d", time.Now().Unix())
	fmt.Printf("Evaluation ID: %s\n", evalID)

	if wait {
		fmt.Println("Running evaluation...")
		// Mock evaluation progress
		for i := 1; i <= 5; i++ {
			time.Sleep(1 * time.Second)
			fmt.Printf("Progress: %d/5 test cases completed\n", i)
		}

		// Mock results
		fmt.Println("\nEvaluation Results:")
		fmt.Println("==================")
		fmt.Printf("Total cases: 5\n")
		fmt.Printf("Passed: 4\n")
		fmt.Printf("Failed: 1\n")
		fmt.Printf("Success rate: 80.0%%\n")
		fmt.Printf("Average score: 0.85\n")
		fmt.Printf("Total cost: $2.50\n")
	} else {
		fmt.Printf("Use 'agentctl prompt eval-status %s' to check progress\n", evalID)
	}

	return nil
}

func runPromptDeploy(cmd *cobra.Command, args []string) error {
	name := args[0]
	
	stable, _ := cmd.Flags().GetInt("stable")
	canary, _ := cmd.Flags().GetInt("canary")
	canaryRatio, _ := cmd.Flags().GetFloat64("canary-ratio")

	if stable == 0 {
		return fmt.Errorf("stable version is required (use --stable)")
	}

	fmt.Printf("Deploying prompt: %s\n", name)
	fmt.Printf("Stable version: %d\n", stable)
	
	if canary > 0 {
		fmt.Printf("Canary version: %d\n", canary)
		fmt.Printf("Canary ratio: %.1f%%\n", canaryRatio*100)
	}

	// Mock deployment
	time.Sleep(1 * time.Second)
	fmt.Printf("Deployment successful!\n")
	
	if canary > 0 {
		fmt.Printf("Traffic split: %.1f%% stable, %.1f%% canary\n", 
			(1-canaryRatio)*100, canaryRatio*100)
	}

	return nil
}

func runPromptTest(cmd *cobra.Command, args []string) error {
	name := args[0]
	
	// Parse inputs
	var inputs map[string]interface{}
	inputsStr, _ := cmd.Flags().GetString("inputs")
	inputsFile, _ := cmd.Flags().GetString("inputs-file")
	version, _ := cmd.Flags().GetInt("version")
	
	if inputsFile != "" {
		data, err := os.ReadFile(inputsFile)
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

	fmt.Printf("Testing prompt: %s", name)
	if version > 0 {
		fmt.Printf("@%d", version)
	}
	fmt.Println()

	fmt.Printf("Inputs: %v\n", inputs)
	
	// Mock test execution
	fmt.Println("Executing prompt...")
	time.Sleep(2 * time.Second)

	// Mock response
	fmt.Println("\nResponse:")
	fmt.Println("=========")
	fmt.Println("Based on the provided document, here are the key insights:")
	fmt.Println()
	fmt.Println("1. Main topics:")
	fmt.Println("   - Market analysis and trends")
	fmt.Println("   - Customer behavior patterns")
	fmt.Println("   - Competitive landscape")
	fmt.Println()
	fmt.Println("2. Key findings:")
	fmt.Println("   - 25% increase in customer engagement")
	fmt.Println("   - Shift towards digital channels")
	fmt.Println("   - Price sensitivity in target segment")
	fmt.Println()
	fmt.Println("3. Recommendations:")
	fmt.Println("   - Invest in digital marketing")
	fmt.Println("   - Develop competitive pricing strategy")
	fmt.Println("   - Focus on customer retention")
	fmt.Println()
	fmt.Printf("Tokens used: 150 prompt + 200 completion = 350 total\n")
	fmt.Printf("Cost: $0.05\n")
	fmt.Printf("Latency: 1.2s\n")

	return nil
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(substr) > 0 && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}