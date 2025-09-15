package main

import (
	"context"
	"encoding/json"
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
	var workflow Workflow
	if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Execute workflow
	execution, err := o.dagExecutor.ExecuteWorkflow(&workflow)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to execute workflow: %v", err), http.StatusInternalServerError)
		return
	}

	// Store execution
	o.mutex.Lock()
	o.workflows[execution.ID] = execution
	o.mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := map[string]interface{}{
		"id":     execution.ID,
		"status": execution.GetStatus(),
		"name":   workflow.Metadata.Name,
	}
	json.NewEncoder(w).Encode(response)
}

func (o *Orchestrator) handleListWorkflows(w http.ResponseWriter, r *http.Request) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	workflows := make([]map[string]interface{}, 0)
	for id, execution := range o.workflows {
		workflows = append(workflows, map[string]interface{}{
			"id":     id,
			"status": execution.GetStatus(),
			"name":   execution.Workflow.Metadata.Name,
			"start_time": execution.StartTime,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"workflows": workflows,
	})
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
	vars := mux.Vars(r)
	workflowID := vars["id"]

	o.mutex.RLock()
	execution, exists := o.workflows[workflowID]
	o.mutex.RUnlock()

	if !exists {
		http.Error(w, "Workflow not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"id":         execution.ID,
		"status":     execution.GetStatus(),
		"agents":     execution.GetAgents(),
		"start_time": execution.StartTime,
		"end_time":   execution.EndTime,
	}

	if execution.Error != "" {
		response["error"] = execution.Error
	}

	json.NewEncoder(w).Encode(response)
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

// Workflow structures (copied from deploy.go for consistency)
type Workflow struct {
	APIVersion string `yaml:"apiVersion" json:"apiVersion"`
	Kind       string `yaml:"kind" json:"kind"`
	Metadata   struct {
		Name      string            `yaml:"name" json:"name"`
		Namespace string            `yaml:"namespace,omitempty" json:"namespace,omitempty"`
		Labels    map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	} `yaml:"metadata" json:"metadata"`
	Spec WorkflowSpec `yaml:"spec" json:"spec"`
}

type WorkflowSpec struct {
	Agents   []Agent   `yaml:"agents" json:"agents"`
	Triggers []Trigger `yaml:"triggers,omitempty" json:"triggers,omitempty"`
	Config   ConfigSpec    `yaml:"config,omitempty" json:"config,omitempty"`
}

type Agent struct {
	Name      string            `yaml:"name" json:"name"`
	Image     string            `yaml:"image" json:"image"`
	LLM       LLMConfig         `yaml:"llm" json:"llm"`
	DependsOn []string          `yaml:"dependsOn,omitempty" json:"dependsOn,omitempty"`
	Resources Resources         `yaml:"resources,omitempty" json:"resources,omitempty"`
	Env       map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	Timeout   string            `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Retries   int               `yaml:"retries,omitempty" json:"retries,omitempty"`
}

type LLMConfig struct {
	Provider string            `yaml:"provider" json:"provider"`
	Model    string            `yaml:"model" json:"model"`
	Config   map[string]string `yaml:"config,omitempty" json:"config,omitempty"`
}

type Resources struct {
	Memory string `yaml:"memory,omitempty" json:"memory,omitempty"`
	CPU    string `yaml:"cpu,omitempty" json:"cpu,omitempty"`
}

type Trigger struct {
	Schedule string `yaml:"schedule,omitempty" json:"schedule,omitempty"`
	Webhook  string `yaml:"webhook,omitempty" json:"webhook,omitempty"`
	Event    string `yaml:"event,omitempty" json:"event,omitempty"`
}

type ConfigSpec struct {
	Parallelism int    `yaml:"parallelism,omitempty" json:"parallelism,omitempty"`
	Timeout     string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	RetryPolicy string `yaml:"retryPolicy,omitempty" json:"retryPolicy,omitempty"`
}