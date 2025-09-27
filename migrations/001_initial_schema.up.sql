-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Organizations table
CREATE TABLE IF NOT EXISTS IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Projects table
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(org_id, name)
);

-- AOR: Workflow specifications
CREATE TABLE IF NOT EXISTS workflow_spec (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    version INTEGER NOT NULL,
    dag JSONB NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(org_id, name, version)
);

-- AOR: Workflow runs
CREATE TABLE IF NOT EXISTS workflow_run (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_spec_id UUID NOT NULL REFERENCES workflow_spec(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('queued','running','succeeded','failed','canceled','partial-success')),
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    cost_cents BIGINT DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- AOR: Step runs
CREATE TABLE IF NOT EXISTS step_run (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_run_id UUID NOT NULL REFERENCES workflow_run(id) ON DELETE CASCADE,
    node_id TEXT NOT NULL,
    attempt INTEGER NOT NULL DEFAULT 1,
    status TEXT NOT NULL CHECK (status IN ('queued','running','succeeded','failed','canceled')),
    worker_id TEXT,
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    input_ref TEXT,
    output_ref TEXT,
    error TEXT,
    cost_cents BIGINT DEFAULT 0,
    tokens_prompt INTEGER DEFAULT 0,
    tokens_completion INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- POP: Prompt templates
CREATE TABLE IF NOT EXISTS prompt_template (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    version INTEGER NOT NULL,
    template TEXT NOT NULL,
    schema JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(org_id, name, version)
);

-- POP: Evaluation suites
CREATE TABLE IF NOT EXISTS prompt_suite (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    cases JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- POP: Prompt deployments
CREATE TABLE IF NOT EXISTS prompt_deployment (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    prompt_name TEXT NOT NULL,
    stable_version INTEGER NOT NULL,
    canary_version INTEGER,
    canary_ratio FLOAT CHECK (canary_ratio BETWEEN 0 AND 1) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- SCL: Context bundles
CREATE TABLE IF NOT EXISTS context_bundle (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    hash TEXT NOT NULL,
    schema_uri TEXT,
    trust_score FLOAT DEFAULT 0.5,
    redaction_map JSONB DEFAULT '{}',
    provenance JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(org_id, hash)
);

-- CAS: Budgets
CREATE TABLE IF NOT EXISTS budget (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    period_type TEXT NOT NULL CHECK (period_type IN ('daily','weekly','monthly')),
    limit_cents BIGINT NOT NULL,
    spent_cents BIGINT DEFAULT 0,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- CAS: Provider configurations
CREATE TABLE IF NOT EXISTS provider_config (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    provider_name TEXT NOT NULL,
    model_name TEXT NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    cost_per_token_prompt DECIMAL(10,8),
    cost_per_token_completion DECIMAL(10,8),
    qps_limit INTEGER DEFAULT 100,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(org_id, provider_name, model_name)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_workflow_run_status ON workflow_run(status);
CREATE INDEX IF NOT EXISTS idx_workflow_run_created_at ON workflow_run(created_at);
CREATE INDEX IF NOT EXISTS idx_step_run_workflow_run_id ON step_run(workflow_run_id);
CREATE INDEX IF NOT EXISTS idx_step_run_status ON step_run(status);
CREATE INDEX IF NOT EXISTS idx_prompt_template_org_name ON prompt_template(org_id, name);
CREATE INDEX IF NOT EXISTS idx_context_bundle_hash ON context_bundle(hash);
CREATE INDEX IF NOT EXISTS idx_budget_org_period ON budget(org_id, period_start, period_end);