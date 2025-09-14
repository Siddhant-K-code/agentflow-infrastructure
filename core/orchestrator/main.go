package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func main() {
	// Load configuration
	config := loadConfig()
	
	// Initialize orchestrator
	orchestrator, err := NewOrchestrator(config)
	if err != nil {
		log.Fatalf("Failed to create orchestrator: %v", err)
	}
	
	// Setup HTTP server
	router := mux.NewRouter()
	setupRoutes(router, orchestrator)
	
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: router,
	}
	
	// Start server
	go func() {
		log.Printf("🚀 AgentFlow Orchestrator starting on port %d", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("🛑 Shutting down orchestrator...")
	
	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	
	orchestrator.Shutdown()
	log.Println("✅ Orchestrator shutdown complete")
}

type Config struct {
	Port         int    `mapstructure:"port"`
	NatsURL      string `mapstructure:"nats_url"`
	PostgresURL  string `mapstructure:"postgres_url"`
	ClickhouseURL string `mapstructure:"clickhouse_url"`
	TracingURL   string `mapstructure:"tracing_url"`
}

func loadConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/agentflow")
	
	// Set defaults
	viper.SetDefault("port", 8080)
	viper.SetDefault("nats_url", "nats://localhost:4222")
	viper.SetDefault("postgres_url", "postgres://localhost/agentflow?sslmode=disable")
	viper.SetDefault("clickhouse_url", "tcp://localhost:9000")
	viper.SetDefault("tracing_url", "http://localhost:14268/api/traces")
	
	// Read environment variables
	viper.AutomaticEnv()
	
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Config file not found, using defaults and environment variables")
	}
	
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}
	
	return &config
}

// Orchestrator is the main workflow orchestration engine
type Orchestrator struct {
	config      *Config
	dagExecutor *DAGExecutor
	workflows   map[string]*WorkflowExecution
	mutex       sync.RWMutex
}

func NewOrchestrator(config *Config) (*Orchestrator, error) {
	executor, err := NewDAGExecutor(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create DAG executor: %w", err)
	}
	
	return &Orchestrator{
		config:      config,
		dagExecutor: executor,
		workflows:   make(map[string]*WorkflowExecution),
	}, nil
}

func (o *Orchestrator) Shutdown() {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	
	// Stop all running workflows
	for _, workflow := range o.workflows {
		workflow.Cancel()
	}
	
	o.dagExecutor.Shutdown()
}

func setupRoutes(router *mux.Router, orchestrator *Orchestrator) {
	api := router.PathPrefix("/api/v1").Subrouter()
	
	// Workflow management
	api.HandleFunc("/workflows", orchestrator.handleCreateWorkflow).Methods("POST")
	api.HandleFunc("/workflows", orchestrator.handleListWorkflows).Methods("GET")
	api.HandleFunc("/workflows/{id}", orchestrator.handleGetWorkflow).Methods("GET")
	api.HandleFunc("/workflows/{id}", orchestrator.handleDeleteWorkflow).Methods("DELETE")
	api.HandleFunc("/workflows/{id}/status", orchestrator.handleGetWorkflowStatus).Methods("GET")
	api.HandleFunc("/workflows/{id}/logs", orchestrator.handleGetWorkflowLogs).Methods("GET")
	
	// WebSocket for live updates
	api.HandleFunc("/workflows/{id}/live", orchestrator.handleWorkflowLive).Methods("GET")
	
	// Triggers
	api.HandleFunc("/trigger/{webhook}", orchestrator.handleWebhookTrigger).Methods("POST")
	
	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
}

func (o *Orchestrator) handleCreateWorkflow(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement workflow creation
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status": "created", "id": "workflow-123"}`))
}

func (o *Orchestrator) handleListWorkflows(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement workflow listing
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"workflows": []}`))
}

func (o *Orchestrator) handleGetWorkflow(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement workflow retrieval
	vars := mux.Vars(r)
	workflowID := vars["id"]
	
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"id": "%s", "status": "running"}`, workflowID)))
}

func (o *Orchestrator) handleDeleteWorkflow(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement workflow deletion
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "deleted"}`))
}

func (o *Orchestrator) handleGetWorkflowStatus(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement workflow status retrieval
	vars := mux.Vars(r)
	workflowID := vars["id"]
	
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"id": "%s", "status": "running", "agents": []}`, workflowID)))
}

func (o *Orchestrator) handleGetWorkflowLogs(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement workflow logs retrieval
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"logs": []}`))
}

func (o *Orchestrator) handleWorkflowLive(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement WebSocket connection for live updates
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("WebSocket live view not implemented yet"))
}

func (o *Orchestrator) handleWebhookTrigger(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement webhook triggers
	vars := mux.Vars(r)
	webhook := vars["webhook"]
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"webhook": "%s", "triggered": true}`, webhook)))
}