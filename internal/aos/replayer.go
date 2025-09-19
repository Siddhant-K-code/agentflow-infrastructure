package aos

import (
	"fmt"
	"time"

	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/db"
)

type Replayer struct {
	postgres   *db.PostgresDB
	clickhouse *db.ClickHouseDB
}

func NewReplayer(pg *db.PostgresDB, ch *db.ClickHouseDB) *Replayer {
	return &Replayer{
		postgres:   pg,
		clickhouse: ch,
	}
}

// Replay executes a replay of a workflow run
func (r *Replayer) Replay(ctx context.Context, orgID uuid.UUID, req *ReplayRequest, originalEvents []TraceEvent) (*ReplayResponse, error) {
	replayRunID := uuid.New()

	response := &ReplayResponse{
		ReplayRunID: replayRunID,
		Status:      ReplayStatusQueued,
		Differences: make([]Difference, 0),
		Summary: ReplaySummary{
			TotalSteps:     0,
			MatchingSteps:  0,
			DifferentSteps: 0,
			FailedSteps:    0,
		},
	}

	// Start replay execution
	go r.executeReplay(context.Background(), orgID, req, originalEvents, response)

	return response, nil
}

func (r *Replayer) executeReplay(ctx context.Context, orgID uuid.UUID, req *ReplayRequest, originalEvents []TraceEvent, response *ReplayResponse) {
	response.Status = ReplayStatusRunning

	// Group original events by step
	stepEvents := r.groupEventsByStep(originalEvents)

	// Execute replay based on mode
	switch req.Mode {
	case ReplayModeShadow:
		r.executeShadowReplay(ctx, orgID, req, stepEvents, response)
	case ReplayModeLive:
		r.executeLiveReplay(ctx, orgID, req, stepEvents, response)
	case ReplayModeDebug:
		r.executeDebugReplay(ctx, orgID, req, stepEvents, response)
	default:
		response.Status = ReplayStatusFailed
		return
	}

	// Calculate final summary
	r.calculateReplaySummary(response)
	response.Status = ReplayStatusCompleted
}

func (r *Replayer) executeShadowReplay(ctx context.Context, orgID uuid.UUID, req *ReplayRequest, stepEvents map[uuid.UUID][]TraceEvent, response *ReplayResponse) {
	// Shadow replay: execute steps without affecting state
	for stepID, events := range stepEvents {
		if r.shouldSkipStep(stepID.String(), req.StepFilter) {
			continue
		}

		response.Summary.TotalSteps++

		// Mock step execution
		replayResult := r.mockStepExecution(events)

		// Compare with original
		differences := r.compareStepResults(events, replayResult)
		response.Differences = append(response.Differences, differences...)

		if len(differences) == 0 {
			response.Summary.MatchingSteps++
		} else {
			response.Summary.DifferentSteps++
		}

		// Simulate processing time
		time.Sleep(100 * time.Millisecond)
	}
}

func (r *Replayer) executeLiveReplay(ctx context.Context, orgID uuid.UUID, req *ReplayRequest, stepEvents map[uuid.UUID][]TraceEvent, response *ReplayResponse) {
	// Live replay: full execution with state changes
	// This would integrate with the actual AOR system
	// For now, we'll simulate it
	r.executeShadowReplay(ctx, orgID, req, stepEvents, response)
}

func (r *Replayer) executeDebugReplay(ctx context.Context, orgID uuid.UUID, req *ReplayRequest, stepEvents map[uuid.UUID][]TraceEvent, response *ReplayResponse) {
	// Debug replay: step-by-step with detailed analysis
	// This would provide interactive debugging capabilities
	r.executeShadowReplay(ctx, orgID, req, stepEvents, response)
}

func (r *Replayer) groupEventsByStep(events []TraceEvent) map[uuid.UUID][]TraceEvent {
	stepEvents := make(map[uuid.UUID][]TraceEvent)

	for _, event := range events {
		if _, exists := stepEvents[event.StepID]; !exists {
			stepEvents[event.StepID] = make([]TraceEvent, 0)
		}
		stepEvents[event.StepID] = append(stepEvents[event.StepID], event)
	}

	return stepEvents
}

func (r *Replayer) shouldSkipStep(stepID string, filter []string) bool {
	if len(filter) == 0 {
		return false
	}

	for _, allowedStep := range filter {
		if stepID == allowedStep {
			return false
		}
	}

	return true
}

func (r *Replayer) mockStepExecution(originalEvents []TraceEvent) map[string]interface{} {
	// Mock step execution that simulates running the step again
	// In production, this would actually re-execute the step

	result := map[string]interface{}{
		"output":            "Mock replay output",
		"latency_ms":        150 + (time.Now().UnixNano()%100), // Add some variance
		"cost_cents":        100,
		"tokens_prompt":     80,
		"tokens_completion": 40,
		"provider":          "openai",
		"model":             "gpt-4",
	}

	// Add some randomness to simulate real differences
	if time.Now().UnixNano()%10 < 2 { // 20% chance of difference
		result["output"] = "Mock replay output with variation"
		result["cost_cents"] = 110
	}

	return result
}

func (r *Replayer) compareStepResults(originalEvents []TraceEvent, replayResult map[string]interface{}) []Difference {
	differences := make([]Difference, 0)

	// Find the completion event from original
	var originalCompletion *TraceEvent
	for _, event := range originalEvents {
		if event.EventType == EventTypeCompleted {
			originalCompletion = &event
			break
		}
	}

	if originalCompletion == nil {
		return differences
	}

	stepID := originalCompletion.StepID.String()

	// Compare output
	if originalOutput, exists := originalCompletion.Payload["output"]; exists {
		if replayOutput, exists := replayResult["output"]; exists {
			if fmt.Sprintf("%v", originalOutput) != fmt.Sprintf("%v", replayOutput) {
				differences = append(differences, Difference{
					StepID:       stepID,
					Field:        "output",
					Original:     originalOutput,
					Replay:       replayOutput,
					DiffType:     DiffTypeOutput,
					Significance: r.calculateSignificance(originalOutput, replayOutput),
				})
			}
		}
	}

	// Compare latency
	if replayLatency, exists := replayResult["latency_ms"]; exists {
		originalLatency := int64(originalCompletion.LatencyMs)
		replayLatencyInt := int64(replayLatency.(int))
		
		if abs64(originalLatency-replayLatencyInt) > 50 { // 50ms threshold
			differences = append(differences, Difference{
				StepID:       stepID,
				Field:        "latency_ms",
				Original:     originalLatency,
				Replay:       replayLatencyInt,
				DiffType:     DiffTypeLatency,
				Significance: "medium",
			})
		}
	}

	// Compare cost
	if replayCost, exists := replayResult["cost_cents"]; exists {
		originalCost := originalCompletion.CostCents
		replayCostInt := int64(replayCost.(int))
		
		if originalCost != replayCostInt {
			significance := "low"
			if abs64(originalCost-replayCostInt) > originalCost/10 { // >10% difference
				significance = "high"
			}

			differences = append(differences, Difference{
				StepID:       stepID,
				Field:        "cost_cents",
				Original:     originalCost,
				Replay:       replayCostInt,
				DiffType:     DiffTypeCost,
				Significance: significance,
			})
		}
	}

	// Compare tokens
	if replayTokensPrompt, exists := replayResult["tokens_prompt"]; exists {
		originalTokens := int64(originalCompletion.TokensPrompt)
		replayTokensInt := int64(replayTokensPrompt.(int))
		
		if originalTokens != replayTokensInt {
			differences = append(differences, Difference{
				StepID:       stepID,
				Field:        "tokens_prompt",
				Original:     originalTokens,
				Replay:       replayTokensInt,
				DiffType:     DiffTypeTokens,
				Significance: "low",
			})
		}
	}

	return differences
}

func (r *Replayer) calculateSignificance(original, replay interface{}) string {
	// Simple significance calculation
	originalStr := fmt.Sprintf("%v", original)
	replayStr := fmt.Sprintf("%v", replay)

	if originalStr == replayStr {
		return "none"
	}

	// Calculate string similarity (very basic)
	similarity := r.calculateStringSimilarity(originalStr, replayStr)
	
	if similarity > 0.9 {
		return "low"
	} else if similarity > 0.7 {
		return "medium"
	} else {
		return "high"
	}
}

func (r *Replayer) calculateStringSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// Simple character-based similarity
	matches := 0
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	minLen := len(s1)
	if len(s2) < minLen {
		minLen = len(s2)
	}

	for i := 0; i < minLen; i++ {
		if s1[i] == s2[i] {
			matches++
		}
	}

	return float64(matches) / float64(maxLen)
}

func (r *Replayer) calculateReplaySummary(response *ReplayResponse) {
	if response.Summary.TotalSteps > 0 {
		response.Summary.SimilarityScore = float64(response.Summary.MatchingSteps) / float64(response.Summary.TotalSteps)
	}

	// Calculate cost and latency differences
	totalCostDiff := int64(0)
	totalLatencyDiff := int64(0)
	costDiffCount := 0
	latencyDiffCount := 0

	for _, diff := range response.Differences {
		switch diff.DiffType {
		case DiffTypeCost:
			if originalCost, ok := diff.Original.(int64); ok {
				if replayCost, ok := diff.Replay.(int64); ok {
					totalCostDiff += replayCost - originalCost
					costDiffCount++
				}
			}
		case DiffTypeLatency:
			if originalLatency, ok := diff.Original.(int64); ok {
				if replayLatency, ok := diff.Replay.(int64); ok {
					totalLatencyDiff += replayLatency - originalLatency
					latencyDiffCount++
				}
			}
		}
	}

	if costDiffCount > 0 {
		response.Summary.CostDifference = totalCostDiff / int64(costDiffCount)
	}

	if latencyDiffCount > 0 {
		response.Summary.LatencyDifference = time.Duration(totalLatencyDiff/int64(latencyDiffCount)) * time.Millisecond
	}
}

func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}