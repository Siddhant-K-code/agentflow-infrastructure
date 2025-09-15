package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Agent struct {
	Name string `json:"name"`
	Port string `json:"port"`
}

type TaskRequest struct {
	ID       string                 `json:"id"`
	Input    map[string]interface{} `json:"input"`
	Metadata map[string]string      `json:"metadata"`
}

type TaskResponse struct {
	ID     string                 `json:"id"`
	Output map[string]interface{} `json:"output"`
	Status string                 `json:"status"`
	Agent  string                 `json:"agent"`
}

func main() {
	agent := Agent{
		Name: getEnv("AGENT_NAME", "hello-world"),
		Port: getEnv("PORT", "3000"),
	}

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/execute", createExecuteHandler(agent))
	http.HandleFunc("/status", statusHandler)

	log.Printf("🤖 Agent '%s' starting on port %s", agent.Name, agent.Port)
	log.Fatal(http.ListenAndServe(":"+agent.Port, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent":   getEnv("AGENT_NAME", "hello-world"),
		"status":  "running",
		"uptime":  "unknown",
		"version": "1.0.0",
	})
}

func createExecuteHandler(agent Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req TaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		log.Printf("📥 Executing task %s with input: %+v", req.ID, req.Input)

		// Simulate processing time
		time.Sleep(time.Duration(1+len(agent.Name)%3) * time.Second)

		// Generate response based on agent type
		output := generateOutput(agent.Name, req.Input)

		response := TaskResponse{
			ID:     req.ID,
			Output: output,
			Status: "completed",
			Agent:  agent.Name,
		}

		log.Printf("✅ Task %s completed: %+v", req.ID, response.Output)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func generateOutput(agentName string, input map[string]interface{}) map[string]interface{} {
	switch agentName {
	case "data-collector":
		return map[string]interface{}{
			"collected_data": []string{"item1", "item2", "item3"},
			"timestamp":      time.Now().Format(time.RFC3339),
			"source":         "mock-api",
			"count":          3,
		}
	case "data-processor":
		return map[string]interface{}{
			"processed_data": map[string]interface{}{
				"cleaned":    true,
				"validated":  true,
				"enriched":   true,
				"item_count": 3,
			},
			"processing_time": "2.5s",
			"quality_score":   0.95,
			"timestamp":       time.Now().Format(time.RFC3339),
		}
	case "data-publisher":
		return map[string]interface{}{
			"published":     true,
			"endpoint":      "https://api.target.com/publish",
			"status_code":   200,
			"records_sent":  3,
			"timestamp":     time.Now().Format(time.RFC3339),
		}
	default:
		return map[string]interface{}{
			"message":   fmt.Sprintf("Hello from %s!", agentName),
			"input":     input,
			"timestamp": time.Now().Format(time.RFC3339),
			"processed": true,
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}