package cas

import (
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type QuotaManager struct {
	redis *redis.Client
}

func NewQuotaManager(redisClient *redis.Client) *QuotaManager {
	return &QuotaManager{
		redis: redisClient,
	}
}

// CheckQuota checks current quota status for a provider/model
func (qm *QuotaManager) CheckQuota(ctx context.Context, providerName, modelName string) (*QuotaStatus, error) {
	key := qm.buildQuotaKey(providerName, modelName)
	
	// Get current QPS and concurrent calls
	pipe := qm.redis.Pipeline()
	qpsResult := pipe.Get(ctx, key+":qps")
	concurrentResult := pipe.Get(ctx, key+":concurrent")
	limitResult := pipe.Get(ctx, key+":limit")
	maxConcurrentResult := pipe.Get(ctx, key+":max_concurrent")
	lastResetResult := pipe.Get(ctx, key+":last_reset")
	
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to check quota: %w", err)
	}

	status := &QuotaStatus{
		ProviderName:    providerName,
		ModelName:       modelName,
		CurrentQPS:      0,
		LimitQPS:        100, // Default limit
		ConcurrentCalls: 0,
		MaxConcurrent:   10, // Default max concurrent
		LastReset:       time.Now().Truncate(time.Minute),
		NextReset:       time.Now().Truncate(time.Minute).Add(time.Minute),
	}

	// Parse results
	if qpsStr, err := qpsResult.Result(); err == nil {
		status.CurrentQPS, _ = strconv.Atoi(qpsStr)
	}

	if concurrentStr, err := concurrentResult.Result(); err == nil {
		status.ConcurrentCalls, _ = strconv.Atoi(concurrentStr)
	}

	if limitStr, err := limitResult.Result(); err == nil {
		status.LimitQPS, _ = strconv.Atoi(limitStr)
	}

	if maxConcurrentStr, err := maxConcurrentResult.Result(); err == nil {
		status.MaxConcurrent, _ = strconv.Atoi(maxConcurrentStr)
	}

	if lastResetStr, err := lastResetResult.Result(); err == nil {
		if timestamp, err := strconv.ParseInt(lastResetStr, 10, 64); err == nil {
			status.LastReset = time.Unix(timestamp, 0)
			status.NextReset = status.LastReset.Add(time.Minute)
		}
	}

	return status, nil
}

// ReserveQuota reserves quota for a request
func (qm *QuotaManager) ReserveQuota(ctx context.Context, providerName, modelName string) error {
	key := qm.buildQuotaKey(providerName, modelName)
	
	// Use Lua script for atomic quota reservation
	script := `
		local qps_key = KEYS[1] .. ":qps"
		local concurrent_key = KEYS[1] .. ":concurrent"
		local limit_key = KEYS[1] .. ":limit"
		local max_concurrent_key = KEYS[1] .. ":max_concurrent"
		local last_reset_key = KEYS[1] .. ":last_reset"
		
		local current_time = tonumber(ARGV[1])
		local current_minute = math.floor(current_time / 60) * 60
		
		-- Get current values
		local current_qps = tonumber(redis.call('GET', qps_key) or 0)
		local current_concurrent = tonumber(redis.call('GET', concurrent_key) or 0)
		local limit_qps = tonumber(redis.call('GET', limit_key) or 100)
		local max_concurrent = tonumber(redis.call('GET', max_concurrent_key) or 10)
		local last_reset = tonumber(redis.call('GET', last_reset_key) or 0)
		
		-- Reset counters if minute has changed
		if current_minute > last_reset then
			redis.call('SET', qps_key, 0)
			redis.call('SET', last_reset_key, current_minute)
			current_qps = 0
		end
		
		-- Check limits
		if current_qps >= limit_qps then
			return {0, "QPS limit exceeded"}
		end
		
		if current_concurrent >= max_concurrent then
			return {0, "Concurrent limit exceeded"}
		end
		
		-- Reserve quota
		redis.call('INCR', qps_key)
		redis.call('INCR', concurrent_key)
		redis.call('EXPIRE', qps_key, 120)
		redis.call('EXPIRE', concurrent_key, 300)
		
		return {1, "OK"}
	`

	result, err := qm.redis.Eval(ctx, script, []string{key}, time.Now().Unix()).Result()
	if err != nil {
		return fmt.Errorf("failed to reserve quota: %w", err)
	}

	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) < 2 {
		return fmt.Errorf("unexpected quota script result")
	}

	success, ok := resultSlice[0].(int64)
	if !ok || success != 1 {
		message, _ := resultSlice[1].(string)
		return fmt.Errorf("quota reservation failed: %s", message)
	}

	return nil
}

// ReleaseQuota releases quota after request completion
func (qm *QuotaManager) ReleaseQuota(ctx context.Context, providerName, modelName string) error {
	key := qm.buildQuotaKey(providerName, modelName)
	concurrentKey := key + ":concurrent"
	
	// Decrement concurrent counter
	result, err := qm.redis.Decr(ctx, concurrentKey).Result()
	if err != nil {
		return fmt.Errorf("failed to release quota: %w", err)
	}

	// Ensure counter doesn't go below 0
	if result < 0 {
		qm.redis.Set(ctx, concurrentKey, 0, 5*time.Minute)
	}

	return nil
}

// RecordUsage records actual usage for monitoring
func (qm *QuotaManager) RecordUsage(ctx context.Context, providerName, modelName string, tokensUsed int) error {
	key := qm.buildQuotaKey(providerName, modelName)
	usageKey := key + ":usage"
	
	// Record token usage
	pipe := qm.redis.Pipeline()
	pipe.IncrBy(ctx, usageKey, int64(tokensUsed))
	pipe.Expire(ctx, usageKey, 24*time.Hour)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to record usage: %w", err)
	}

	// Release concurrent quota
	return qm.ReleaseQuota(ctx, providerName, modelName)
}

// SetQuotaLimits sets quota limits for a provider/model
func (qm *QuotaManager) SetQuotaLimits(ctx context.Context, providerName, modelName string, qpsLimit, maxConcurrent int) error {
	key := qm.buildQuotaKey(providerName, modelName)
	
	pipe := qm.redis.Pipeline()
	pipe.Set(ctx, key+":limit", qpsLimit, 0) // No expiration for limits
	pipe.Set(ctx, key+":max_concurrent", maxConcurrent, 0)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set quota limits: %w", err)
	}

	return nil
}

// GetUsageStats retrieves usage statistics
func (qm *QuotaManager) GetUsageStats(ctx context.Context, providerName, modelName string, timeRange time.Duration) (*UsageStats, error) {
	key := qm.buildQuotaKey(providerName, modelName)
	
	// Get current usage
	usageResult, err := qm.redis.Get(ctx, key+":usage").Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get usage stats: %w", err)
	}

	stats := &UsageStats{
		ProviderName: providerName,
		ModelName:    modelName,
		TimeRange:    timeRange,
		TokensUsed:   0,
		RequestCount: 0,
		Timestamp:    time.Now(),
	}

	if usageResult != "" {
		stats.TokensUsed, _ = strconv.ParseInt(usageResult, 10, 64)
	}

	// Get request count (approximate from QPS history)
	qpsResult, err := qm.redis.Get(ctx, key+":qps").Result()
	if err == nil {
		stats.RequestCount, _ = strconv.ParseInt(qpsResult, 10, 64)
	}

	return stats, nil
}

// ResetQuota resets quota counters for a provider/model
func (qm *QuotaManager) ResetQuota(ctx context.Context, providerName, modelName string) error {
	key := qm.buildQuotaKey(providerName, modelName)
	
	pipe := qm.redis.Pipeline()
	pipe.Del(ctx, key+":qps")
	pipe.Del(ctx, key+":concurrent")
	pipe.Del(ctx, key+":usage")
	pipe.Set(ctx, key+":last_reset", time.Now().Unix(), time.Hour)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to reset quota: %w", err)
	}

	return nil
}

// GetAllQuotaStatus retrieves quota status for all providers
func (qm *QuotaManager) GetAllQuotaStatus(ctx context.Context) (map[string]*QuotaStatus, error) {
	// Get all quota keys
	pattern := "quota:*:qps"
	keys, err := qm.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get quota keys: %w", err)
	}

	statuses := make(map[string]*QuotaStatus)
	
	for _, key := range keys {
		// Extract provider and model from key
		// Format: quota:provider:model:qps
		parts := qm.parseQuotaKey(key)
		if len(parts) < 2 {
			continue
		}

		providerName := parts[0]
		modelName := parts[1]
		
		status, err := qm.CheckQuota(ctx, providerName, modelName)
		if err != nil {
			continue
		}

		statusKey := fmt.Sprintf("%s:%s", providerName, modelName)
		statuses[statusKey] = status
	}

	return statuses, nil
}

// Shutdown gracefully shuts down the quota manager
func (qm *QuotaManager) Shutdown(ctx context.Context) error {
	// Clean up any resources if needed
	return nil
}

// Helper methods

func (qm *QuotaManager) buildQuotaKey(providerName, modelName string) string {
	return fmt.Sprintf("quota:%s:%s", providerName, modelName)
}

func (qm *QuotaManager) parseQuotaKey(key string) []string {
	// Parse key format: quota:provider:model:suffix
	if len(key) < 6 || key[:6] != "quota:" {
		return nil
	}

	parts := key[6:] // Remove "quota:" prefix
	// Find the last colon to separate suffix
	lastColon := -1
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == ':' {
			lastColon = i
			break
		}
	}

	if lastColon == -1 {
		return nil
	}

	providerModel := parts[:lastColon]
	// Split provider and model
	colonIndex := -1
	for i, char := range providerModel {
		if char == ':' {
			colonIndex = i
			break
		}
	}

	if colonIndex == -1 {
		return []string{providerModel}
	}

	return []string{providerModel[:colonIndex], providerModel[colonIndex+1:]}
}

// Supporting types

type UsageStats struct {
	ProviderName string        `json:"provider_name"`
	ModelName    string        `json:"model_name"`
	TimeRange    time.Duration `json:"time_range"`
	TokensUsed   int64         `json:"tokens_used"`
	RequestCount int64         `json:"request_count"`
	Timestamp    time.Time     `json:"timestamp"`
}