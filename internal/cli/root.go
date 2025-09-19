package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "agentctl",
	Short: "AgentFlow CLI - Manage agent workflows and infrastructure",
	Long: `AgentFlow CLI provides command-line access to the AgentFlow platform.
	
Use agentctl to:
- Submit and manage workflow runs
- Monitor execution status and traces
- Manage prompts and deployments
- Configure cost budgets and policies
- Analyze performance and costs`,
	Version: "1.0.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.agentctl.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().String("endpoint", "http://localhost:8080", "AgentFlow API endpoint")
	rootCmd.PersistentFlags().String("org", "", "Organization ID")
	rootCmd.PersistentFlags().String("token", "", "Authentication token")

	// Bind flags to viper
	_ = viper.BindPFlag("endpoint", rootCmd.PersistentFlags().Lookup("endpoint"))
	_ = viper.BindPFlag("org", rootCmd.PersistentFlags().Lookup("org"))
	_ = viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))

	// Add subcommands
	rootCmd.AddCommand(workflowCmd)
	rootCmd.AddCommand(promptCmd)
	rootCmd.AddCommand(traceCmd)
	rootCmd.AddCommand(budgetCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(statusCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".agentctl" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".agentctl")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
