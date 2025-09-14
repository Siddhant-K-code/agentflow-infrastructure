-- Agent Orchestration Runtime (AOR) schemas

-- workflows are versioned DAG specs
CREATE TABLE workflow_spec(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL,
  name TEXT NOT NULL,
  version INT NOT NULL,
  dag JSONB NOT NULL,                 -- nodes/edges, step types
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(org_id, name, version)
);

-- a run is an instance executing a workflow_spec
CREATE TABLE workflow_run(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workflow_spec_id UUID NOT NULL REFERENCES workflow_spec(id),
  status TEXT NOT NULL CHECK (status IN ('queued','running','succeeded','failed','canceled','partial-success')),
  started_at TIMESTAMPTZ DEFAULT NOW(),
  ended_at TIMESTAMPTZ,
  cost_cents BIGINT DEFAULT 0,
  metadata JSONB DEFAULT '{}',
  budget_cents BIGINT,
  tags JSONB DEFAULT '[]'
);

-- each node execution
CREATE TABLE step_run(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workflow_run_id UUID NOT NULL REFERENCES workflow_run(id),
  node_id TEXT NOT NULL,
  attempt INT NOT NULL DEFAULT 1,
  status TEXT NOT NULL CHECK (status IN ('queued','running','succeeded','failed','canceled','retrying')),
  worker_id TEXT,
  started_at TIMESTAMPTZ DEFAULT NOW(),
  ended_at TIMESTAMPTZ,
  input_ref TEXT,  -- S3 key
  output_ref TEXT, -- S3 key
  error TEXT,
  cost_cents BIGINT DEFAULT 0,
  tokens_prompt INT DEFAULT 0,
  tokens_completion INT DEFAULT 0
);

-- PromptOps Platform (POP) schemas
CREATE TABLE prompt_template(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL,
  name TEXT NOT NULL,
  version INT NOT NULL,
  template TEXT NOT NULL,          -- e.g., handlebars/mustache/roma
  schema JSONB,                   -- JSONSchema for inputs
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(org_id, name, version)
);

CREATE TABLE prompt_suite( -- eval suite
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL,
  name TEXT NOT NULL,
  cases JSONB NOT NULL,            -- [{input, expected, scoring}]
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE prompt_deployment(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL,
  prompt_name TEXT NOT NULL,
  stable_version INT NOT NULL,
  canary_version INT,
  canary_ratio FLOAT CHECK (canary_ratio BETWEEN 0 AND 1) DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Secure Context Layer (SCL) schemas
CREATE TABLE context_bundle(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL,
  hash TEXT NOT NULL,                      -- content hash
  schema_uri TEXT,
  trust_score FLOAT DEFAULT 0.5,
  redaction_map JSONB DEFAULT '{}',        -- offsets -> tokens
  provenance JSONB DEFAULT '{}',           -- sources, attestations
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Cost-Aware Scheduler (CAS) schemas
CREATE TABLE budget_config(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL,
  period TEXT NOT NULL CHECK (period IN ('daily', 'weekly', 'monthly')),
  limit_cents BIGINT NOT NULL,
  alert_threshold_ratio FLOAT DEFAULT 0.8,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(org_id, period)
);

CREATE TABLE provider_config(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL,
  provider_name TEXT NOT NULL,
  model_name TEXT NOT NULL,
  price_per_prompt_token_cents INT NOT NULL,
  price_per_completion_token_cents INT NOT NULL,
  quality_tier TEXT CHECK (quality_tier IN ('Gold', 'Silver', 'Bronze')) DEFAULT 'Silver',
  max_qps INT DEFAULT 10,
  enabled BOOLEAN DEFAULT true,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Agent Observability Stack (AOS) - using ClickHouse schema syntax
-- This would be created in ClickHouse, not Postgres
-- trace_event table structure documented for reference

-- Organizations and projects
CREATE TABLE organization(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE project(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL REFERENCES organization(id),
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(org_id, name)
);

-- Indexes for performance
CREATE INDEX idx_workflow_run_org_id ON workflow_run(org_id);
CREATE INDEX idx_workflow_run_status ON workflow_run(status);
CREATE INDEX idx_step_run_workflow_run_id ON step_run(workflow_run_id);
CREATE INDEX idx_step_run_status ON step_run(status);
CREATE INDEX idx_prompt_template_org_name ON prompt_template(org_id, name);
CREATE INDEX idx_context_bundle_org_hash ON context_bundle(org_id, hash);