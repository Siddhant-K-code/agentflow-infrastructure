package cas

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/agentflow/infrastructure/internal/db"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type ProviderRouter struct {
	postgres *db.PostgresDB
	redis    *redis.Client
	bandit   *MultiArmedBandit
}

func NewProviderRouter(pg *db.PostgresDB, redisClient *redis.Client) *ProviderRouter {
	return &ProviderRouter{
		postgres: pg,
		redis:    redisClient,
		bandit:   NewMultiArmedBandit(),
	}
}

// GetAvailableProviders retrieves providers available for a quality tier
func (pr *ProviderRouter) GetAvailableProviders(ctx context.Context, orgID uuid.UUID, qualityTier QualityTier) ([]ProviderConfig, error) {
	query := `SELECT id, org_id, provider_name, model_name, config, 
			  cost_per_token_prompt, cost_per_token_completion, qps_limit, enabled, created_at
			  FROM provider_config 
			  WHERE org_id = $1 AND enabled = true`

	rows, err := pr.postgres.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query providers: %w", err)
	}
	defer rows.Close()

	providers := make([]ProviderConfig, 0)
	for rows.Next() {
		var provider ProviderConfig
		var configJSON []byte

		err := rows.Scan(
			&provider.ID, &provider.OrgID, &provider.ProviderName, &provider.ModelName,
			&configJSON, &provider.CostPerTokenPrompt, &provider.CostPerTokenCompletion,
			&provider.QPSLimit, &provider.Enabled, &provider.CreatedAt,
		)
		if err != nil {
			continue
		}

		if err := json.Unmarshal(configJSON, &provider.Config); err != nil {
			continue
		}

		// Filter by quality tier
		if pr.isProviderSuitableForQuality(provider, qualityTier) {
			providers = append(providers, provider)
		}
	}

	return providers, nil
}

// GetAllProviders retrieves all providers for an organization
func (pr *ProviderRouter) GetAllProviders(ctx context.Context, orgID uuid.UUID) ([]ProviderConfig, error) {
	query := `SELECT id, org_id, provider_name, model_name, config, 
			  cost_per_token_prompt, cost_per_token_completion, qps_limit, enabled, created_at
			  FROM provider_config 
			  WHERE org_id = $1`

	rows, err := pr.postgres.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query providers: %w", err)
	}
	defer rows.Close()

	providers := make([]ProviderConfig, 0)
	for rows.Next() {
		var provider ProviderConfig
		var configJSON []byte

		err := rows.Scan(
			&provider.ID, &provider.OrgID, &provider.ProviderName, &provider.ModelName,
			&configJSON, &provider.CostPerTokenPrompt, &provider.CostPerTokenCompletion,
			&provider.QPSLimit, &provider.Enabled, &provider.CreatedAt,
		)
		if err != nil {
			continue
		}

		if err := json.Unmarshal(configJSON, &provider.Config); err != nil {
			continue
		}

		providers = append(providers, provider)
	}

	return providers, nil
}

// SelectOptimalProvider selects the best provider based on multiple criteria
func (pr *ProviderRouter) SelectOptimalProvider(ctx context.Context, req *RoutingRequest, providers []ProviderConfig, budgetStatus *BudgetStatus) (*RoutingResponse, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	// Score each provider
	scoredProviders := make([]ScoredProvider, 0, len(providers))
	for _, provider := range providers {
		score, reason := pr.scoreProvider(ctx, provider, req, budgetStatus)
		scoredProviders = append(scoredProviders, ScoredProvider{
			Provider: provider,
			Score:    score,
			Reason:   reason,
		})
	}

	// Sort by score (highest first)
	sort.Slice(scoredProviders, func(i, j int) bool {
		return scoredProviders[i].Score > scoredProviders[j].Score
	})

	// Use bandit algorithm for exploration vs exploitation
	selectedProvider := pr.bandit.SelectProvider(ctx, scoredProviders)

	// Calculate estimated cost and latency
	estimatedCost := pr.estimateCost(selectedProvider.Provider, req.PromptTokens, req.MaxTokens)
	estimatedLatency := pr.estimateLatency(ctx, selectedProvider.Provider)

	// Generate alternatives
	alternatives := make([]Alternative, 0)
	for i := 1; i < len(scoredProviders) && i < 3; i++ { // Top 3 alternatives
		alt := scoredProviders[i]
		alternatives = append(alternatives, Alternative{
			ProviderName:     alt.Provider.ProviderName,
			ModelName:        alt.Provider.ModelName,
			EstimatedCost:    pr.estimateCost(alt.Provider, req.PromptTokens, req.MaxTokens),
			EstimatedLatency: pr.estimateLatency(ctx, alt.Provider),
			QualityScore:     pr.getQualityScore(alt.Provider, req.QualityTier),
			Reason:           alt.Reason,
		})
	}

	return &RoutingResponse{
		ProviderName:     selectedProvider.Provider.ProviderName,
		ModelName:        selectedProvider.Provider.ModelName,
		Config:           selectedProvider.Provider.Config,
		EstimatedCost:    estimatedCost,
		EstimatedLatency: estimatedLatency,
		Confidence:       selectedProvider.Score,
		Reason:           selectedProvider.Reason,
		Alternatives:     alternatives,
	}, nil
}

// UpdateProviderConfig updates provider configuration
func (pr *ProviderRouter) UpdateProviderConfig(ctx context.Context, orgID uuid.UUID, providerName, modelName string, config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	query := `UPDATE provider_config SET config = $1 
			  WHERE org_id = $2 AND provider_name = $3 AND model_name = $4`

	_, err = pr.postgres.ExecContext(ctx, query, configJSON, orgID, providerName, modelName)
	if err != nil {
		return fmt.Errorf("failed to update provider config: %w", err)
	}

	return nil
}

// GetProviderMetrics retrieves performance metrics for providers
func (pr *ProviderRouter) GetProviderMetrics(ctx context.Context, orgID uuid.UUID, timeRange time.Duration) ([]ProviderMetrics, error) {
	// Mock implementation - in production would query actual metrics from AOS
	providers, err := pr.GetAllProviders(ctx, orgID)
	if err != nil {
		return nil, err
	}

	metrics := make([]ProviderMetrics, 0, len(providers))
	for _, provider := range providers {
		metric := ProviderMetrics{
			ProviderName:     provider.ProviderName,
			ModelName:        provider.ModelName,
			AvgLatency:       pr.estimateLatency(ctx, provider),
			P95Latency:       pr.estimateLatency(ctx, provider) * 2, // Mock P95
			SuccessRate:      0.95 + (float64(time.Now().UnixNano()%10) / 100), // Mock success rate
			AvgCostPerToken:  (provider.CostPerTokenPrompt + provider.CostPerTokenCompletion) / 2,
			QualityScore:     pr.getQualityScore(provider, QualityGold),
			ReliabilityScore: 0.9 + (float64(time.Now().UnixNano()%10) / 100),
			LastUpdated:      time.Now(),
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// Helper methods

type ScoredProvider struct {
	Provider ProviderConfig
	Score    float64
	Reason   string
}

func (pr *ProviderRouter) isProviderSuitableForQuality(provider ProviderConfig, qualityTier QualityTier) bool {
	// Simple quality mapping - in production would be more sophisticated
	qualityScore := pr.getQualityScore(provider, qualityTier)
	
	switch qualityTier {
	case QualityGold:
		return qualityScore >= 0.8
	case QualitySilver:
		return qualityScore >= 0.6
	case QualityBronze:
		return qualityScore >= 0.4
	default:
		return true
	}
}

func (pr *ProviderRouter) getQualityScore(provider ProviderConfig, qualityTier QualityTier) float64 {
	// Mock quality scoring based on provider and model
	baseScore := 0.5

	// Provider-specific scoring
	switch provider.ProviderName {
	case "openai":
		baseScore = 0.9
	case "anthropic":
		baseScore = 0.85
	case "google":
		baseScore = 0.8
	case "cohere":
		baseScore = 0.75
	default:
		baseScore = 0.6
	}

	// Model-specific adjustments
	if provider.ModelName == "gpt-4" || provider.ModelName == "claude-3-opus" {
		baseScore += 0.1
	} else if provider.ModelName == "gpt-3.5-turbo" || provider.ModelName == "claude-3-sonnet" {
		baseScore += 0.05
	}

	// Clamp to [0, 1]
	if baseScore > 1.0 {
		baseScore = 1.0
	}
	if baseScore < 0.0 {
		baseScore = 0.0
	}

	return baseScore
}

func (pr *ProviderRouter) scoreProvider(ctx context.Context, provider ProviderConfig, req *RoutingRequest, budgetStatus *BudgetStatus) (float64, string) {
	score := 0.0
	reasons := make([]string, 0)

	// Cost scoring (40% weight)
	estimatedCost := pr.estimateCost(provider, req.PromptTokens, req.MaxTokens)
	costScore := pr.calculateCostScore(estimatedCost, req.BudgetCents, budgetStatus)
	score += costScore * 0.4
	if costScore > 0.8 {
		reasons = append(reasons, "cost-effective")
	}

	// Quality scoring (30% weight)
	qualityScore := pr.getQualityScore(provider, req.QualityTier)
	score += qualityScore * 0.3
	if qualityScore > 0.8 {
		reasons = append(reasons, "high-quality")
	}

	// Latency scoring (20% weight)
	estimatedLatency := pr.estimateLatency(ctx, provider)
	latencyScore := pr.calculateLatencyScore(estimatedLatency, req.LatencySLA)
	score += latencyScore * 0.2
	if latencyScore > 0.8 {
		reasons = append(reasons, "low-latency")
	}

	// Reliability scoring (10% weight)
	reliabilityScore := pr.getReliabilityScore(ctx, provider)
	score += reliabilityScore * 0.1
	if reliabilityScore > 0.9 {
		reasons = append(reasons, "reliable")
	}

	reason := "optimal choice"
	if len(reasons) > 0 {
		reason = fmt.Sprintf("selected for: %v", reasons)
	}

	return score, reason
}

func (pr *ProviderRouter) estimateCost(provider ProviderConfig, promptTokens, maxTokens int) int64 {
	promptCost := float64(promptTokens) * provider.CostPerTokenPrompt
	completionCost := float64(maxTokens) * provider.CostPerTokenCompletion
	totalCost := promptCost + completionCost
	return int64(totalCost * 100) // Convert to cents
}

func (pr *ProviderRouter) estimateLatency(ctx context.Context, provider ProviderConfig) time.Duration {
	// Mock latency estimation - in production would use historical data
	baseLatency := 1000 * time.Millisecond

	switch provider.ProviderName {
	case "openai":
		baseLatency = 800 * time.Millisecond
	case "anthropic":
		baseLatency = 1200 * time.Millisecond
	case "google":
		baseLatency = 600 * time.Millisecond
	case "cohere":
		baseLatency = 900 * time.Millisecond
	}

	// Add some variance
	variance := time.Duration(time.Now().UnixNano()%200) * time.Millisecond
	return baseLatency + variance
}

func (pr *ProviderRouter) calculateCostScore(estimatedCost, budgetCents int64, budgetStatus *BudgetStatus) float64 {
	if budgetCents <= 0 {
		budgetCents = budgetStatus.RemainingCents
	}

	if budgetCents <= 0 {
		return 0.0 // No budget remaining
	}

	// Score based on cost efficiency
	costRatio := float64(estimatedCost) / float64(budgetCents)
	if costRatio > 1.0 {
		return 0.0 // Cost exceeds budget
	}

	// Higher score for lower cost ratio
	return 1.0 - costRatio
}

func (pr *ProviderRouter) calculateLatencyScore(estimatedLatency, slaLatency time.Duration) float64 {
	if slaLatency <= 0 {
		return 0.8 // Default good score if no SLA specified
	}

	if estimatedLatency > slaLatency {
		return 0.0 // Exceeds SLA
	}

	// Higher score for lower latency
	ratio := float64(estimatedLatency) / float64(slaLatency)
	return 1.0 - ratio
}

func (pr *ProviderRouter) getReliabilityScore(ctx context.Context, provider ProviderConfig) float64 {
	// Mock reliability scoring - in production would use historical data
	baseReliability := 0.95

	// Provider-specific reliability
	switch provider.ProviderName {
	case "openai":
		baseReliability = 0.98
	case "anthropic":
		baseReliability = 0.96
	case "google":
		baseReliability = 0.94
	case "cohere":
		baseReliability = 0.92
	}

	return baseReliability
}