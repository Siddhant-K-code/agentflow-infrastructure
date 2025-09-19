package cas

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// MultiArmedBandit implements Upper Confidence Bound (UCB) algorithm for provider selection
type MultiArmedBandit struct {
	arms map[string]*BanditArm
	c    float64 // Exploration parameter
}

type BanditArm struct {
	ProviderKey   string    // provider:model
	Pulls         int       // Number of times selected
	TotalReward   float64   // Sum of rewards
	AverageReward float64   // Average reward
	LastPull      time.Time // Last time this arm was pulled
}

func NewMultiArmedBandit() *MultiArmedBandit {
	return &MultiArmedBandit{
		arms: make(map[string]*BanditArm),
		c:    1.414, // sqrt(2), common choice for UCB
	}
}

// SelectProvider selects a provider using UCB algorithm
func (mab *MultiArmedBandit) SelectProvider(ctx context.Context, scoredProviders []ScoredProvider) ScoredProvider {
	if len(scoredProviders) == 0 {
		panic("no providers available")
	}

	if len(scoredProviders) == 1 {
		return scoredProviders[0]
	}

	totalPulls := mab.getTotalPulls()
	
	// If we haven't tried all arms yet, try untried ones first
	for _, provider := range scoredProviders {
		key := mab.getProviderKey(provider.Provider)
		if _, exists := mab.arms[key]; !exists {
			mab.initializeArm(key)
			return provider
		}
	}

	// Calculate UCB values for each provider
	bestProvider := scoredProviders[0]
	bestUCB := -math.Inf(1)

	for _, provider := range scoredProviders {
		key := mab.getProviderKey(provider.Provider)
		arm := mab.arms[key]
		
		if arm.Pulls == 0 {
			// Untried arm gets highest priority
			return provider
		}

		// Calculate UCB value
		ucb := arm.AverageReward + mab.c*math.Sqrt(math.Log(float64(totalPulls))/float64(arm.Pulls))
		
		// Add base score from routing algorithm
		ucb += provider.Score * 0.3 // Weight the routing score
		
		if ucb > bestUCB {
			bestUCB = ucb
			bestProvider = provider
		}
	}

	// Update the selected arm
	key := mab.getProviderKey(bestProvider.Provider)
	mab.arms[key].Pulls++
	mab.arms[key].LastPull = time.Now()

	return bestProvider
}

// UpdateReward updates the reward for a provider based on performance
func (mab *MultiArmedBandit) UpdateReward(providerName, modelName string, reward float64) {
	key := providerName + ":" + modelName
	arm, exists := mab.arms[key]
	if !exists {
		mab.initializeArm(key)
		arm = mab.arms[key]
	}

	arm.Pulls++
	arm.TotalReward += reward
	arm.AverageReward = arm.TotalReward / float64(arm.Pulls)
}

// CalculateReward calculates reward based on performance metrics
func (mab *MultiArmedBandit) CalculateReward(actualCost, estimatedCost int64, actualLatency, estimatedLatency time.Duration, success bool) float64 {
	if !success {
		return -1.0 // Penalty for failures
	}

	reward := 1.0 // Base reward for success

	// Cost accuracy bonus/penalty
	if estimatedCost > 0 {
		costAccuracy := 1.0 - math.Abs(float64(actualCost-estimatedCost))/float64(estimatedCost)
		reward += (costAccuracy - 0.5) * 0.6 // Penalty for inaccuracy
	}

	// Latency accuracy bonus/penalty
	if estimatedLatency > 0 {
		latencyAccuracy := 1.0 - math.Abs(float64(actualLatency-estimatedLatency))/float64(estimatedLatency)
		reward += (latencyAccuracy - 0.5) * 0.6 // Penalty for inaccuracy
	}

	// Cost efficiency bonus
	if actualCost < estimatedCost {
		savings := float64(estimatedCost-actualCost) / float64(estimatedCost)
		reward += savings * 0.2
	}

	// Latency efficiency bonus
	if actualLatency < estimatedLatency {
		speedup := float64(estimatedLatency-actualLatency) / float64(estimatedLatency)
		reward += speedup * 0.2
	}

	// Clamp reward to reasonable range
	if reward > 2.0 {
		reward = 2.0
	}
	if reward < -1.0 {
		reward = -1.0
	}

	return reward
}

// GetArmStats returns statistics for all arms
func (mab *MultiArmedBandit) GetArmStats() map[string]BanditArm {
	stats := make(map[string]BanditArm)
	for key, arm := range mab.arms {
		stats[key] = *arm
	}
	return stats
}

// ResetArm resets statistics for a specific provider
func (mab *MultiArmedBandit) ResetArm(providerName, modelName string) {
	key := providerName + ":" + modelName
	if arm, exists := mab.arms[key]; exists {
		arm.Pulls = 0
		arm.TotalReward = 0
		arm.AverageReward = 0
		arm.LastPull = time.Time{}
	}
}

// DecayRewards applies time-based decay to rewards to adapt to changing conditions
func (mab *MultiArmedBandit) DecayRewards(decayFactor float64) {
	for _, arm := range mab.arms {
		arm.TotalReward *= decayFactor
		if arm.Pulls > 0 {
			arm.AverageReward = arm.TotalReward / float64(arm.Pulls)
		}
	}
}

// Helper methods

func (mab *MultiArmedBandit) getProviderKey(provider ProviderConfig) string {
	return provider.ProviderName + ":" + provider.ModelName
}

func (mab *MultiArmedBandit) initializeArm(key string) {
	mab.arms[key] = &BanditArm{
		ProviderKey:   key,
		Pulls:         0,
		TotalReward:   0,
		AverageReward: 0,
		LastPull:      time.Time{},
	}
}

func (mab *MultiArmedBandit) getTotalPulls() int {
	total := 0
	for _, arm := range mab.arms {
		total += arm.Pulls
	}
	return total
}

// EpsilonGreedy implements epsilon-greedy algorithm as an alternative to UCB
type EpsilonGreedy struct {
	arms    map[string]*BanditArm
	epsilon float64 // Exploration probability
}

func NewEpsilonGreedy(epsilon float64) *EpsilonGreedy {
	return &EpsilonGreedy{
		arms:    make(map[string]*BanditArm),
		epsilon: epsilon,
	}
}

func (eg *EpsilonGreedy) SelectProvider(ctx context.Context, scoredProviders []ScoredProvider) ScoredProvider {
	if len(scoredProviders) == 0 {
		panic("no providers available")
	}

	if len(scoredProviders) == 1 {
		return scoredProviders[0]
	}

	// Epsilon-greedy selection
	if rand.Float64() < eg.epsilon {
		// Explore: random selection
		return scoredProviders[rand.Intn(len(scoredProviders))]
	}

	// Exploit: select best performing provider
	bestProvider := scoredProviders[0]
	bestReward := -math.Inf(1)

	for _, provider := range scoredProviders {
		key := provider.Provider.ProviderName + ":" + provider.Provider.ModelName
		arm, exists := eg.arms[key]
		
		reward := provider.Score // Use routing score as baseline
		if exists && arm.Pulls > 0 {
			reward = arm.AverageReward
		}

		if reward > bestReward {
			bestReward = reward
			bestProvider = provider
		}
	}

	// Update arm statistics
	key := bestProvider.Provider.ProviderName + ":" + bestProvider.Provider.ModelName
	if _, exists := eg.arms[key]; !exists {
		eg.arms[key] = &BanditArm{
			ProviderKey: key,
		}
	}
	eg.arms[key].Pulls++
	eg.arms[key].LastPull = time.Now()

	return bestProvider
}

func (eg *EpsilonGreedy) UpdateReward(providerName, modelName string, reward float64) {
	key := providerName + ":" + modelName
	arm, exists := eg.arms[key]
	if !exists {
		eg.arms[key] = &BanditArm{
			ProviderKey: key,
		}
		arm = eg.arms[key]
	}

	arm.TotalReward += reward
	if arm.Pulls > 0 {
		arm.AverageReward = arm.TotalReward / float64(arm.Pulls)
	}
}