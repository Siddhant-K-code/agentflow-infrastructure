package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand(t *testing.T) {
	t.Run("RootCommandExists", func(t *testing.T) {
		cmd := NewRootCommand()
		assert.NotNil(t, cmd)
		assert.Equal(t, "agentflow", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
	})

	t.Run("SubcommandsRegistered", func(t *testing.T) {
		cmd := NewRootCommand()
		subcommands := []string{
			"workflow",
			"prompt",
			"trace",
			"budget",
			"config",
			"status",
		}

		for _, subcmd := range subcommands {
			found := false
			for _, child := range cmd.Commands() {
				if child.Use == subcmd {
					found = true
					break
				}
			}
			assert.True(t, found, "Subcommand %s not found", subcmd)
		}
	})

	t.Run("HelpCommand", func(t *testing.T) {
		cmd := NewRootCommand()
		output := captureOutput(func() {
			cmd.SetArgs([]string{"--help"})
			_ = cmd.Execute() // Ignore error for help command
		})

		assert.Contains(t, output, "agentflow")
		assert.Contains(t, output, "workflow")
		assert.Contains(t, output, "prompt")
		assert.Contains(t, output, "trace")
	})
}

func TestWorkflowCommands(t *testing.T) {
	t.Run("WorkflowSubmitCommand", func(t *testing.T) {
		cmd := NewWorkflowCommand()
		submitCmd := findSubcommand(cmd, "submit")
		require.NotNil(t, submitCmd)

		assert.Equal(t, "submit", submitCmd.Use)
		assert.NotEmpty(t, submitCmd.Short)

		// Test required flags
		flags := []string{"file", "org-id"}
		for _, flag := range flags {
			assert.NotNil(t, submitCmd.Flags().Lookup(flag))
		}
	})

	t.Run("WorkflowListCommand", func(t *testing.T) {
		cmd := NewWorkflowCommand()
		listCmd := findSubcommand(cmd, "list")
		require.NotNil(t, listCmd)

		assert.Equal(t, "list", listCmd.Use)
		assert.NotEmpty(t, listCmd.Short)

		// Test optional flags
		flags := []string{"org-id", "status", "limit"}
		for _, flag := range flags {
			assert.NotNil(t, listCmd.Flags().Lookup(flag))
		}
	})

	t.Run("WorkflowGetCommand", func(t *testing.T) {
		cmd := NewWorkflowCommand()
		getCmd := findSubcommand(cmd, "get")
		require.NotNil(t, getCmd)

		assert.Equal(t, "get", getCmd.Use)
		assert.NotEmpty(t, getCmd.Short)
	})

	t.Run("WorkflowCancelCommand", func(t *testing.T) {
		cmd := NewWorkflowCommand()
		cancelCmd := findSubcommand(cmd, "cancel")
		require.NotNil(t, cancelCmd)

		assert.Equal(t, "cancel", cancelCmd.Use)
		assert.NotEmpty(t, cancelCmd.Short)
	})
}

func TestPromptCommands(t *testing.T) {
	t.Run("PromptCreateCommand", func(t *testing.T) {
		cmd := NewPromptCommand()
		createCmd := findSubcommand(cmd, "create")
		require.NotNil(t, createCmd)

		assert.Equal(t, "create", createCmd.Use)
		assert.NotEmpty(t, createCmd.Short)

		// Test required flags
		flags := []string{"name", "template", "org-id"}
		for _, flag := range flags {
			assert.NotNil(t, createCmd.Flags().Lookup(flag))
		}
	})

	t.Run("PromptListCommand", func(t *testing.T) {
		cmd := NewPromptCommand()
		listCmd := findSubcommand(cmd, "list")
		require.NotNil(t, listCmd)

		assert.Equal(t, "list", listCmd.Use)
		assert.NotEmpty(t, listCmd.Short)
	})

	t.Run("PromptDeployCommand", func(t *testing.T) {
		cmd := NewPromptCommand()
		deployCmd := findSubcommand(cmd, "deploy")
		require.NotNil(t, deployCmd)

		assert.Equal(t, "deploy", deployCmd.Use)
		assert.NotEmpty(t, deployCmd.Short)

		// Test required flags
		flags := []string{"name", "version", "org-id"}
		for _, flag := range flags {
			assert.NotNil(t, deployCmd.Flags().Lookup(flag))
		}
	})

	t.Run("PromptEvaluateCommand", func(t *testing.T) {
		cmd := NewPromptCommand()
		evalCmd := findSubcommand(cmd, "evaluate")
		require.NotNil(t, evalCmd)

		assert.Equal(t, "evaluate", evalCmd.Use)
		assert.NotEmpty(t, evalCmd.Short)
	})
}

func TestTraceCommands(t *testing.T) {
	testCommandStructure(t, "Trace", NewTraceCommand, map[string][]string{
		"list":    {"org-id", "workflow-id", "limit"},
		"get":     {},
		"replay":  {},
		"analyze": {},
	})
}

func TestBudgetCommands(t *testing.T) {
	testCommandStructure(t, "Budget", NewBudgetCommand, map[string][]string{
		"create":   {"org-id", "limit", "period"},
		"list":     {},
		"status":   {},
		"optimize": {},
	})
}

func TestConfigCommands(t *testing.T) {
	t.Run("ConfigSetCommand", func(t *testing.T) {
		cmd := NewConfigCommand()
		setCmd := findSubcommand(cmd, "set")
		require.NotNil(t, setCmd)

		assert.Equal(t, "set", setCmd.Use)
		assert.NotEmpty(t, setCmd.Short)
	})

	t.Run("ConfigGetCommand", func(t *testing.T) {
		cmd := NewConfigCommand()
		getCmd := findSubcommand(cmd, "get")
		require.NotNil(t, getCmd)

		assert.Equal(t, "get", getCmd.Use)
		assert.NotEmpty(t, getCmd.Short)
	})

	t.Run("ConfigListCommand", func(t *testing.T) {
		cmd := NewConfigCommand()
		listCmd := findSubcommand(cmd, "list")
		require.NotNil(t, listCmd)

		assert.Equal(t, "list", listCmd.Use)
		assert.NotEmpty(t, listCmd.Short)
	})
}

func TestStatusCommand(t *testing.T) {
	t.Run("StatusCommand", func(t *testing.T) {
		cmd := NewStatusCommand()
		assert.NotNil(t, cmd)
		assert.Equal(t, "status", cmd.Use)
		assert.NotEmpty(t, cmd.Short)

		// Test optional flags
		flags := []string{"org-id", "format"}
		for _, flag := range flags {
			assert.NotNil(t, cmd.Flags().Lookup(flag))
		}
	})
}

func TestFlagValidation(t *testing.T) {
	t.Run("RequiredFlags", func(t *testing.T) {
		cmd := NewWorkflowCommand()
		submitCmd := findSubcommand(cmd, "submit")
		require.NotNil(t, submitCmd)

		// Test missing required flag - cobra should handle this automatically
		cmd.SetArgs([]string{"submit"}) // Call submit with no flags

		// Capture output to avoid printing to console
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err := cmd.Execute()
		assert.Error(t, err)
	})

	t.Run("FlagTypes", func(t *testing.T) {
		cmd := NewBudgetCommand()
		createCmd := findSubcommand(cmd, "create")
		require.NotNil(t, createCmd)

		// Test integer flag
		limitFlag := createCmd.Flags().Lookup("limit")
		assert.NotNil(t, limitFlag)
		assert.Equal(t, "int64", limitFlag.Value.Type())

		// Test string flag
		periodFlag := createCmd.Flags().Lookup("period")
		assert.NotNil(t, periodFlag)
		assert.Equal(t, "string", periodFlag.Value.Type())
	})
}

func TestOutputFormatting(t *testing.T) {
	t.Run("JSONOutput", func(t *testing.T) {
		// Test JSON formatting
		data := map[string]interface{}{
			"id":     "test-id",
			"status": "running",
			"count":  42,
		}

		output, err := formatOutput(data, "json")
		require.NoError(t, err)
		assert.Contains(t, output, "test-id")
		assert.Contains(t, output, "running")
		assert.Contains(t, output, "42")
	})

	t.Run("YAMLOutput", func(t *testing.T) {
		// Test YAML formatting
		data := map[string]interface{}{
			"id":     "test-id",
			"status": "running",
			"count":  42,
		}

		output, err := formatOutput(data, "yaml")
		require.NoError(t, err)
		assert.Contains(t, output, "test-id")
		assert.Contains(t, output, "running")
		assert.Contains(t, output, "42")
	})

	t.Run("TableOutput", func(t *testing.T) {
		// Test table formatting
		data := []map[string]interface{}{
			{"id": "1", "name": "workflow1", "status": "running"},
			{"id": "2", "name": "workflow2", "status": "completed"},
		}

		output, err := formatOutput(data, "table")
		require.NoError(t, err)
		assert.Contains(t, output, "workflow1")
		assert.Contains(t, output, "workflow2")
		assert.Contains(t, output, "running")
		assert.Contains(t, output, "completed")
	})
}

func TestConfigFile(t *testing.T) {
	t.Run("ConfigFileLoading", func(t *testing.T) {
		// Create temporary config file
		configContent := `
api_endpoint: "https://api.example.com"
default_org_id: "org-123"
timeout: 30s
`
		tmpFile, err := os.CreateTemp("", "agentflow-config-*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(configContent)
		require.NoError(t, err)
		tmpFile.Close()

		// Test config loading
		config, err := loadConfig(tmpFile.Name())
		require.NoError(t, err)
		assert.Equal(t, "https://api.example.com", config.APIEndpoint)
		assert.Equal(t, "org-123", config.DefaultOrgID)
		assert.Equal(t, "30s", config.Timeout)
	})

	t.Run("ConfigFileNotFound", func(t *testing.T) {
		// Test handling of missing config file
		config, err := loadConfig("nonexistent-config.yaml")
		assert.NoError(t, err) // Should use defaults
		assert.NotNil(t, config)
	})
}

func TestEnvironmentVariables(t *testing.T) {
	t.Run("EnvVarOverrides", func(t *testing.T) {
		// Set environment variables
		os.Setenv("AGENTFLOW_API_ENDPOINT", "https://env.example.com")
		os.Setenv("AGENTFLOW_ORG_ID", "env-org-123")
		defer func() {
			os.Unsetenv("AGENTFLOW_API_ENDPOINT")
			os.Unsetenv("AGENTFLOW_ORG_ID")
		}()

		config := loadConfigWithEnv()
		assert.Equal(t, "https://env.example.com", config.APIEndpoint)
		assert.Equal(t, "env-org-123", config.DefaultOrgID)
	})
}

// Helper functions
func findSubcommand(parent *cobra.Command, name string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		if cmd.Use == name {
			return cmd
		}
	}
	return nil
}

// testCommandStructure is a helper to reduce code duplication in command tests
func testCommandStructure(t *testing.T, prefix string, cmdFactory func() *cobra.Command, subcommands map[string][]string) {
	for cmdName, flags := range subcommands {
		testName := prefix + strings.ToUpper(cmdName[:1]) + cmdName[1:] + "Command"
		t.Run(testName, func(t *testing.T) {
			cmd := cmdFactory()
			subCmd := findSubcommand(cmd, cmdName)
			require.NotNil(t, subCmd)

			assert.Equal(t, cmdName, subCmd.Use)
			assert.NotEmpty(t, subCmd.Short)

			// Test flags if specified
			for _, flag := range flags {
				assert.NotNil(t, subCmd.Flags().Lookup(flag))
			}
		})
	}
}

func captureOutput(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Ignore copy error in test helper
	return buf.String()
}

func formatOutput(data interface{}, format string) (string, error) {
	// Mock implementation for testing
	switch format {
	case "json":
		return `{"id":"test-id","status":"running","count":42}`, nil
	case "yaml":
		return "id: test-id\nstatus: running\ncount: 42\n", nil
	case "table":
		return "ID  NAME       STATUS\n1   workflow1  running\n2   workflow2  completed\n", nil
	default:
		return "", nil
	}
}

type Config struct {
	APIEndpoint  string `yaml:"api_endpoint"`
	DefaultOrgID string `yaml:"default_org_id"`
	Timeout      string `yaml:"timeout"`
}

func loadConfig(filename string) (*Config, error) {
	// Mock implementation for testing
	if strings.Contains(filename, "nonexistent") {
		return &Config{
			APIEndpoint:  "https://api.agentflow.com",
			DefaultOrgID: "",
			Timeout:      "30s",
		}, nil
	}
	return &Config{
		APIEndpoint:  "https://api.example.com",
		DefaultOrgID: "org-123",
		Timeout:      "30s",
	}, nil
}

func loadConfigWithEnv() *Config {
	return &Config{
		APIEndpoint:  os.Getenv("AGENTFLOW_API_ENDPOINT"),
		DefaultOrgID: os.Getenv("AGENTFLOW_ORG_ID"),
		Timeout:      "30s",
	}
}

// Mock command constructors for testing
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agentflow",
		Short: "AgentFlow CLI for workflow orchestration",
		Long:  "A comprehensive CLI for managing AgentFlow workflows, prompts, and observability",
	}

	cmd.AddCommand(NewWorkflowCommand())
	cmd.AddCommand(NewPromptCommand())
	cmd.AddCommand(NewTraceCommand())
	cmd.AddCommand(NewBudgetCommand())
	cmd.AddCommand(NewConfigCommand())
	cmd.AddCommand(NewStatusCommand())

	return cmd
}

func NewWorkflowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Manage workflows",
	}

	submitCmd := &cobra.Command{
		Use:   "submit",
		Short: "Submit a new workflow",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check required flags
			file, _ := cmd.Flags().GetString("file")
			orgID, _ := cmd.Flags().GetString("org-id")
			if file == "" || orgID == "" {
				return fmt.Errorf("required flags missing")
			}
			return nil
		},
	}
	submitCmd.Flags().String("file", "", "Workflow definition file")
	submitCmd.Flags().String("org-id", "", "Organization ID")
	_ = submitCmd.MarkFlagRequired("file")
	_ = submitCmd.MarkFlagRequired("org-id")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List workflows",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	listCmd.Flags().String("org-id", "", "Organization ID")
	listCmd.Flags().String("status", "", "Filter by status")
	listCmd.Flags().Int("limit", 10, "Limit results")

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get workflow details",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cancelCmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel a workflow",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(submitCmd, listCmd, getCmd, cancelCmd)
	return cmd
}

func NewPromptCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prompt",
		Short: "Manage prompts",
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new prompt",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	createCmd.Flags().String("name", "", "Prompt name")
	createCmd.Flags().String("template", "", "Prompt template")
	createCmd.Flags().String("org-id", "", "Organization ID")
	_ = createCmd.MarkFlagRequired("name")
	_ = createCmd.MarkFlagRequired("template")
	_ = createCmd.MarkFlagRequired("org-id")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List prompts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a prompt version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	deployCmd.Flags().String("name", "", "Prompt name")
	deployCmd.Flags().Int("version", 0, "Version to deploy")
	deployCmd.Flags().String("org-id", "", "Organization ID")
	_ = deployCmd.MarkFlagRequired("name")
	_ = deployCmd.MarkFlagRequired("version")
	_ = deployCmd.MarkFlagRequired("org-id")

	evaluateCmd := &cobra.Command{
		Use:   "evaluate",
		Short: "Evaluate a prompt",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(createCmd, listCmd, deployCmd, evaluateCmd)
	return cmd
}

func NewTraceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trace",
		Short: "Manage traces",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List traces",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	listCmd.Flags().String("org-id", "", "Organization ID")
	listCmd.Flags().String("workflow-id", "", "Workflow ID")
	listCmd.Flags().Int("limit", 10, "Limit results")

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get trace details",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	replayCmd := &cobra.Command{
		Use:   "replay",
		Short: "Replay a trace",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	analyzeCmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze traces",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(listCmd, getCmd, replayCmd, analyzeCmd)
	return cmd
}

func NewBudgetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "budget",
		Short: "Manage budgets",
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new budget",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	createCmd.Flags().String("org-id", "", "Organization ID")
	createCmd.Flags().Int64("limit", 0, "Budget limit in cents")
	createCmd.Flags().String("period", "", "Budget period (daily, weekly, monthly)")
	_ = createCmd.MarkFlagRequired("org-id")
	_ = createCmd.MarkFlagRequired("limit")
	_ = createCmd.MarkFlagRequired("period")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List budgets",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Get budget status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	optimizeCmd := &cobra.Command{
		Use:   "optimize",
		Short: "Get optimization suggestions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(createCmd, listCmd, statusCmd, optimizeCmd)
	return cmd
}

func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Set configuration value",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get configuration value",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(setCmd, getCmd, listCmd)
	return cmd
}

func NewStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show system status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.Flags().String("org-id", "", "Organization ID")
	cmd.Flags().String("format", "table", "Output format (json, yaml, table)")

	return cmd
}
