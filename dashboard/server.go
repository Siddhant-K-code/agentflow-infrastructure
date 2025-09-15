package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := getEnv("PORT", "3001")
	dashboardDir := getEnv("DASHBOARD_DIR", "./dashboard")
	
	// Serve static files
	fs := http.FileServer(http.Dir(dashboardDir))
	http.Handle("/", fs)
	
	log.Printf("🌐 Dashboard server starting on http://localhost:%s", port)
	log.Printf("📂 Serving files from: %s", dashboardDir)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}