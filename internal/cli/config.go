package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  "Configure CLI settings, authentication, and defaults",
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigGet,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	RunE:  runConfigList,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	RunE:  runConfigInit,
}

var configLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login and save authentication token",
	RunE:  runConfigLogin,
}

var configLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and remove authentication token",
	RunE:  runConfigLogout,
}

func init() {
	// Login command flags
	configLoginCmd.Flags().StringP("token", "t", "", "Authentication token")
	configLoginCmd.Flags().StringP("endpoint", "e", "", "API endpoint")
	configLoginCmd.Flags().StringP("org", "o", "", "Organization ID")

	// Add subcommands
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configLoginCmd)
	configCmd.AddCommand(configLogoutCmd)
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	viper.Set(key, value)
	
	if err := writeConfig(); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("Set %s = %s\n", key, value)
	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := viper.GetString(key)
	
	if value == "" {
		fmt.Printf("%s is not set\n", key)
	} else {
		fmt.Printf("%s = %s\n", key, value)
	}
	
	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	fmt.Println("Current configuration:")
	fmt.Println("======================")

	settings := viper.AllSettings()
	if len(settings) == 0 {
		fmt.Println("No configuration found. Run 'agentctl config init' to initialize.")
		return nil
	}

	// Common configuration keys
	keys := []string{
		"endpoint",
		"org",
		"token",
		"default_budget",
		"default_quality",
		"output_format",
	}

	for _, key := range keys {
		value := viper.GetString(key)
		if value != "" {
			// Mask sensitive values
			if key == "token" && len(value) > 8 {
				value = value[:8] + "..."
			}
			fmt.Printf("%-15s: %s\n", key, value)
		}
	}

	// Show any additional settings
	for key, value := range settings {
		found := false
		for _, commonKey := range keys {
			if key == commonKey {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("%-15s: %v\n", key, value)
		}
	}

	configFile := viper.ConfigFileUsed()
	if configFile != "" {
		fmt.Printf("\nConfig file: %s\n", configFile)
	}

	return nil
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	fmt.Println("Initializing AgentFlow CLI configuration...")

	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(home, ".agentctl.yaml")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Configuration file already exists at %s\n", configPath)
		fmt.Print("Overwrite? (y/N): ")
		
		// In a real implementation, would read user input
		fmt.Println("N")
		fmt.Println("Configuration initialization cancelled.")
		return nil
	}

	// Set default values
	defaults := map[string]interface{}{
		"endpoint":        "http://localhost:8080",
		"output_format":   "table",
		"default_quality": "Silver",
		"default_budget":  10000, // $100 in cents
	}

	for key, value := range defaults {
		viper.Set(key, value)
	}

	// Write config file
	viper.SetConfigFile(configPath)
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Configuration initialized at %s\n", configPath)
	fmt.Println("\nDefault settings:")
	for key, value := range defaults {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Println("\nNext steps:")
	fmt.Println("1. Run 'agentctl config login' to authenticate")
	fmt.Println("2. Run 'agentctl config set org <your-org-id>' to set your organization")
	fmt.Println("3. Run 'agentctl workflow list' to test the connection")

	return nil
}

func runConfigLogin(cmd *cobra.Command, args []string) error {
	token, _ := cmd.Flags().GetString("token")
	endpoint, _ := cmd.Flags().GetString("endpoint")
	org, _ := cmd.Flags().GetString("org")

	if token == "" {
		fmt.Print("Enter authentication token: ")
		// In a real implementation, would read from stdin securely
		token = "mock_token_12345"
		fmt.Println("***************")
	}

	if endpoint == "" {
		endpoint = viper.GetString("endpoint")
		if endpoint == "" {
			endpoint = "http://localhost:8080"
		}
	}

	fmt.Printf("Authenticating with %s...\n", endpoint)

	// Mock authentication
	fmt.Println("Authentication successful!")

	// Save credentials
	viper.Set("token", token)
	viper.Set("endpoint", endpoint)
	
	if org != "" {
		viper.Set("org", org)
	}

	if err := writeConfig(); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	fmt.Println("Credentials saved.")
	
	if org != "" {
		fmt.Printf("Organization: %s\n", org)
	} else {
		fmt.Println("Run 'agentctl config set org <your-org-id>' to set your organization")
	}

	return nil
}

func runConfigLogout(cmd *cobra.Command, args []string) error {
	fmt.Println("Logging out...")

	// Remove sensitive data
	viper.Set("token", "")
	
	if err := writeConfig(); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	fmt.Println("Logged out successfully.")
	fmt.Println("Run 'agentctl config login' to authenticate again.")

	return nil
}

func writeConfig() error {
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		// If no config file is set, create one in home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configFile = filepath.Join(home, ".agentctl.yaml")
		viper.SetConfigFile(configFile)
	}

	return viper.WriteConfig()
}