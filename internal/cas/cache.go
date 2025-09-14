package cas

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type CacheManager struct {
	redis *redis.Client
}

func NewCacheManager(redisClient *redis.Client) *CacheManager {
	return &CacheManager{
		redis: redisClient,
	}
}

// Get retrieves a cached response
func (cm *CacheManager) Get(ctx context.Context, orgID uuid.UUID, promptHash, inputHash string) (*CacheResponse, error) {
	key := cm.buildCacheKey(orgID, promptHash, inputHash)
	
	result, err := cm.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return &CacheResponse{Hit: false}, nil
		}
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var cachedData CachedData
	if err := json.Unmarshal([]byte(result), &cachedData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	// Check if expired
	if time.Now().After(cachedData.ExpiresAt) {
		// Remove expired entry
		cm.redis.Del(ctx, key)
		return &CacheResponse{Hit: false}, nil
	}

	return &CacheResponse{
		Hit:       true,
		Response:  cachedData.Response,
		CreatedAt: cachedData.CreatedAt,
		ExpiresAt: cachedData.ExpiresAt,
	}, nil
}

// Put stores a response in cache
func (cm *CacheManager) Put(ctx context.Context, orgID uuid.UUID, req *CacheRequest) error {
	if !req.Policy.Enabled {
		return nil // Caching disabled
	}

	key := cm.buildCacheKey(orgID, req.PromptHash, req.InputHash)
	
	// Check privacy level
	if !cm.isPrivacyLevelAllowed(req.Policy.PrivacyLevel, orgID) {
		return fmt.Errorf("privacy level %s not allowed for caching", req.Policy.PrivacyLevel)
	}

	cachedData := CachedData{
		Response:  req.Response,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(req.TTL),
		Policy:    req.Policy,
	}

	data, err := json.Marshal(cachedData)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	// Store with TTL
	err = cm.redis.Set(ctx, key, data, req.TTL).Err()
	if err != nil {
		return fmt.Errorf("failed to store in cache: %w", err)
	}

	// Update cache statistics
	cm.updateCacheStats(ctx, orgID, "put")

	return nil
}

// Delete removes an entry from cache
func (cm *CacheManager) Delete(ctx context.Context, orgID uuid.UUID, promptHash, inputHash string) error {
	key := cm.buildCacheKey(orgID, promptHash, inputHash)
	return cm.redis.Del(ctx, key).Err()
}

// Clear clears all cache entries for an organization
func (cm *CacheManager) Clear(ctx context.Context, orgID uuid.UUID) error {
	pattern := fmt.Sprintf("cache:%s:*", orgID.String())
	
	keys, err := cm.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get cache keys: %w", err)
	}

	if len(keys) > 0 {
		err = cm.redis.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete cache keys: %w", err)
		}
	}

	return nil
}

// GetStats retrieves cache statistics
func (cm *CacheManager) GetStats(ctx context.Context, orgID uuid.UUID) (*CacheStats, error) {
	statsKey := fmt.Sprintf("cache_stats:%s", orgID.String())
	
	result, err := cm.redis.HGetAll(ctx, statsKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache stats: %w", err)
	}

	stats := &CacheStats{
		OrgID: orgID,
	}

	if hitStr, exists := result["hits"]; exists {
		fmt.Sscanf(hitStr, "%d", &stats.Hits)
	}

	if missStr, exists := result["misses"]; exists {
		fmt.Sscanf(missStr, "%d", &stats.Misses)
	}

	if putStr, exists := result["puts"]; exists {
		fmt.Sscanf(putStr, "%d", &stats.Puts)
	}

	// Calculate hit rate
	total := stats.Hits + stats.Misses
	if total > 0 {
		stats.HitRate = float64(stats.Hits) / float64(total)
	}

	// Get cache size
	pattern := fmt.Sprintf("cache:%s:*", orgID.String())
	keys, err := cm.redis.Keys(ctx, pattern).Result()
	if err == nil {
		stats.Size = int64(len(keys))
	}

	return stats, nil
}

// Flush flushes pending cache operations
func (cm *CacheManager) Flush(ctx context.Context) error {
	// Redis operations are already atomic, so nothing to flush
	return nil
}

// GenerateHash generates a hash for cache key
func (cm *CacheManager) GenerateHash(content interface{}) (string, error) {
	data, err := json.Marshal(content)
	if err != nil {
		return "", fmt.Errorf("failed to marshal content: %w", err)
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// GetCachePolicy returns the appropriate cache policy for a request
func (cm *CacheManager) GetCachePolicy(orgID uuid.UUID, qualityTier QualityTier, promptType string) CachePolicy {
	policy := CachePolicy{
		Enabled:      true,
		TTL:          1 * time.Hour, // Default TTL
		PrivacyLevel: PrivacyOrg,    // Default privacy level
	}

	// Adjust based on quality tier
	switch qualityTier {
	case QualityGold:
		policy.TTL = 30 * time.Minute // Shorter TTL for high quality
	case QualitySilver:
		policy.TTL = 1 * time.Hour
	case QualityBronze:
		policy.TTL = 4 * time.Hour // Longer TTL for lower quality
	}

	// Adjust based on prompt type
	switch promptType {
	case "system":
		policy.TTL = 24 * time.Hour // System prompts can be cached longer
	case "user":
		policy.TTL = 15 * time.Minute // User prompts should be cached shorter
	case "sensitive":
		policy.Enabled = false // Don't cache sensitive prompts
	}

	return policy
}

// Helper methods

func (cm *CacheManager) buildCacheKey(orgID uuid.UUID, promptHash, inputHash string) string {
	return fmt.Sprintf("cache:%s:%s:%s", orgID.String(), promptHash, inputHash)
}

func (cm *CacheManager) isPrivacyLevelAllowed(level PrivacyLevel, orgID uuid.UUID) bool {
	// Simple privacy level validation
	// In production, would check against organization policies
	switch level {
	case PrivacyPublic:
		return false // Don't allow public caching for now
	case PrivacyOrg, PrivacyProject, PrivacyUser:
		return true
	default:
		return false
	}
}

func (cm *CacheManager) updateCacheStats(ctx context.Context, orgID uuid.UUID, operation string) {
	statsKey := fmt.Sprintf("cache_stats:%s", orgID.String())
	
	switch operation {
	case "hit":
		cm.redis.HIncrBy(ctx, statsKey, "hits", 1)
	case "miss":
		cm.redis.HIncrBy(ctx, statsKey, "misses", 1)
	case "put":
		cm.redis.HIncrBy(ctx, statsKey, "puts", 1)
	}

	// Set expiration for stats (reset monthly)
	cm.redis.Expire(ctx, statsKey, 30*24*time.Hour)
}

// Supporting types

type CachedData struct {
	Response  map[string]interface{} `json:"response"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt time.Time              `json:"expires_at"`
	Policy    CachePolicy            `json:"policy"`
}

type CacheStats struct {
	OrgID   uuid.UUID `json:"org_id"`
	Hits    int64     `json:"hits"`
	Misses  int64     `json:"misses"`
	Puts    int64     `json:"puts"`
	HitRate float64   `json:"hit_rate"`
	Size    int64     `json:"size"`
}

// CacheWarmer pre-populates cache with common responses
type CacheWarmer struct {
	cache *CacheManager
}

func NewCacheWarmer(cache *CacheManager) *CacheWarmer {
	return &CacheWarmer{cache: cache}
}

// WarmCache pre-populates cache with common prompt responses
func (cw *CacheWarmer) WarmCache(ctx context.Context, orgID uuid.UUID, commonPrompts []WarmupPrompt) error {
	for _, prompt := range commonPrompts {
		// Generate mock response for warming
		response := map[string]interface{}{
			"text":         fmt.Sprintf("Warmed response for %s", prompt.Name),
			"tokens":       prompt.EstimatedTokens,
			"warmed":       true,
			"warmed_at":    time.Now(),
		}

		req := &CacheRequest{
			PromptHash: prompt.Hash,
			InputHash:  prompt.InputHash,
			Response:   response,
			TTL:        prompt.TTL,
			Policy:     prompt.Policy,
		}

		if err := cw.cache.Put(ctx, orgID, req); err != nil {
			// Log error but continue warming other prompts
			fmt.Printf("Failed to warm cache for prompt %s: %v\n", prompt.Name, err)
		}
	}

	return nil
}

type WarmupPrompt struct {
	Name             string                 `json:"name"`
	Hash             string                 `json:"hash"`
	InputHash        string                 `json:"input_hash"`
	EstimatedTokens  int                    `json:"estimated_tokens"`
	TTL              time.Duration          `json:"ttl"`
	Policy           CachePolicy            `json:"policy"`
}