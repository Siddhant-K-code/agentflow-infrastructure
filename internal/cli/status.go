package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show system status and health",
	Long:  "Display overall system status, health metrics, and service availability",
	RunE:  runStatus,
}

func init() {
	statusCmd.Flags().StringP("output", "o", "summary", "Output format (summary, detailed, json)")
	statusCmd.Flags().BoolP("services", "s", false, "Show service status")
	statusCmd.Flags().BoolP("metrics", "m", false, "Show system metrics")
	statusCmd.Flags().BoolP("quotas", "q", false, "Show quota status")
}

func runStatus(cmd *cobra.Command, args []string) error {
	output, _ := cmd.Flags().GetString("output")
	showServices, _ := cmd.Flags().GetBool("services")
	showMetrics, _ := cmd.Flags().GetBool("metrics")
	showQuotas, _ := cmd.Flags().GetBool("quotas")

	// Mock system status data
	status := map[string]interface{}{
		"timestamp":      time.Now().Format(time.RFC3339),
		"overall_status": "healthy",
		"version":        "1.0.0",
		"uptime":         "7d 14h 32m",
		"services": map[string]interface{}{
			"control_plane": map[string]interface{}{
				"status":  "healthy",
				"uptime":  "7d 14h 32m",
				"version": "1.0.0",
				"endpoints": []string{
					"http://localhost:8080/api/v1",
				},
			},
			"workers": map[string]interface{}{
				"status":       "healthy",
				"active_count": 5,
				"total_count":  5,
				"avg_cpu":      45.2,
				"avg_memory":   62.8,
			},
			"database": map[string]interface{}{
				"status":          "healthy",
				"type":            "postgresql",
				"connections":     12,
				"max_connections": 100,
			},
			"clickhouse": map[string]interface{}{
				"status":     "healthy",
				"type":       "clickhouse",
				"disk_usage": 15.6,
				"query_rate": 125.3,
			},
			"redis": map[string]interface{}{
				"status":            "healthy",
				"memory_usage":      234.5,
				"connected_clients": 8,
			},
			"nats": map[string]interface{}{
				"status":           "healthy",
				"messages_per_sec": 45.2,
				"connections":      15,
			},
		},
		"metrics": map[string]interface{}{
			"workflows": map[string]interface{}{
				"total_runs_24h": 1247,
				"success_rate":   96.8,
				"avg_duration":   "2m 15s",
				"active_runs":    23,
			},
			"costs": map[string]interface{}{
				"total_24h":          125.50,
				"avg_per_run":        0.85,
				"budget_utilization": 65.2,
			},
			"performance": map[string]interface{}{
				"avg_latency":    "1.2s",
				"p95_latency":    "3.8s",
				"error_rate":     0.8,
				"cache_hit_rate": 34.5,
			},
		},
		"quotas": map[string]interface{}{
			"openai": map[string]interface{}{
				"current_qps": 15,
				"limit_qps":   100,
				"utilization": 15.0,
				"status":      "healthy",
			},
			"anthropic": map[string]interface{}{
				"current_qps": 8,
				"limit_qps":   50,
				"utilization": 16.0,
				"status":      "healthy",
			},
			"google": map[string]interface{}{
				"current_qps": 5,
				"limit_qps":   75,
				"utilization": 6.7,
				"status":      "healthy",
			},
		},
		"alerts": []map[string]interface{}{
			{
				"severity":  "warning",
				"message":   "Cache hit rate below optimal (34.5%)",
				"timestamp": time.Now().Add(-15 * time.Minute).Format(time.RFC3339),
			},
		},
	}

	if output == "json" {
		outputBytes, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(outputBytes))
		return nil
	}

	// Summary or detailed output
	fmt.Println("AgentFlow System Status")
	fmt.Println("=======================")
	fmt.Printf("Overall Status: %s\n", getStatusIcon(status["overall_status"].(string)))
	fmt.Printf("Version: %s\n", status["version"])
	fmt.Printf("Uptime: %s\n", status["uptime"])
	fmt.Printf("Timestamp: %s\n", status["timestamp"])

	if showServices || output == "detailed" {
		fmt.Println("\nService Status:")
		fmt.Println("---------------")
		services := status["services"].(map[string]interface{})

		for serviceName, serviceData := range services {
			service := serviceData.(map[string]interface{})
			serviceStatus := service["status"].(string)
			fmt.Printf("%-15s: %s", serviceName, getStatusIcon(serviceStatus))

			// Add service-specific details
			switch serviceName {
			case "control_plane":
				fmt.Printf(" (v%s)", service["version"])
			case "workers":
				fmt.Printf(" (%d/%d active)", service["active_count"], service["total_count"])
			case "database":
				fmt.Printf(" (%d/%d connections)", service["connections"], service["max_connections"])
			case "clickhouse":
				fmt.Printf(" (%.1f%% disk)", service["disk_usage"])
			case "redis":
				fmt.Printf(" (%.1fMB memory)", service["memory_usage"])
			case "nats":
				fmt.Printf(" (%.1f msg/s)", service["messages_per_sec"])
			}
			fmt.Println()
		}
	}

	if showMetrics || output == "detailed" {
		fmt.Println("\nSystem Metrics:")
		fmt.Println("---------------")
		metrics := status["metrics"].(map[string]interface{})

		workflows := metrics["workflows"].(map[string]interface{})
		fmt.Printf("Workflows (24h): %d runs, %.1f%% success rate, %s avg duration\n",
			workflows["total_runs_24h"], workflows["success_rate"], workflows["avg_duration"])
		fmt.Printf("Active runs: %d\n", workflows["active_runs"])

		costs := metrics["costs"].(map[string]interface{})
		fmt.Printf("Costs (24h): $%.2f total, $%.2f avg/run, %.1f%% budget used\n",
			costs["total_24h"], costs["avg_per_run"], costs["budget_utilization"])

		performance := metrics["performance"].(map[string]interface{})
		fmt.Printf("Performance: %s avg latency, %s P95, %.1f%% error rate\n",
			performance["avg_latency"], performance["p95_latency"], performance["error_rate"])
		fmt.Printf("Cache hit rate: %.1f%%\n", performance["cache_hit_rate"])
	}

	if showQuotas || output == "detailed" {
		fmt.Println("\nProvider Quotas:")
		fmt.Println("----------------")
		quotas := status["quotas"].(map[string]interface{})

		fmt.Printf("%-12s %-10s %-10s %-12s %-8s\n", "PROVIDER", "CURRENT", "LIMIT", "UTILIZATION", "STATUS")
		fmt.Println("----------------------------------------------------------")

		for provider, quotaData := range quotas {
			quota := quotaData.(map[string]interface{})
			fmt.Printf("%-12s %-10d %-10d %-12.1f%% %s\n",
				provider,
				int(quota["current_qps"].(float64)),
				int(quota["limit_qps"].(float64)),
				quota["utilization"].(float64),
				getStatusIcon(quota["status"].(string)),
			)
		}
	}

	// Show alerts
	alerts := status["alerts"].([]map[string]interface{})
	if len(alerts) > 0 {
		fmt.Println("\nActive Alerts:")
		fmt.Println("--------------")
		for _, alert := range alerts {
			severity := alert["severity"].(string)
			message := alert["message"].(string)
			timestamp, _ := time.Parse(time.RFC3339, alert["timestamp"].(string))

			icon := "ℹ"
			switch severity {
			case "critical":
				icon = "🔴"
			case "warning":
				icon = "⚠️"
			case "info":
				icon = "ℹ️"
			}

			fmt.Printf("%s [%s] %s (%s ago)\n",
				icon, severity, message,
				time.Since(timestamp).Truncate(time.Minute))
		}
	}

	// Show recommendations
	fmt.Println("\nRecommendations:")
	fmt.Println("----------------")

	// Generate recommendations based on status
	recommendations := []string{}

	metrics := status["metrics"].(map[string]interface{})
	performance := metrics["performance"].(map[string]interface{})

	if performance["cache_hit_rate"].(float64) < 40 {
		recommendations = append(recommendations, "• Consider optimizing caching strategy to improve hit rate")
	}

	if performance["error_rate"].(float64) > 1.0 {
		recommendations = append(recommendations, "• Investigate error rate - consider adding retry logic")
	}

	costs := metrics["costs"].(map[string]interface{})
	if costs["budget_utilization"].(float64) > 80 {
		recommendations = append(recommendations, "• Budget utilization is high - consider cost optimization")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "• System is running optimally")
	}

	for _, rec := range recommendations {
		fmt.Println(rec)
	}

	return nil
}

func getStatusIcon(status string) string {
	switch status {
	case "healthy":
		return "✅ " + status
	case "warning":
		return "⚠️  " + status
	case "critical":
		return "🔴 " + status
	case "degraded":
		return "🟡 " + status
	case "down":
		return "❌ " + status
	default:
		return "❓ " + status
	}
}
