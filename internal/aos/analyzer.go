package aos

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/db"
)

type TraceAnalyzer struct {
	clickhouse *db.ClickHouseDB
}

func NewTraceAnalyzer(ch *db.ClickHouseDB) *TraceAnalyzer {
	return &TraceAnalyzer{
		clickhouse: ch,
	}
}

// QueryEvents retrieves trace events based on query parameters
func (ta *TraceAnalyzer) QueryEvents(ctx context.Context, query *TraceQuery) ([]TraceEvent, int64, error) {
	// Build WHERE clause
	whereClause, args := ta.buildWhereClause(query)

	// Build main query
	sqlQuery := fmt.Sprintf(`
		SELECT 
			org_id, run_id, step_id, ts, event_type, payload,
			cost_cents, tokens_prompt, tokens_completion,
			provider, model, quality_tier, latency_ms
		FROM trace_event 
		WHERE %s
		ORDER BY ts DESC
		LIMIT %d OFFSET %d
	`, whereClause, query.Limit, query.Offset)

	// Execute query
	rows, err := ta.clickhouse.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Parse results
	events := make([]TraceEvent, 0)
	for rows.Next() {
		var event TraceEvent
		var payloadStr string

		err := rows.Scan(
			&event.OrgID, &event.RunID, &event.StepID, &event.Timestamp,
			&event.EventType, &payloadStr, &event.CostCents,
			&event.TokensPrompt, &event.TokensCompletion,
			&event.Provider, &event.Model, &event.QualityTier, &event.LatencyMs,
		)
		if err != nil {
			continue // Skip malformed rows
		}

		// Parse payload JSON
		event.Payload = parsePayload(payloadStr)
		events = append(events, event)
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT count() FROM trace_event WHERE %s
	`, whereClause)

	var totalCount int64
	err = ta.clickhouse.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		totalCount = int64(len(events)) // Fallback
	}

	return events, totalCount, nil
}

// GenerateSummary generates a summary for a set of trace events
func (ta *TraceAnalyzer) GenerateSummary(ctx context.Context, events []TraceEvent) (*TraceSummary, error) {
	if len(events) == 0 {
		return &TraceSummary{}, nil
	}

	summary := &TraceSummary{
		TotalEvents:       int64(len(events)),
		ProviderBreakdown: make(map[string]int64),
		ModelBreakdown:    make(map[string]int64),
	}

	totalLatency := int64(0)
	latencyCount := int64(0)

	for _, event := range events {
		// Accumulate costs and tokens
		summary.TotalCost += event.CostCents
		summary.TotalTokens += int64(event.TokensPrompt + event.TokensCompletion)

		// Track latency
		if event.LatencyMs > 0 {
			totalLatency += int64(event.LatencyMs)
			latencyCount++
		}

		// Count errors
		if event.EventType == EventTypeError {
			summary.ErrorCount++
		}

		// Provider breakdown
		if event.Provider != "" {
			summary.ProviderBreakdown[event.Provider]++
		}

		// Model breakdown
		if event.Model != "" {
			summary.ModelBreakdown[event.Model]++
		}
	}

	// Calculate averages
	if latencyCount > 0 {
		summary.AverageLatency = time.Duration(totalLatency/latencyCount) * time.Millisecond
	}

	// Calculate success rate
	if summary.TotalEvents > 0 {
		successCount := summary.TotalEvents - summary.ErrorCount
		summary.SuccessRate = float64(successCount) / float64(summary.TotalEvents)
	}

	return summary, nil
}

// GetCostBreakdown retrieves cost breakdown data
func (ta *TraceAnalyzer) GetCostBreakdown(ctx context.Context, query string) ([]CostBreakdown, error) {
	rows, err := ta.clickhouse.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute cost query: %w", err)
	}
	defer rows.Close()

	breakdown := make([]CostBreakdown, 0)
	totalCost := int64(0)

	// First pass: collect data and calculate total
	tempResults := make([]struct {
		dimensions map[string]string
		cost       int64
		count      int64
	}, 0)

	for rows.Next() {
		var provider, model string
		var cost, count int64

		err := rows.Scan(&provider, &model, &cost, &count)
		if err != nil {
			continue
		}

		dimensions := map[string]string{
			"provider": provider,
			"model":    model,
		}

		tempResults = append(tempResults, struct {
			dimensions map[string]string
			cost       int64
			count      int64
		}{dimensions, cost, count})

		totalCost += cost
	}

	// Second pass: calculate percentages
	for _, result := range tempResults {
		percentage := 0.0
		if totalCost > 0 {
			percentage = float64(result.cost) / float64(totalCost) * 100
		}

		breakdown = append(breakdown, CostBreakdown{
			Dimensions: result.dimensions,
			Cost:       result.cost,
			Percentage: percentage,
			Count:      result.count,
		})
	}

	return breakdown, nil
}

// GetCostTrends retrieves cost trend data over time
func (ta *TraceAnalyzer) GetCostTrends(ctx context.Context, query string) ([]CostTrend, error) {
	// Modify query to group by time intervals
	trendQuery := `
		SELECT 
			toStartOfHour(ts) as hour,
			sum(cost_cents) as cost,
			count() as count
		FROM trace_event 
		WHERE event_type = 'model_io'
		AND ts >= now() - INTERVAL 24 HOUR
		GROUP BY hour
		ORDER BY hour
	`

	rows, err := ta.clickhouse.Query(ctx, trendQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to execute trend query: %w", err)
	}
	defer rows.Close()

	trends := make([]CostTrend, 0)
	for rows.Next() {
		var timestamp time.Time
		var cost, count int64

		err := rows.Scan(&timestamp, &cost, &count)
		if err != nil {
			continue
		}

		trends = append(trends, CostTrend{
			Timestamp: timestamp,
			Cost:      cost,
			Count:     count,
		})
	}

	return trends, nil
}

// GetCostProjections generates cost projections
func (ta *TraceAnalyzer) GetCostProjections(ctx context.Context, query string) ([]CostProjection, error) {
	// Mock implementation - in production would use time series forecasting
	projections := []CostProjection{
		{
			Period:     "daily",
			Projected:  50000, // $500 in cents
			Confidence: 0.85,
			Trend:      "stable",
		},
		{
			Period:     "weekly",
			Projected:  350000, // $3500 in cents
			Confidence: 0.75,
			Trend:      "increasing",
		},
		{
			Period:     "monthly",
			Projected:  1500000, // $15000 in cents
			Confidence: 0.65,
			Trend:      "increasing",
		},
	}

	return projections, nil
}

// GetCostSavings calculates potential cost savings
func (ta *TraceAnalyzer) GetCostSavings(ctx context.Context, query string) (*CostSavings, error) {
	// Mock implementation - in production would analyze actual usage patterns
	savings := &CostSavings{
		CachingEnabled:      5000, // $50 saved through caching
		ProviderRouting:     8000, // $80 saved through smart routing
		QualityOptimization: 3000, // $30 saved through quality optimization
	}

	savings.TotalSavings = savings.CachingEnabled + savings.ProviderRouting + savings.QualityOptimization

	return savings, nil
}

// GetQualityMetrics retrieves quality metrics for drift analysis
func (ta *TraceAnalyzer) GetQualityMetrics(ctx context.Context, req *QualityDriftRequest) ([]QualityMetric, error) {
	// Mock implementation - in production would analyze actual quality data
	metrics := []QualityMetric{
		{
			Name:      "accuracy",
			Current:   0.85,
			Baseline:  0.88,
			Change:    -0.03,
			Timestamp: time.Now(),
		},
		{
			Name:      "relevance",
			Current:   0.92,
			Baseline:  0.90,
			Change:    0.02,
			Timestamp: time.Now(),
		},
		{
			Name:      "coherence",
			Current:   0.78,
			Baseline:  0.82,
			Change:    -0.04,
			Timestamp: time.Now(),
		},
	}

	return metrics, nil
}

// QueryMetrics retrieves aggregated metrics
func (ta *TraceAnalyzer) QueryMetrics(ctx context.Context, query *MetricsQuery) ([]MetricSeries, error) {
	series := make([]MetricSeries, 0)

	for _, metric := range query.Metrics {
		metricSeries, err := ta.getMetricSeries(ctx, metric, query)
		if err != nil {
			continue // Skip failed metrics
		}
		series = append(series, *metricSeries)
	}

	return series, nil
}

func (ta *TraceAnalyzer) getMetricSeries(ctx context.Context, metric string, query *MetricsQuery) (*MetricSeries, error) {
	var sqlQuery string
	var labels map[string]string

	switch metric {
	case "request_count":
		sqlQuery = ta.buildRequestCountQuery(query)
		labels = map[string]string{"metric": "request_count"}
	case "success_rate":
		sqlQuery = ta.buildSuccessRateQuery(query)
		labels = map[string]string{"metric": "success_rate"}
	case "avg_latency":
		sqlQuery = ta.buildLatencyQuery(query)
		labels = map[string]string{"metric": "avg_latency"}
	case "total_cost":
		sqlQuery = ta.buildCostQuery(query)
		labels = map[string]string{"metric": "total_cost"}
	default:
		return nil, fmt.Errorf("unknown metric: %s", metric)
	}

	rows, err := ta.clickhouse.Query(ctx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to execute metric query: %w", err)
	}
	defer rows.Close()

	dataPoints := make([]MetricDataPoint, 0)
	for rows.Next() {
		var timestamp time.Time
		var value float64

		err := rows.Scan(&timestamp, &value)
		if err != nil {
			continue
		}

		dataPoints = append(dataPoints, MetricDataPoint{
			Timestamp: timestamp,
			Value:     value,
		})
	}

	return &MetricSeries{
		Name:       metric,
		Labels:     labels,
		DataPoints: dataPoints,
	}, nil
}

// Helper methods

func (ta *TraceAnalyzer) buildWhereClause(query *TraceQuery) (string, []interface{}) {
	conditions := []string{"org_id = ?"}
	args := []interface{}{query.OrgID}

	if query.RunID != nil {
		conditions = append(conditions, "run_id = ?")
		args = append(args, *query.RunID)
	}

	if query.StepID != nil {
		conditions = append(conditions, "step_id = ?")
		args = append(args, *query.StepID)
	}

	if query.StartTime != nil {
		conditions = append(conditions, "ts >= ?")
		args = append(args, *query.StartTime)
	}

	if query.EndTime != nil {
		conditions = append(conditions, "ts <= ?")
		args = append(args, *query.EndTime)
	}

	if query.EventType != nil {
		conditions = append(conditions, "event_type = ?")
		args = append(args, *query.EventType)
	}

	if query.Provider != nil {
		conditions = append(conditions, "provider = ?")
		args = append(args, *query.Provider)
	}

	if query.Model != nil {
		conditions = append(conditions, "model = ?")
		args = append(args, *query.Model)
	}

	return strings.Join(conditions, " AND "), args
}

func (ta *TraceAnalyzer) buildRequestCountQuery(query *MetricsQuery) string {
	return fmt.Sprintf(`
		SELECT 
			toStartOfInterval(ts, INTERVAL 1 HOUR) as timestamp,
			count() as value
		FROM trace_event 
		WHERE org_id = '%s'
		AND ts >= '%s' AND ts <= '%s'
		GROUP BY timestamp
		ORDER BY timestamp
	`, query.OrgID, query.StartTime.Format("2006-01-02 15:04:05"), query.EndTime.Format("2006-01-02 15:04:05"))
}

func (ta *TraceAnalyzer) buildSuccessRateQuery(query *MetricsQuery) string {
	return fmt.Sprintf(`
		SELECT 
			toStartOfInterval(ts, INTERVAL 1 HOUR) as timestamp,
			(countIf(event_type != 'error') * 100.0 / count()) as value
		FROM trace_event 
		WHERE org_id = '%s'
		AND ts >= '%s' AND ts <= '%s'
		GROUP BY timestamp
		ORDER BY timestamp
	`, query.OrgID, query.StartTime.Format("2006-01-02 15:04:05"), query.EndTime.Format("2006-01-02 15:04:05"))
}

func (ta *TraceAnalyzer) buildLatencyQuery(query *MetricsQuery) string {
	return fmt.Sprintf(`
		SELECT 
			toStartOfInterval(ts, INTERVAL 1 HOUR) as timestamp,
			avg(latency_ms) as value
		FROM trace_event 
		WHERE org_id = '%s'
		AND ts >= '%s' AND ts <= '%s'
		AND latency_ms > 0
		GROUP BY timestamp
		ORDER BY timestamp
	`, query.OrgID, query.StartTime.Format("2006-01-02 15:04:05"), query.EndTime.Format("2006-01-02 15:04:05"))
}

func (ta *TraceAnalyzer) buildCostQuery(query *MetricsQuery) string {
	return fmt.Sprintf(`
		SELECT 
			toStartOfInterval(ts, INTERVAL 1 HOUR) as timestamp,
			sum(cost_cents) as value
		FROM trace_event 
		WHERE org_id = '%s'
		AND ts >= '%s' AND ts <= '%s'
		AND event_type = 'model_io'
		GROUP BY timestamp
		ORDER BY timestamp
	`, query.OrgID, query.StartTime.Format("2006-01-02 15:04:05"), query.EndTime.Format("2006-01-02 15:04:05"))
}

func parsePayload(payloadStr string) map[string]interface{} {
	// Simple JSON parsing - in production would use proper JSON unmarshaling
	payload := make(map[string]interface{})
	if payloadStr == "" || payloadStr == "{}" {
		return payload
	}

	// Mock parsing for demo
	payload["raw"] = payloadStr
	return payload
}
