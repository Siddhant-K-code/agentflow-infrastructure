package cas

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderRouting(t *testing.T) {
	t.Run("RoutingRequest", func(t *testing.T) {
		req := &RoutingRequest{
			OrgID:        uuid.New(),
			QualityTier:  QualityGold,
			PromptTokens: 100,
			MaxTokens:    200,
			LatencySLA:   2 * time.Second,
			BudgetCents:  1000,
			Constraints: map[string]interface{}{
				"region": "us-east-1",
			},
		}

		assert.NotEqual(t, uuid.Nil, req.OrgID)
		assert.Equal(t, QualityGold, req.QualityTier)
		assert.Equal(t, 100, req.PromptTokens)
		assert.Equal(t, 200, req.MaxTokens)
		assert.Equal(t, 2*time.Second, req.LatencySLA)
		assert.Equal(t, int64(1000), req.BudgetCents)
	})

	t.Run("ProviderConfig", func(t *testing.T) {
		config := &ProviderConfig{
			ID:                     uuid.New(),
			OrgID:                  uuid.New(),
			ProviderName:           "openai",
			ModelName:              "gpt-4",
			CostPerTokenPrompt:     0.00003,
			CostPerTokenCompletion: 0.00006,
			QPSLimit:               100,
			Enabled:                true,
			Config: map[string]interface{}{
				"temperature": 0.7,
				"max_tokens":  2000,
			},
		}

		assert.Equal(t, "openai", config.ProviderName)
		assert.Equal(t, "gpt-4", config.ModelName)
		assert.Greater(t, config.CostPerTokenPrompt, 0.0)
		assert.Greater(t, config.CostPerTokenCompletion, 0.0)
		assert.Equal(t, 100, config.QPSLimit)
		assert.True(t, config.Enabled)
	})

	t.Run("RoutingResponse", func(t *testing.T) {
		response := &RoutingResponse{
			ProviderName:     "openai",
			ModelName:        "gpt-4",
			EstimatedCost:    150,
			EstimatedLatency: 1500 * time.Millisecond,
			Confidence:       0.85,
			Reason:           "selected for high quality and low latency",
			Config: map[string]interface{}{
				"temperature": 0.7,
			},
			Alternatives: []Alternative{
				{
					ProviderName:     "anthropic",
					ModelName:        "claude-3-opus",
					EstimatedCost:    120,
					EstimatedLatency: 2000 * time.Millisecond,
					QualityScore:     0.9,
					Reason:           "cheaper alternative with similar quality",
				},
			},
		}

		assert.Equal(t, "openai", response.ProviderName)
		assert.Equal(t, "gpt-4", response.ModelName)
		assert.Equal(t, int64(150), response.EstimatedCost)
		assert.Equal(t, 1500*time.Millisecond, response.EstimatedLatency)
		assert.Greater(t, response.Confidence, 0.0)
		assert.Len(t, response.Alternatives, 1)
	})
}

func TestBudgetManagement(t *testing.T) {
	t.Run("BudgetCreation", func(t *testing.T) {
		budget := &Budget{
			ID:          uuid.New(),
			OrgID:       uuid.New(),
			PeriodType:  PeriodMonthly,
			LimitCents:  100000, // $1000
			SpentCents:  25000,  // $250
			PeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			PeriodEnd:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			CreatedAt:   time.Now(),
		}

		assert.NotEqual(t, uuid.Nil, budget.ID)
		assert.Equal(t, PeriodMonthly, budget.PeriodType)
		assert.Equal(t, int64(100000), budget.LimitCents)
		assert.Equal(t, int64(25000), budget.SpentCents)
		assert.True(t, budget.PeriodEnd.After(budget.PeriodStart))
	})

	t.Run("BudgetStatus", func(t *testing.T) {
		status := &BudgetStatus{
			BudgetID:       uuid.New(),
			LimitCents:     100000,
			SpentCents:     75000,
			RemainingCents: 25000,
			UtilizationPct: 75.0,
			Status:         BudgetStatusWarning,
		}

		assert.Equal(t, int64(100000), status.LimitCents)
		assert.Equal(t, int64(75000), status.SpentCents)
		assert.Equal(t, int64(25000), status.RemainingCents)
		assert.Equal(t, 75.0, status.UtilizationPct)
		assert.Equal(t, BudgetStatusWarning, status.Status)

		// Test status determination logic
		if status.UtilizationPct >= 90 {
			assert.Equal(t, BudgetStatusCritical, status.Status)
		} else if status.UtilizationPct >= 75 {
			assert.Equal(t, BudgetStatusWarning, status.Status)
		} else {
			assert.Equal(t, BudgetStatusHealthy, status.Status)
		}
	})
}

func TestQuotaManagement(t *testing.T) {
	t.Run("QuotaStatus", func(t *testing.T) {
		status := &QuotaStatus{
			ProviderName:    "openai",
			ModelName:       "gpt-4",
			CurrentQPS:      15,
			LimitQPS:        100,
			ConcurrentCalls: 5,
			MaxConcurrent:   20,
			LastReset:       time.Now().Truncate(time.Minute),
			NextReset:       time.Now().Truncate(time.Minute).Add(time.Minute),
		}

		assert.Equal(t, "openai", status.ProviderName)
		assert.Equal(t, "gpt-4", status.ModelName)
		assert.Equal(t, 15, status.CurrentQPS)
		assert.Equal(t, 100, status.LimitQPS)
		assert.Less(t, status.CurrentQPS, status.LimitQPS)
		assert.Less(t, status.ConcurrentCalls, status.MaxConcurrent)
		assert.True(t, status.NextReset.After(status.LastReset))
	})

	t.Run("QuotaUtilization", func(t *testing.T) {
		status := &QuotaStatus{
			CurrentQPS: 80,
			LimitQPS:   100,
		}

		utilization := float64(status.CurrentQPS) / float64(status.LimitQPS)
		assert.Equal(t, 0.8, utilization)

		// Test quota availability
		available := status.CurrentQPS < status.LimitQPS
		assert.True(t, available)
	})
}

func TestCaching(t *testing.T) {
	t.Run("CacheRequest", func(t *testing.T) {
		req := &CacheRequest{
			Key:        "prompt_hash_input_hash",
			PromptHash: "abc123",
			InputHash:  "def456",
			Response: map[string]interface{}{
				"text":   "Cached response",
				"tokens": 50,
			},
			TTL: 1 * time.Hour,
			Policy: CachePolicy{
				Enabled:      true,
				TTL:          1 * time.Hour,
				PrivacyLevel: PrivacyOrg,
			},
		}

		assert.NotEmpty(t, req.Key)
		assert.NotEmpty(t, req.PromptHash)
		assert.NotEmpty(t, req.InputHash)
		assert.NotNil(t, req.Response)
		assert.Equal(t, 1*time.Hour, req.TTL)
		assert.True(t, req.Policy.Enabled)
		assert.Equal(t, PrivacyOrg, req.Policy.PrivacyLevel)
	})

	t.Run("CacheResponse", func(t *testing.T) {
		response := &CacheResponse{
			Hit: true,
			Response: map[string]interface{}{
				"text":   "Cached response",
				"tokens": 50,
			},
			CreatedAt: time.Now().Add(-30 * time.Minute),
			ExpiresAt: time.Now().Add(30 * time.Minute),
		}

		assert.True(t, response.Hit)
		assert.NotNil(t, response.Response)
		assert.True(t, response.ExpiresAt.After(response.CreatedAt))
		assert.True(t, response.ExpiresAt.After(time.Now()))
	})

	t.Run("CachePolicy", func(t *testing.T) {
		policy := CachePolicy{
			Enabled:      true,
			TTL:          2 * time.Hour,
			PrivacyLevel: PrivacyProject,
			Conditions: map[string]interface{}{
				"min_tokens": 10,
				"max_cost":   100,
			},
		}

		assert.True(t, policy.Enabled)
		assert.Equal(t, 2*time.Hour, policy.TTL)
		assert.Equal(t, PrivacyProject, policy.PrivacyLevel)
		assert.NotNil(t, policy.Conditions)
	})
}

func TestMultiArmedBandit(t *testing.T) {
	bandit := NewMultiArmedBandit()

	t.Run("ArmInitialization", func(t *testing.T) {
		// Test that arms are created on first use
		assert.Empty(t, bandit.arms)

		// Simulate provider selection
		providers := []ScoredProvider{
			{
				Provider: ProviderConfig{
					ProviderName: "openai",
					ModelName:    "gpt-4",
				},
				Score:  0.8,
				Reason: "high quality",
			},
			{
				Provider: ProviderConfig{
					ProviderName: "anthropic",
					ModelName:    "claude-3-opus",
				},
				Score:  0.75,
				Reason: "good alternative",
			},
		}

		selected := bandit.SelectProvider(context.Background(), providers)
		assert.NotEmpty(t, selected.Provider.ProviderName)
		assert.NotEmpty(t, bandit.arms)
	})

	t.Run("RewardUpdate", func(t *testing.T) {
		bandit.UpdateReward("openai", "gpt-4", 0.9)
		bandit.UpdateReward("openai", "gpt-4", 0.8)

		stats := bandit.GetArmStats()
		arm, exists := stats["openai:gpt-4"]
		require.True(t, exists)
		assert.Equal(t, 2, arm.Pulls)
		assert.InDelta(t, 1.7, arm.TotalReward, 0.001)
		assert.InDelta(t, 0.85, arm.AverageReward, 0.001)
	})

	t.Run("RewardCalculation", func(t *testing.T) {
		// Test successful execution with accurate estimates
		reward := bandit.CalculateReward(
			100, 100, // actual cost = estimated cost
			1500*time.Millisecond, 1500*time.Millisecond, // actual latency = estimated latency
			true, // success
		)
		assert.Greater(t, reward, 1.0) // Should get bonus for accuracy

		// Test failed execution
		reward = bandit.CalculateReward(100, 100, 1500*time.Millisecond, 1500*time.Millisecond, false)
		assert.Equal(t, -1.0, reward) // Penalty for failure

		// Test cost and latency overrun
		reward = bandit.CalculateReward(
			200, 100, // 100% cost overrun
			3000*time.Millisecond, 1500*time.Millisecond, // 100% latency overrun
			true,
		)
		assert.Less(t, reward, 1.0) // Should be penalized for overruns
	})
}

func TestOptimizationSuggestions(t *testing.T) {
	t.Run("OptimizationSuggestion", func(t *testing.T) {
		suggestion := OptimizationSuggestion{
			Type:            OptimizationProviderSwitch,
			Title:           "Switch to more cost-effective provider",
			Description:     "40% of requests could use cheaper provider with similar quality",
			PotentialSaving: 15000,
			Confidence:      0.85,
			Impact:          ImpactMedium,
			Actions: []string{
				"Configure routing rules",
				"Set up A/B testing",
				"Monitor quality metrics",
			},
			Metadata: map[string]interface{}{
				"current_provider":   "openai",
				"suggested_provider": "anthropic",
			},
		}

		assert.Equal(t, OptimizationProviderSwitch, suggestion.Type)
		assert.NotEmpty(t, suggestion.Title)
		assert.Greater(t, suggestion.PotentialSaving, int64(0))
		assert.Greater(t, suggestion.Confidence, 0.0)
		assert.LessOrEqual(t, suggestion.Confidence, 1.0)
		assert.Equal(t, ImpactMedium, suggestion.Impact)
		assert.Len(t, suggestion.Actions, 3)
	})

	t.Run("OptimizationTypes", func(t *testing.T) {
		types := []OptimizationType{
			OptimizationProviderSwitch,
			OptimizationModelDowngrade,
			OptimizationCaching,
			OptimizationBatching,
			OptimizationScheduling,
		}

		for _, optimizationType := range types {
			assert.NotEmpty(t, string(optimizationType))
		}
	})
}

func TestBatchProcessing(t *testing.T) {
	t.Run("BatchRequest", func(t *testing.T) {
		req := &BatchRequest{
			Operations: []BatchOperation{
				{
					ID:   "op1",
					Type: "llm_call",
					Payload: map[string]interface{}{
						"prompt": "Analyze this text",
						"text":   "Sample text 1",
					},
				},
				{
					ID:   "op2",
					Type: "llm_call",
					Payload: map[string]interface{}{
						"prompt": "Analyze this text",
						"text":   "Sample text 2",
					},
				},
			},
			Policy: BatchPolicy{
				MaxBatchSize:   10,
				MaxWaitTime:    5 * time.Second,
				CompatibleOnly: true,
			},
		}

		assert.Len(t, req.Operations, 2)
		assert.Equal(t, 10, req.Policy.MaxBatchSize)
		assert.Equal(t, 5*time.Second, req.Policy.MaxWaitTime)
		assert.True(t, req.Policy.CompatibleOnly)
	})

	t.Run("BatchResponse", func(t *testing.T) {
		response := &BatchResponse{
			BatchID: "batch_123",
			Results: []BatchResult{
				{
					OperationID: "op1",
					Status:      "success",
					Result: map[string]interface{}{
						"analysis": "Result 1",
					},
				},
				{
					OperationID: "op2",
					Status:      "success",
					Result: map[string]interface{}{
						"analysis": "Result 2",
					},
				},
			},
			Summary: BatchSummary{
				TotalOperations: 2,
				SuccessCount:    2,
				FailureCount:    0,
				TotalCostCents:  200,
				TotalSavings:    40,
			},
			CreatedAt: time.Now(),
		}

		assert.Equal(t, "batch_123", response.BatchID)
		assert.Len(t, response.Results, 2)
		assert.Equal(t, 2, response.Summary.TotalOperations)
		assert.Equal(t, 2, response.Summary.SuccessCount)
		assert.Equal(t, 0, response.Summary.FailureCount)
		assert.Equal(t, int64(200), response.Summary.TotalCostCents)
		assert.Equal(t, int64(40), response.Summary.TotalSavings)
	})
}

// Benchmark tests
func BenchmarkProviderSelection(b *testing.B) {
	bandit := NewMultiArmedBandit()
	providers := []ScoredProvider{
		{
			Provider: ProviderConfig{ProviderName: "openai", ModelName: "gpt-4"},
			Score:    0.8,
		},
		{
			Provider: ProviderConfig{ProviderName: "anthropic", ModelName: "claude-3-opus"},
			Score:    0.75,
		},
		{
			Provider: ProviderConfig{ProviderName: "google", ModelName: "gemini-pro"},
			Score:    0.7,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bandit.SelectProvider(context.Background(), providers)
	}
}

func BenchmarkCostCalculation(b *testing.B) {
	config := &ProviderConfig{
		CostPerTokenPrompt:     0.00003,
		CostPerTokenCompletion: 0.00006,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		promptCost := float64(100) * config.CostPerTokenPrompt
		completionCost := float64(200) * config.CostPerTokenCompletion
		_ = int64((promptCost + completionCost) * 100)
	}
}
