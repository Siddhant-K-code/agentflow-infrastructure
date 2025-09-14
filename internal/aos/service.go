package aos

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/agentflow/infrastructure/internal/config"
	"github.com/agentflow/infrastructure/internal/db"
	"github.com/google/uuid"
)

type Service struct {
	cfg         *config.Config
	clickhouse  *db.ClickHouseDB
	postgres    *db.PostgresDB
	collector   *EventCollector
	analyzer    *TraceAnalyzer
	replayer    *Replayer
}

func NewService(cfg *config.Config, ch *db.ClickHouseDB, pg *db.PostgresDB) *Service {
	service := &Service{
		cfg:        cfg,
		clickhouse: ch,
		postgres:   pg,
	}

	service.collector = NewEventCollector(ch)
	service.analyzer = NewTraceAnalyzer(ch)
	service.replayer = NewReplayer(pg, ch)

	return service
}

// IngestEvent ingests a trace event into the observability system
func (s *Service) IngestEvent(ctx context.Context, event *TraceEvent) error {
	return s.collector.Ingest(ctx, event)
}

// IngestEvents ingests multiple trace events in batch
func (s *Service) IngestEvents(ctx context.Context, events []TraceEvent) error {
	return s.collector.IngestBatch(ctx, events)
}

// QueryTrace retrieves trace events based on query parameters
func (s *Service) QueryTrace(ctx context.Context, query *TraceQuery) (*TraceResponse, error) {
	events, totalCount, err := s.analyzer.QueryEvents(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}

	summary, err := s.analyzer.GenerateSummary(ctx, events)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}

	return &TraceResponse{
		Events:     events,
		TotalCount: totalCount,
		Summary:    *summary,
	}, nil
}

// GetRunTrace retrieves the complete trace for a specific workflow run
func (s *Service) GetRunTrace(ctx context.Context, orgID, runID uuid.UUID) (*TraceResponse, error) {
	query := &TraceQuery{
		OrgID: orgID,
		RunID: &runID,
		Limit: 10000, // Large limit for complete trace
	}

	return s.QueryTrace(ctx, query)
}

// ReplayRun replays a workflow run for debugging and comparison
func (s *Service) ReplayRun(ctx context.Context, orgID uuid.UUID, req *ReplayRequest) (*ReplayResponse, error) {
	// Validate the original run exists
	originalTrace, err := s.GetRunTrace(ctx, orgID, req.RunID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original trace: %w", err)
	}

	if len(originalTrace.Events) == 0 {
		return nil, fmt.Errorf("no trace events found for run %s", req.RunID)
	}

	// Execute replay
	replayResult, err := s.replayer.Replay(ctx, orgID, req, originalTrace.Events)
	if err != nil {
		return nil, fmt.Errorf("replay failed: %w", err)
	}

	return replayResult, nil
}

// AnalyzeCosts performs cost analysis and returns breakdown and trends
func (s *Service) AnalyzeCosts(ctx context.Context, req *CostAnalysisRequest) (*CostAnalysisResponse, error) {
	// Build cost query
	query := s.buildCostQuery(req)

	// Execute query
	breakdown, err := s.analyzer.GetCostBreakdown(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost breakdown: %w", err)
	}

	trends, err := s.analyzer.GetCostTrends(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost trends: %w", err)
	}

	projections, err := s.analyzer.GetCostProjections(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost projections: %w", err)
	}

	savings, err := s.analyzer.GetCostSavings(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost savings: %w", err)
	}

	// Calculate total cost
	totalCost := int64(0)
	for _, item := range breakdown {
		totalCost += item.Cost
	}

	return &CostAnalysisResponse{
		TotalCost:   totalCost,
		Breakdown:   breakdown,
		Trends:      trends,
		Projections: projections,
		Savings:     *savings,
	}, nil
}

// AnalyzeQualityDrift analyzes quality drift for a prompt over time
func (s *Service) AnalyzeQualityDrift(ctx context.Context, req *QualityDriftRequest) (*QualityDriftResponse, error) {
	// Get quality metrics over time
	metrics, err := s.analyzer.GetQualityMetrics(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get quality metrics: %w", err)
	}

	// Calculate drift score
	driftScore := s.calculateDriftScore(metrics)

	// Determine trend
	trend := s.determineTrend(metrics)

	// Generate alerts
	alerts := s.generateQualityAlerts(metrics)

	// Generate recommendations
	recommendations := s.generateRecommendations(metrics, alerts)

	return &QualityDriftResponse{
		DriftScore:      driftScore,
		Trend:           trend,
		Metrics:         metrics,
		Alerts:          alerts,
		Recommendations: recommendations,
	}, nil
}

// QueryMetrics retrieves aggregated metrics
func (s *Service) QueryMetrics(ctx context.Context, query *MetricsQuery) (*MetricsResponse, error) {
	series, err := s.analyzer.QueryMetrics(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}

	return &MetricsResponse{
		Series: series,
	}, nil
}

// GetDashboardData retrieves data for observability dashboards
func (s *Service) GetDashboardData(ctx context.Context, orgID uuid.UUID, timeRange string) (map[string]interface{}, error) {
	endTime := time.Now()
	var startTime time.Time

	switch timeRange {
	case "1h":
		startTime = endTime.Add(-1 * time.Hour)
	case "24h":
		startTime = endTime.Add(-24 * time.Hour)
	case "7d":
		startTime = endTime.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = endTime.Add(-30 * 24 * time.Hour)
	default:
		startTime = endTime.Add(-24 * time.Hour)
	}

	// Get overview metrics
	overviewQuery := &MetricsQuery{
		OrgID:     orgID,
		StartTime: startTime,
		EndTime:   endTime,
		Metrics:   []string{"request_count", "success_rate", "avg_latency", "total_cost"},
		Interval:  "1h",
	}

	overview, err := s.QueryMetrics(ctx, overviewQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get overview metrics: %w", err)
	}

	// Get cost breakdown
	costQuery := &CostAnalysisRequest{
		OrgID:     orgID,
		StartTime: startTime,
		EndTime:   endTime,
		GroupBy:   []string{"provider", "model"},
	}

	costAnalysis, err := s.AnalyzeCosts(ctx, costQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost analysis: %w", err)
	}

	// Get recent errors
	errorQuery := &TraceQuery{
		OrgID:     orgID,
		EventType: stringPtr(EventTypeError),
		StartTime: &startTime,
		EndTime:   &endTime,
		Limit:     50,
	}

	errors, err := s.QueryTrace(ctx, errorQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent errors: %w", err)
	}

	return map[string]interface{}{
		"overview":      overview,
		"cost_analysis": costAnalysis,
		"recent_errors": errors,
		"time_range":    timeRange,
		"generated_at":  time.Now(),
	}, nil
}

// Helper methods

func (s *Service) buildCostQuery(req *CostAnalysisRequest) string {
	// Build ClickHouse query for cost analysis
	// This is a simplified version - production would be more sophisticated
	query := fmt.Sprintf(`
		SELECT 
			%s,
			sum(cost_cents) as total_cost,
			count() as event_count
		FROM trace_event 
		WHERE org_id = '%s' 
		AND ts >= '%s' 
		AND ts <= '%s'
		AND event_type = 'model_io'
		GROUP BY %s
		ORDER BY total_cost DESC
	`, 
		joinGroupBy(req.GroupBy),
		req.OrgID.String(),
		req.StartTime.Format("2006-01-02 15:04:05"),
		req.EndTime.Format("2006-01-02 15:04:05"),
		joinGroupBy(req.GroupBy),
	)

	return query
}

func (s *Service) calculateDriftScore(metrics []QualityMetric) float64 {
	if len(metrics) == 0 {
		return 0.0
	}

	totalDrift := 0.0
	for _, metric := range metrics {
		drift := abs(metric.Current - metric.Baseline)
		totalDrift += drift
	}

	return totalDrift / float64(len(metrics))
}

func (s *Service) determineTrend(metrics []QualityMetric) string {
	if len(metrics) == 0 {
		return "stable"
	}

	improving := 0
	degrading := 0

	for _, metric := range metrics {
		if metric.Change > 0.05 {
			improving++
		} else if metric.Change < -0.05 {
			degrading++
		}
	}

	if improving > degrading {
		return "improving"
	} else if degrading > improving {
		return "degrading"
	}

	return "stable"
}

func (s *Service) generateQualityAlerts(metrics []QualityMetric) []QualityAlert {
	alerts := make([]QualityAlert, 0)

	for _, metric := range metrics {
		if abs(metric.Change) > 0.2 { // 20% change threshold
			severity := "medium"
			if abs(metric.Change) > 0.5 {
				severity = "high"
			}

			alerts = append(alerts, QualityAlert{
				Severity:  severity,
				Message:   fmt.Sprintf("Significant change in %s: %.2f%% change", metric.Name, metric.Change*100),
				Metric:    metric.Name,
				Threshold: 0.2,
				Actual:    abs(metric.Change),
				Timestamp: time.Now(),
			})
		}
	}

	return alerts
}

func (s *Service) generateRecommendations(metrics []QualityMetric, alerts []QualityAlert) []string {
	recommendations := make([]string, 0)

	if len(alerts) > 0 {
		recommendations = append(recommendations, "Review recent prompt changes for quality impact")
		recommendations = append(recommendations, "Consider rolling back to previous stable version")
		recommendations = append(recommendations, "Run additional evaluation tests")
	}

	for _, metric := range metrics {
		if metric.Change < -0.3 {
			recommendations = append(recommendations, fmt.Sprintf("Investigate degradation in %s metric", metric.Name))
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Quality metrics are stable")
	}

	return recommendations
}

// Utility functions

func stringPtr(s string) *string {
	return &s
}

func joinGroupBy(fields []string) string {
	if len(fields) == 0 {
		return "1" // GROUP BY 1 for aggregation without grouping
	}
	result := ""
	for i, field := range fields {
		if i > 0 {
			result += ", "
		}
		result += field
	}
	return result
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}