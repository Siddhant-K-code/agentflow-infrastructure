package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	Commit    = "unknown"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "agentflow",
	Short: "AgentFlow - Kubernetes for AI Agents",
	Long: `AgentFlow is a comprehensive platform for deploying and orchestrating 
multi-agent AI workflows. Deploy complex agent workflows using simple YAML configurations.`,
	Version: fmt.Sprintf("%s (built at %s, commit %s)", Version, BuildTime, Commit),
}

func init() {
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(liveViewCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(replayCmd)
	rootCmd.AddCommand(initCmd)
}