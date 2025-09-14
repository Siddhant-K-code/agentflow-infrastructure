package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func main() {
	// Load configuration
	config := loadConfig()
	
	// Initialize LLM router
	router, err := NewLLMRouter(config)
	if err != nil {
		log.Fatalf("Failed to create LLM router: %v", err)
	}
	
	// Setup HTTP server
	httpRouter := mux.NewRouter()
	setupRoutes(httpRouter, router)
	
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: httpRouter,
	}
	
	// Start server
	go func() {
		log.Printf("🤖 AgentFlow LLM Router starting on port %d", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("🛑 Shutting down LLM router...")
	
	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	
	router.Shutdown()
	log.Println("✅ LLM router shutdown complete")
}

type Config struct {
	Port     int                    `mapstructure:"port"`
	NatsURL  string                 `mapstructure:"nats_url"`
	Providers map[string]ProviderConfig `mapstructure:"providers"`
}

type ProviderConfig struct {
	APIKey   string  `mapstructure:"api_key"`
	BaseURL  string  `mapstructure:"base_url"`
	CostPer1K float64 `mapstructure:"cost_per_1k"`
	RateLimit int    `mapstructure:"rate_limit"`
}

func loadConfig() *Config {
	viper.SetConfigName("llm-router-config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/agentflow")
	
	// Set defaults
	viper.SetDefault("port", 8082)
	viper.SetDefault("nats_url", "nats://localhost:4222")
	
	// Default providers
	viper.SetDefault("providers.openai.base_url", "https://api.openai.com/v1")
	viper.SetDefault("providers.openai.cost_per_1k", 0.002)
	viper.SetDefault("providers.anthropic.base_url", "https://api.anthropic.com")
	viper.SetDefault("providers.anthropic.cost_per_1k", 0.008)
	
	// Read environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Config file not found, using defaults and environment variables")
	}
	
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}
	
	return &config
}

// LLMRouter manages routing to different LLM providers
type LLMRouter struct {
	config    *Config
	providers map[string]*LLMProvider
	metrics   *RouterMetrics
	mutex     sync.RWMutex
}

type LLMProvider struct {
	Name      string
	Config    ProviderConfig
	Client    *http.Client
	Stats     ProviderStats
	mutex     sync.RWMutex
}

type ProviderStats struct {
	TotalRequests   int64   `json:"total_requests"`
	SuccessfulCalls int64   `json:"successful_calls"`
	FailedCalls     int64   `json:"failed_calls"`
	TotalCost       float64 `json:"total_cost"`
	AvgLatency      float64 `json:"avg_latency_ms"`
	LastUsed        time.Time `json:"last_used"`
}

type RouterMetrics struct {
	TotalRequests     int64             `json:"total_requests"`
	CostSavings       float64           `json:"cost_savings"`
	ProviderStats     map[string]ProviderStats `json:"provider_stats"`
	OptimalSelections int64             `json:"optimal_selections"`
	mutex             sync.RWMutex
}

func NewLLMRouter(config *Config) (*LLMRouter, error) {
	router := &LLMRouter{
		config:    config,
		providers: make(map[string]*LLMProvider),
		metrics:   &RouterMetrics{
			ProviderStats: make(map[string]ProviderStats),
		},
	}
	
	// Initialize providers
	for name, providerConfig := range config.Providers {
		provider := &LLMProvider{
			Name:   name,
			Config: providerConfig,
			Client: &http.Client{
				Timeout: 30 * time.Second,
			},
		}
		router.providers[name] = provider
		log.Printf("✅ Initialized LLM provider: %s", name)
	}
	
	return router, nil
}

func (r *LLMRouter) Shutdown() {
	log.Println("🛑 Shutting down LLM router...")
	// Cleanup if needed
}

type LLMRequest struct {
	Model       string                 `json:"model"`
	Messages    []Message              `json:"messages"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMResponse struct {
	ID      string    `json:"id"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Usage   Usage     `json:"usage"`
	Cost    float64   `json:"cost"`
	Provider string   `json:"provider"`
	Latency int64     `json:"latency_ms"`
}

type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func setupRoutes(router *mux.Router, llmRouter *LLMRouter) {
	api := router.PathPrefix("/api/v1").Subrouter()
	
	// LLM completion endpoint
	api.HandleFunc("/chat/completions", llmRouter.handleChatCompletions).Methods("POST")
	
	// Router management
	api.HandleFunc("/providers", llmRouter.handleListProviders).Methods("GET")
	api.HandleFunc("/metrics", llmRouter.handleGetMetrics).Methods("GET")
	api.HandleFunc("/optimal-provider", llmRouter.handleGetOptimalProvider).Methods("POST")
	
	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "llm-router"}`))
	}).Methods("GET")
}

func (r *LLMRouter) handleChatCompletions(w http.ResponseWriter, req *http.Request) {
	var llmReq LLMRequest
	if err := json.NewDecoder(req.Body).Decode(&llmReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Select optimal provider
	provider := r.selectOptimalProvider(llmReq)
	if provider == nil {
		http.Error(w, "No available provider for this request", http.StatusServiceUnavailable)
		return
	}
	
	// Make LLM call
	response, err := r.makeLLMCall(provider, llmReq)
	if err != nil {
		log.Printf("LLM call failed: %v", err)
		http.Error(w, "LLM call failed", http.StatusInternalServerError)
		return
	}
	
	// Update metrics
	r.updateMetrics(provider, response)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (r *LLMRouter) selectOptimalProvider(req LLMRequest) *LLMProvider {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	// Simple cost-based selection algorithm
	var bestProvider *LLMProvider
	var lowestCost float64 = float64(^uint(0) >> 1) // Max float64
	
	for _, provider := range r.providers {
		if r.canHandleModel(provider, req.Model) {
			cost := r.estimateCost(provider, req)
			if cost < lowestCost {
				lowestCost = cost
				bestProvider = provider
			}
		}
	}
	
	return bestProvider
}

func (r *LLMRouter) canHandleModel(provider *LLMProvider, model string) bool {
	// Simple model mapping - in production this would be more sophisticated
	switch provider.Name {
	case "openai":
		return strings.HasPrefix(model, "gpt-")
	case "anthropic":
		return strings.HasPrefix(model, "claude-")
	default:
		return false
	}
}

func (r *LLMRouter) estimateCost(provider *LLMProvider, req LLMRequest) float64 {
	// Estimate tokens (very rough approximation)
	estimatedTokens := 0
	for _, msg := range req.Messages {
		estimatedTokens += len(strings.Fields(msg.Content)) * 2 // Rough token estimation
	}
	
	return float64(estimatedTokens) / 1000.0 * provider.Config.CostPer1K
}

func (r *LLMRouter) makeLLMCall(provider *LLMProvider, req LLMRequest) (*LLMResponse, error) {
	start := time.Now()
	
	// TODO: Implement actual LLM API calls to different providers
	// For now, return a mock response
	
	latency := time.Since(start).Milliseconds()
	
	response := &LLMResponse{
		ID:    fmt.Sprintf("resp_%d", time.Now().Unix()),
		Model: req.Model,
		Choices: []Choice{
			{
				Message: Message{
					Role:    "assistant",
					Content: "This is a mock response from the LLM router.",
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
		},
		Cost:     0.003,
		Provider: provider.Name,
		Latency:  latency,
	}
	
	return response, nil
}

func (r *LLMRouter) updateMetrics(provider *LLMProvider, response *LLMResponse) {
	provider.mutex.Lock()
	provider.Stats.TotalRequests++
	provider.Stats.SuccessfulCalls++
	provider.Stats.TotalCost += response.Cost
	provider.Stats.AvgLatency = (provider.Stats.AvgLatency + float64(response.Latency)) / 2
	provider.Stats.LastUsed = time.Now()
	provider.mutex.Unlock()
	
	r.metrics.mutex.Lock()
	r.metrics.TotalRequests++
	r.metrics.OptimalSelections++
	r.metrics.ProviderStats[provider.Name] = provider.Stats
	r.metrics.mutex.Unlock()
}

func (r *LLMRouter) handleListProviders(w http.ResponseWriter, req *http.Request) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	providers := make([]map[string]interface{}, 0, len(r.providers))
	for name, provider := range r.providers {
		providers = append(providers, map[string]interface{}{
			"name":       name,
			"base_url":   provider.Config.BaseURL,
			"cost_per_1k": provider.Config.CostPer1K,
			"stats":      provider.Stats,
		})
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"providers": providers,
	})
}

func (r *LLMRouter) handleGetMetrics(w http.ResponseWriter, req *http.Request) {
	r.metrics.mutex.RLock()
	defer r.metrics.mutex.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(r.metrics)
}

func (r *LLMRouter) handleGetOptimalProvider(w http.ResponseWriter, req *http.Request) {
	var llmReq LLMRequest
	if err := json.NewDecoder(req.Body).Decode(&llmReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	provider := r.selectOptimalProvider(llmReq)
	if provider == nil {
		http.Error(w, "No available provider", http.StatusServiceUnavailable)
		return
	}
	
	cost := r.estimateCost(provider, llmReq)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"provider":      provider.Name,
		"estimated_cost": cost,
		"reason":        "cost_optimal",
	})
}