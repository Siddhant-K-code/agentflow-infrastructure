-- ClickHouse schema for Agent Observability Stack (AOS)
-- This is separate from the main Postgres schema

CREATE TABLE trace_event (
  org_id UUID,
  run_id UUID,
  step_id UUID,
  ts DateTime64(3),
  event_type LowCardinality(String), -- started/completed/retry/log/tool_call/model_io
  payload JSON,
  cost_cents Int64,
  tokens_prompt Int32,
  tokens_completion Int32,
  provider String,
  model String,
  quality_tier LowCardinality(String),
  latency_ms Int32
) ENGINE=MergeTree 
PARTITION BY toDate(ts) 
ORDER BY (org_id, run_id, ts);