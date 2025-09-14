-- ClickHouse schema for Agent Observability Stack (AOS)
CREATE DATABASE IF NOT EXISTS agentflow;

USE agentflow;

-- Main trace events table
CREATE TABLE IF NOT EXISTS trace_event (
    org_id UUID,
    run_id UUID,
    step_id UUID,
    ts DateTime64(3),
    event_type LowCardinality(String),
    payload JSON,
    cost_cents Int64 DEFAULT 0,
    tokens_prompt Int32 DEFAULT 0,
    tokens_completion Int32 DEFAULT 0,
    provider LowCardinality(String) DEFAULT '',
    model LowCardinality(String) DEFAULT '',
    quality_tier LowCardinality(String) DEFAULT '',
    latency_ms Int32 DEFAULT 0
) ENGINE = MergeTree()
PARTITION BY toDate(ts)
ORDER BY (org_id, run_id, ts)
TTL ts + INTERVAL 90 DAY;

-- Materialized view for cost aggregations
CREATE MATERIALIZED VIEW IF NOT EXISTS cost_by_hour_mv
ENGINE = SummingMergeTree()
PARTITION BY toDate(hour)
ORDER BY (org_id, provider, model, hour)
AS SELECT
    org_id,
    provider,
    model,
    toStartOfHour(ts) as hour,
    sum(cost_cents) as total_cost_cents,
    sum(tokens_prompt) as total_tokens_prompt,
    sum(tokens_completion) as total_tokens_completion,
    count() as event_count
FROM trace_event
WHERE event_type = 'model_io'
GROUP BY org_id, provider, model, hour;

-- Materialized view for quality metrics
CREATE MATERIALIZED VIEW IF NOT EXISTS quality_metrics_mv
ENGINE = SummingMergeTree()
PARTITION BY toDate(hour)
ORDER BY (org_id, quality_tier, hour)
AS SELECT
    org_id,
    quality_tier,
    toStartOfHour(ts) as hour,
    avg(latency_ms) as avg_latency_ms,
    quantile(0.95)(latency_ms) as p95_latency_ms,
    count() as request_count,
    countIf(event_type = 'completed' AND payload['status'] = 'failed') as failure_count
FROM trace_event
WHERE event_type IN ('completed', 'model_io')
GROUP BY org_id, quality_tier, hour;