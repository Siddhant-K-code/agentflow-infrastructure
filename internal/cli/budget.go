package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var budgetCmd = &cobra.Command{
	Use:   "budget",
	Short: "Manage cost budgets and spending",
	Long:  "Create, monitor, and manage cost budgets and spending limits",
}

var budgetCreateCmd = &cobra.Command{
	Use:   "create [amount]",
	Short: "Create a new budget",
	Args:  cobra.ExactArgs(1),
	RunE:  runBudgetCreate,
}

var budgetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List budgets",
	RunE:  runBudgetList,
}

var budgetStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current budget status",
	RunE:  runBudgetStatus,
}

var budgetUpdateCmd = &cobra.Command{
	Use:   "update [budget-id] [amount]",
	Short: "Update budget limit",
	Args:  cobra.ExactArgs(2),
	RunE:  runBudgetUpdate,
}

var budgetDeleteCmd = &cobra.Command{
	Use:   "delete [budget-id]",
	Short: "Delete a budget",
	Args:  cobra.ExactArgs(1),
	RunE:  runBudgetDelete,
}

var budgetAnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze spending patterns",
	RunE:  runBudgetAnalyze,
}

func init() {
	// Create command flags
	budgetCreateCmd.Flags().StringP("period", "p", "monthly", "Budget period (daily, weekly, monthly)")
	budgetCreateCmd.Flags().StringP("project", "", "", "Project ID (optional)")
	budgetCreateCmd.Flags().StringP("description", "d", "", "Budget description")

	// List command flags
	budgetListCmd.Flags().StringP("status", "s", "", "Filter by status (healthy, warning, critical, exceeded)")
	budgetListCmd.Flags().BoolP("active", "a", false, "Show only active budgets")

	// Status command flags
	budgetStatusCmd.Flags().StringP("project", "p", "", "Project ID (optional)")
	budgetStatusCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	// Analyze command flags
	budgetAnalyzeCmd.Flags().StringP("period", "p", "30d", "Analysis period")
	budgetAnalyzeCmd.Flags().StringSliceP("group-by", "g", []string{"provider"}, "Group by (provider, model, workflow, quality)")
	budgetAnalyzeCmd.Flags().BoolP("trends", "t", false, "Show spending trends")
	budgetAnalyzeCmd.Flags().BoolP("forecast", "f", false, "Show spending forecast")

	// Add subcommands
	budgetCmd.AddCommand(budgetCreateCmd)
	budgetCmd.AddCommand(budgetListCmd)
	budgetCmd.AddCommand(budgetStatusCmd)
	budgetCmd.AddCommand(budgetUpdateCmd)
	budgetCmd.AddCommand(budgetDeleteCmd)
	budgetCmd.AddCommand(budgetAnalyzeCmd)
}

func runBudgetCreate(cmd *cobra.Command, args []string) error {
	amountStr := args[0]
	
	// Parse amount (support $100, 100, 10000 cents)
	var amountCents int64
	if amountStr[0] == '$' {
		amount, err := strconv.ParseFloat(amountStr[1:], 64)
		if err != nil {
			return fmt.Errorf("invalid amount format: %s", amountStr)
		}
		amountCents = int64(amount * 100)
	} else {
		amount, err := strconv.ParseInt(amountStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid amount format: %s", amountStr)
		}
		if amount < 1000 { // Assume dollars if less than 1000
			amountCents = amount * 100
		} else { // Assume cents if 1000 or more
			amountCents = amount
		}
	}

	period, _ := cmd.Flags().GetString("period")
	project, _ := cmd.Flags().GetString("project")
	description, _ := cmd.Flags().GetString("description")

	fmt.Printf("Creating budget:\n")
	fmt.Printf("  Amount: $%.2f\n", float64(amountCents)/100)
	fmt.Printf("  Period: %s\n", period)
	if project != "" {
		fmt.Printf("  Project: %s\n", project)
	}
	if description != "" {
		fmt.Printf("  Description: %s\n", description)
	}

	// Mock budget creation
	budgetID := fmt.Sprintf("budget_%d", time.Now().Unix())
	
	fmt.Printf("\nBudget created successfully!\n")
	fmt.Printf("Budget ID: %s\n", budgetID)
	
	// Calculate period dates
	now := time.Now()
	var startDate, endDate time.Time
	
	switch period {
	case "daily":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endDate = startDate.Add(24 * time.Hour)
	case "weekly":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		daysToMonday := weekday - 1
		startDate = time.Date(now.Year(), now.Month(), now.Day()-daysToMonday, 0, 0, 0, 0, now.Location())
		endDate = startDate.Add(7 * 24 * time.Hour)
	case "monthly":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 1, 0)
	}
	
	fmt.Printf("Period: %s to %s\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	return nil
}

func runBudgetList(cmd *cobra.Command, args []string) error {
	statusFilter, _ := cmd.Flags().GetString("status")
	activeOnly, _ := cmd.Flags().GetBool("active")

	fmt.Printf("Listing budgets")
	if statusFilter != "" {
		fmt.Printf(" (status: %s)", statusFilter)
	}
	if activeOnly {
		fmt.Printf(" (active only)")
	}
	fmt.Println()

	// Mock budget list
	budgets := []map[string]interface{}{
		{
			"id":          "budget_1234567890",
			"period":      "monthly",
			"limit":       50000, // cents
			"spent":       32500, // cents
			"utilization": 65.0,
			"status":      "healthy",
			"start_date":  "2024-01-01",
			"end_date":    "2024-02-01",
			"project":     "",
		},
		{
			"id":          "budget_1234567891",
			"period":      "weekly",
			"limit":       10000, // cents
			"spent":       8500,  // cents
			"utilization": 85.0,
			"status":      "warning",
			"start_date":  "2024-01-15",
			"end_date":    "2024-01-22",
			"project":     "proj_analytics",
		},
		{
			"id":          "budget_1234567892",
			"period":      "daily",
			"limit":       2000, // cents
			"spent":       2100, // cents
			"utilization": 105.0,
			"status":      "exceeded",
			"start_date":  "2024-01-15",
			"end_date":    "2024-01-16",
			"project":     "proj_testing",
		},
	}

	// Print table header
	fmt.Printf("%-20s %-8s %-10s %-10s %-10s %-12s %-10s\n", 
		"BUDGET ID", "PERIOD", "LIMIT", "SPENT", "REMAINING", "UTILIZATION", "STATUS")
	fmt.Println("-------------------------------------------------------------------------------------")

	// Print budgets
	for _, budget := range budgets {
		if statusFilter != "" && budget["status"] != statusFilter {
			continue
		}

		limit := budget["limit"].(int)
		spent := budget["spent"].(int)
		remaining := limit - spent
		utilization := budget["utilization"].(float64)
		status := budget["status"].(string)

		// Color coding for status (would use actual colors in terminal)
		statusDisplay := status
		switch status {
		case "exceeded":
			statusDisplay = "EXCEEDED"
		case "critical":
			statusDisplay = "CRITICAL"
		case "warning":
			statusDisplay = "WARNING"
		case "healthy":
			statusDisplay = "HEALTHY"
		}

		fmt.Printf("%-20s %-8s $%-9.2f $%-9.2f $%-9.2f %-12.1f%% %-10s\n",
			budget["id"].(string)[:18]+"...",
			budget["period"].(string),
			float64(limit)/100,
			float64(spent)/100,
			float64(remaining)/100,
			utilization,
			statusDisplay,
		)
	}

	return nil
}

func runBudgetStatus(cmd *cobra.Command, args []string) error {
	project, _ := cmd.Flags().GetString("project")
	output, _ := cmd.Flags().GetString("output")

	// Mock current budget status
	status := map[string]interface{}{
		"current_budget": map[string]interface{}{
			"id":          "budget_1234567890",
			"period":      "monthly",
			"limit_cents": 50000,
			"spent_cents": 32500,
			"remaining_cents": 17500,
			"utilization_pct": 65.0,
			"status":      "healthy",
			"period_start": "2024-01-01T00:00:00Z",
			"period_end":   "2024-02-01T00:00:00Z",
		},
		"daily_spending": []map[string]interface{}{
			{"date": "2024-01-13", "spent_cents": 1200},
			{"date": "2024-01-14", "spent_cents": 1800},
			{"date": "2024-01-15", "spent_cents": 2100},
		},
		"top_spenders": []map[string]interface{}{
			{"category": "LLM Calls", "spent_cents": 28000, "percentage": 86.2},
			{"category": "Tool Calls", "spent_cents": 3500, "percentage": 10.8},
			{"category": "Compute", "spent_cents": 1000, "percentage": 3.0},
		},
		"alerts": []map[string]interface{}{
			{
				"type":    "approaching_limit",
				"message": "Budget is 65% utilized",
				"severity": "info",
			},
		},
	}

	if output == "json" {
		outputBytes, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(outputBytes))
	} else {
		budget := status["current_budget"].(map[string]interface{})
		
		fmt.Println("Current Budget Status:")
		fmt.Println("======================")
		fmt.Printf("Budget ID: %s\n", budget["id"])
		fmt.Printf("Period: %s\n", budget["period"])
		fmt.Printf("Limit: $%.2f\n", float64(budget["limit_cents"].(int))/100)
		fmt.Printf("Spent: $%.2f\n", float64(budget["spent_cents"].(int))/100)
		fmt.Printf("Remaining: $%.2f\n", float64(budget["remaining_cents"].(int))/100)
		fmt.Printf("Utilization: %.1f%%\n", budget["utilization_pct"])
		fmt.Printf("Status: %s\n", budget["status"])

		fmt.Println("\nRecent Daily Spending:")
		dailySpending := status["daily_spending"].([]map[string]interface{})
		for _, day := range dailySpending {
			fmt.Printf("  %s: $%.2f\n", day["date"], float64(day["spent_cents"].(int))/100)
		}

		fmt.Println("\nTop Spending Categories:")
		topSpenders := status["top_spenders"].([]map[string]interface{})
		for _, spender := range topSpenders {
			fmt.Printf("  %-12s: $%-8.2f (%.1f%%)\n", 
				spender["category"], 
				float64(spender["spent_cents"].(int))/100,
				spender["percentage"])
		}

		alerts := status["alerts"].([]map[string]interface{})
		if len(alerts) > 0 {
			fmt.Println("\nAlerts:")
			for _, alert := range alerts {
				fmt.Printf("  [%s] %s\n", alert["severity"], alert["message"])
			}
		}
	}

	return nil
}

func runBudgetUpdate(cmd *cobra.Command, args []string) error {
	budgetID := args[0]
	amountStr := args[1]

	// Parse amount
	var amountCents int64
	if amountStr[0] == '$' {
		amount, err := strconv.ParseFloat(amountStr[1:], 64)
		if err != nil {
			return fmt.Errorf("invalid amount format: %s", amountStr)
		}
		amountCents = int64(amount * 100)
	} else {
		amount, err := strconv.ParseInt(amountStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid amount format: %s", amountStr)
		}
		if amount < 1000 {
			amountCents = amount * 100
		} else {
			amountCents = amount
		}
	}

	fmt.Printf("Updating budget: %s\n", budgetID)
	fmt.Printf("New limit: $%.2f\n", float64(amountCents)/100)

	// Mock update
	fmt.Printf("Budget updated successfully!\n")

	return nil
}

func runBudgetDelete(cmd *cobra.Command, args []string) error {
	budgetID := args[0]

	fmt.Printf("Deleting budget: %s\n", budgetID)
	fmt.Printf("Are you sure? This action cannot be undone. (y/N): ")
	
	// In a real implementation, would wait for user input
	fmt.Printf("y\n")
	fmt.Printf("Budget deleted successfully!\n")

	return nil
}

func runBudgetAnalyze(cmd *cobra.Command, args []string) error {
	period, _ := cmd.Flags().GetString("period")
	groupBy, _ := cmd.Flags().GetStringSlice("group-by")
	showTrends, _ := cmd.Flags().GetBool("trends")
	showForecast, _ := cmd.Flags().GetBool("forecast")

	fmt.Printf("Analyzing spending patterns (period: %s)\n", period)
	fmt.Printf("Grouping by: %v\n", groupBy)

	// Mock analysis results
	fmt.Println("\nSpending Analysis:")
	fmt.Println("==================")
	fmt.Printf("Total spending: $1,250.00\n")
	fmt.Printf("Average daily: $41.67\n")
	fmt.Printf("Peak day: $85.00 (Jan 15)\n")
	fmt.Printf("Lowest day: $12.50 (Jan 8)\n")

	fmt.Println("\nBreakdown by Provider:")
	fmt.Printf("  OpenAI:     $875.00 (70.0%%)\n")
	fmt.Printf("  Anthropic:  $250.00 (20.0%%)\n")
	fmt.Printf("  Google:     $125.00 (10.0%%)\n")

	fmt.Println("\nBreakdown by Model:")
	fmt.Printf("  GPT-4:      $625.00 (50.0%%)\n")
	fmt.Printf("  GPT-3.5:    $250.00 (20.0%%)\n")
	fmt.Printf("  Claude-3:   $250.00 (20.0%%)\n")
	fmt.Printf("  Gemini:     $125.00 (10.0%%)\n")

	if showTrends {
		fmt.Println("\nSpending Trends:")
		fmt.Printf("  Week 1: $200.00 (16.0%%)\n")
		fmt.Printf("  Week 2: $350.00 (28.0%%) ↑\n")
		fmt.Printf("  Week 3: $450.00 (36.0%%) ↑\n")
		fmt.Printf("  Week 4: $250.00 (20.0%%) ↓\n")
		fmt.Printf("  Trend: Increasing (+15%% week-over-week)\n")
	}

	if showForecast {
		fmt.Println("\nSpending Forecast:")
		fmt.Printf("  Next 7 days:  $300.00 ± $50.00\n")
		fmt.Printf("  Next 30 days: $1,400.00 ± $200.00\n")
		fmt.Printf("  Confidence: 75%%\n")
		fmt.Printf("  Budget risk: Low (within limits)\n")
	}

	fmt.Println("\nOptimization Opportunities:")
	fmt.Printf("  • Switch 30%% of GPT-4 calls to GPT-3.5: Save ~$125/month\n")
	fmt.Printf("  • Implement caching: Save ~$75/month\n")
	fmt.Printf("  • Use cheaper providers for simple tasks: Save ~$50/month\n")
	fmt.Printf("  Total potential savings: $250/month (20%%)\n")

	return nil
}