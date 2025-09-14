use serde::{Deserialize, Serialize};
use std::time::{SystemTime, UNIX_EPOCH};
use uuid::Uuid;
use crate::{ExecutionStatus, wasm::AgentResult};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentRequest {
    pub agent_id: String,
    pub workflow_id: String,
    pub image: String,
    pub input: serde_json::Value,
    pub config: AgentConfig,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentConfig {
    pub llm_provider: String,
    pub llm_model: String,
    pub timeout_seconds: Option<u64>,
    pub max_retries: Option<u32>,
    pub environment: Option<std::collections::HashMap<String, String>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentResponse {
    pub execution_id: String,
    pub agent_id: String,
    pub status: ExecutionStatus,
    pub result: Option<AgentResult>,
    pub error: Option<String>,
    pub started_at: u64,
    pub completed_at: Option<u64>,
}

#[derive(Debug, Clone)]
pub struct AgentExecution {
    pub id: String,
    pub request: AgentRequest,
    pub status: ExecutionStatus,
    pub result: Option<AgentResult>,
    pub error: Option<String>,
    pub started_at: u64,
    pub completed_at: Option<u64>,
}

impl AgentExecution {
    pub fn new(request: AgentRequest) -> Self {
        let now = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_secs();

        Self {
            id: Uuid::new_v4().to_string(),
            request,
            status: ExecutionStatus::Pending,
            result: None,
            error: None,
            started_at: now,
            completed_at: None,
        }
    }

    pub fn start(&mut self) {
        self.status = ExecutionStatus::Running;
        self.started_at = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_secs();
    }

    pub fn complete_success(&mut self, result: AgentResult) {
        self.status = ExecutionStatus::Completed;
        self.result = Some(result);
        self.completed_at = Some(
            SystemTime::now()
                .duration_since(UNIX_EPOCH)
                .unwrap()
                .as_secs(),
        );
    }

    pub fn complete_error(&mut self, error: String) {
        self.status = ExecutionStatus::Failed;
        self.error = Some(error);
        self.completed_at = Some(
            SystemTime::now()
                .duration_since(UNIX_EPOCH)
                .unwrap()
                .as_secs(),
        );
    }

    pub async fn cancel(&self) {
        // TODO: Implement cancellation logic
        tracing::info!("🚫 Cancelling agent execution: {}", self.id);
    }

    pub fn duration_ms(&self) -> Option<u64> {
        if let Some(completed) = self.completed_at {
            Some((completed - self.started_at) * 1000)
        } else {
            let now = SystemTime::now()
                .duration_since(UNIX_EPOCH)
                .unwrap()
                .as_secs();
            Some((now - self.started_at) * 1000)
        }
    }

    pub fn is_running(&self) -> bool {
        matches!(self.status, ExecutionStatus::Running)
    }

    pub fn is_completed(&self) -> bool {
        matches!(
            self.status,
            ExecutionStatus::Completed | ExecutionStatus::Failed | ExecutionStatus::Cancelled
        )
    }
}