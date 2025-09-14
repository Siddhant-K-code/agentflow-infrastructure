-- Drop indexes
DROP INDEX IF EXISTS idx_budget_org_period;
DROP INDEX IF EXISTS idx_context_bundle_hash;
DROP INDEX IF EXISTS idx_prompt_template_org_name;
DROP INDEX IF EXISTS idx_step_run_status;
DROP INDEX IF EXISTS idx_step_run_workflow_run_id;
DROP INDEX IF EXISTS idx_workflow_run_created_at;
DROP INDEX IF EXISTS idx_workflow_run_status;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS provider_config;
DROP TABLE IF EXISTS budget;
DROP TABLE IF EXISTS context_bundle;
DROP TABLE IF EXISTS prompt_deployment;
DROP TABLE IF EXISTS prompt_suite;
DROP TABLE IF EXISTS prompt_template;
DROP TABLE IF EXISTS step_run;
DROP TABLE IF EXISTS workflow_run;
DROP TABLE IF EXISTS workflow_spec;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS organizations;