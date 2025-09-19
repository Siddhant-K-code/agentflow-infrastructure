package cas

import (
	"fmt"
	"time"

	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/config"
	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/db"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	cfg        *config.Config
	postgres   *db.PostgresDB
	redis      *redis.Client
	router     *ProviderRouter
	budgetMgr  *BudgetManager
	cache      *CacheManager
	quotaMgr   *QuotaManager
	optimizer  *Optimizer
}

func NewService(cfg *config.Config, pg *db.PostgresDB, redisClient *redis.Client) *Service {
	service := &Service{
		cfg:      cfg,
		postgres: pg,
		redis:    redisClient,
	}

	service.router = NewProviderRouter(pg, redisClient)
	service.budgetMgr = NewBudgetManager(pg)
	service.cache = NewCacheManager(redisClient)
	service.quotaMgr = NewQuotaManager(redisClient)
	service.optimizer = NewOptimizer(pg, redisClient)

	return service
}

// RouteRequest routes a request to the optimal provider/model
func (s *Service) RouteRequest(ctx context.Context, req *RoutingRequest) (*RoutingResponse, error) {
	// Check budget constraints
	budgetStatus, err := s.budgetMgr.CheckBudget(ctx, req.OrgID, req.BudgetCents)
	if err != nil {
		return nil, fmt.Errorf("failed to check budget: %w", err)
	}

	if budgetStatus.Status == BudgetStatusExceeded {
		return nil, fmt.Errorf("budget exceeded: %d/%d cents used", budgetStatus.SpentCents, budgetStatus.LimitCents)
	}

	// Get available providers
	providers, err := s.router.GetAvailableProviders(ctx, req.OrgID, req.QualityTier)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}

	// Filter by quota availability
	availableProviders := make([]ProviderConfig, 0)
	for _, provider := range providers {
		quotaStatus, err := s.quotaMgr.CheckQuota(ctx, provider.ProviderName, provider.ModelName)
		if err != nil {
			continue // Skip providers with quota check errors
		}

		if quotaStatus.CurrentQPS < quotaStatus.LimitQPS {
			availableProviders = append(availableProviders, provider)
		}
	}

	if len(availableProviders) == 0 {
		return nil, fmt.Errorf("no available providers within quota limits")
	}

	// Route to optimal provider
	response, err := s.router.SelectOptimalProvider(ctx, req, availableProviders, budgetStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to select provider: %w", err)
	}

	// Reserve quota
	if err := s.quotaMgr.ReserveQuota(ctx, response.ProviderName, response.ModelName); err != nil {
		// Log warning but don't fail the request
		fmt.Printf("Warning: failed to reserve quota for %s/%s: %v\n", response.ProviderName, response.ModelName, err)
	}

	return response, nil
}

// CacheGet retrieves a cached response
func (s *Service) CacheGet(ctx context.Context, orgID uuid.UUID, promptHash, inputHash string) (*CacheResponse, error) {
	return s.cache.Get(ctx, orgID, promptHash, inputHash)
}

// CachePut stores a response in cache
func (s *Service) CachePut(ctx context.Context, orgID uuid.UUID, req *CacheRequest) error {
	return s.cache.Put(ctx, orgID, req)
}

// RecordUsage records actual usage for budget and quota tracking
func (s *Service) RecordUsage(ctx context.Context, orgID uuid.UUID, providerName, modelName string, costCents int64, tokensUsed int) error {
	// Update budget
	if err := s.budgetMgr.RecordSpending(ctx, orgID, costCents); err != nil {
		return fmt.Errorf("failed to record budget spending: %w", err)
	}

	// Update quota
	if err := s.quotaMgr.RecordUsage(ctx, providerName, modelName, tokensUsed); err != nil {
		return fmt.Errorf("failed to record quota usage: %w", err)
	}

	return nil
}

// GetBudgetStatus retrieves current budget status
func (s *Service) GetBudgetStatus(ctx context.Context, orgID uuid.UUID) (*BudgetStatus, error) {
	return s.budgetMgr.GetStatus(ctx, orgID)
}

// GetQuotaStatus retrieves current quota status for all providers
func (s *Service) GetQuotaStatus(ctx context.Context, orgID uuid.UUID) ([]QuotaStatus, error) {
	providers, err := s.router.GetAllProviders(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}

	statuses := make([]QuotaStatus, 0, len(providers))
	for _, provider := range providers {
		status, err := s.quotaMgr.CheckQuota(ctx, provider.ProviderName, provider.ModelName)
		if err != nil {
			continue // Skip providers with errors
		}
		statuses = append(statuses, *status)
	}

	return statuses, nil
}

// GetOptimizationSuggestions provides cost optimization recommendations
func (s *Service) GetOptimizationSuggestions(ctx context.Context, orgID uuid.UUID, timeRange time.Duration) ([]OptimizationSuggestion, error) {
	return s.optimizer.GenerateSuggestions(ctx, orgID, timeRange)
}

// CreateBudget creates a new budget
func (s *Service) CreateBudget(ctx context.Context, orgID uuid.UUID, periodType PeriodType, limitCents int64, projectID *uuid.UUID) (*Budget, error) {
	budget := &Budget{
		ID:          uuid.New(),
		OrgID:       orgID,
		ProjectID:   projectID,
		PeriodType:  periodType,
		LimitCents:  limitCents,
		SpentCents:  0,
		CreatedAt:   time.Now(),
	}

	// Set period boundaries
	now := time.Now()
	switch periodType {
	case PeriodDaily:
		budget.PeriodStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		budget.PeriodEnd = budget.PeriodStart.Add(24 * time.Hour)
	case PeriodWeekly:
		// Start of week (Monday)
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday = 7
		}
		daysToMonday := weekday - 1
		budget.PeriodStart = time.Date(now.Year(), now.Month(), now.Day()-daysToMonday, 0, 0, 0, 0, now.Location())
		budget.PeriodEnd = budget.PeriodStart.Add(7 * 24 * time.Hour)
	case PeriodMonthly:
		budget.PeriodStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		budget.PeriodEnd = budget.PeriodStart.AddDate(0, 1, 0)
	}

	return s.budgetMgr.CreateBudget(ctx, budget)
}

// UpdateProviderConfig updates provider configuration
func (s *Service) UpdateProviderConfig(ctx context.Context, orgID uuid.UUID, providerName, modelName string, config map[string]interface{}) error {
	return s.router.UpdateProviderConfig(ctx, orgID, providerName, modelName, config)
}

// GetProviderMetrics retrieves performance metrics for providers
func (s *Service) GetProviderMetrics(ctx context.Context, orgID uuid.UUID, timeRange time.Duration) ([]ProviderMetrics, error) {
	return s.router.GetProviderMetrics(ctx, orgID, timeRange)
}

// ProcessBatch processes a batch of operations for cost optimization
func (s *Service) ProcessBatch(ctx context.Context, orgID uuid.UUID, req *BatchRequest) (*BatchResponse, error) {
	batchID := uuid.New().String()
	
	response := &BatchResponse{
		BatchID:   batchID,
		Results:   make([]BatchResult, 0, len(req.Operations)),
		CreatedAt: time.Now(),
	}

	// Group compatible operations
	batches := s.groupCompatibleOperations(req.Operations, req.Policy)

	totalCost := int64(0)
	totalSavings := int64(0)

	for _, batch := range batches {
		batchResult, cost, savings := s.processBatchGroup(ctx, orgID, batch)
		response.Results = append(response.Results, batchResult...)
		totalCost += cost
		totalSavings += savings
	}

	// Calculate summary
	response.Summary = BatchSummary{
		TotalOperations: len(req.Operations),
		TotalCostCents:  totalCost,
		TotalSavings:    totalSavings,
	}

	for _, result := range response.Results {
		if result.Status == "success" {
			response.Summary.SuccessCount++
		} else {
			response.Summary.FailureCount++
		}
	}

	return response, nil
}

// Helper methods

func (s *Service) groupCompatibleOperations(operations []BatchOperation, policy BatchPolicy) [][]BatchOperation {
	if !policy.CompatibleOnly {
		// Simple batching by max size
		batches := make([][]BatchOperation, 0)
		for i := 0; i < len(operations); i += policy.MaxBatchSize {
			end := i + policy.MaxBatchSize
			if end > len(operations) {
				end = len(operations)
			}
			batches = append(batches, operations[i:end])
		}
		return batches
	}

	// Group by compatibility (simplified - would be more sophisticated in production)
	compatible := make(map[string][]BatchOperation)
	for _, op := range operations {
		key := op.Type // Simple grouping by type
		if _, exists := compatible[key]; !exists {
			compatible[key] = make([]BatchOperation, 0)
		}
		compatible[key] = append(compatible[key], op)
	}

	batches := make([][]BatchOperation, 0)
	for _, group := range compatible {
		for i := 0; i < len(group); i += policy.MaxBatchSize {
			end := i + policy.MaxBatchSize
			if end > len(group) {
				end = len(group)
			}
			batches = append(batches, group[i:end])
		}
	}

	return batches
}

func (s *Service) processBatchGroup(ctx context.Context, orgID uuid.UUID, operations []BatchOperation) ([]BatchResult, int64, int64) {
	results := make([]BatchResult, 0, len(operations))
	totalCost := int64(0)
	totalSavings := int64(0)

	// Mock batch processing - in production would actually execute operations
	for _, op := range operations {
		result := BatchResult{
			OperationID: op.ID,
			Status:      "success",
			Result: map[string]interface{}{
				"processed": true,
				"batch_id":  fmt.Sprintf("batch_%d", len(operations)),
			},
		}

		// Mock cost calculation
		cost := int64(100) // Base cost per operation
		savings := int64(20) // Savings from batching

		totalCost += cost
		totalSavings += savings

		results = append(results, result)
	}

	return results, totalCost, totalSavings
}

// Shutdown gracefully shuts down the service
func (s *Service) Shutdown(ctx context.Context) error {
	// Flush any pending operations
	if s.cache != nil {
		s.cache.Flush(ctx)
	}

	if s.quotaMgr != nil {
		s.quotaMgr.Shutdown(ctx)
	}

	return nil
}