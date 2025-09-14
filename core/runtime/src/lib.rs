use anyhow::{Context, Result};
use async_nats::Client as NatsClient;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use tracing::{debug, error, info, warn};
use uuid::Uuid;
use wasmtime::*;

pub mod wasm;
pub mod agent;
pub mod server;

pub use wasm::WasmSandbox;
pub use agent::{AgentExecution, AgentRequest, AgentResponse};
pub use server::RuntimeServer;

#[derive(Debug, Clone)]
pub struct RuntimeConfig {
    pub port: u16,
    pub nats_url: String,
    pub orchestrator_url: String,
}

pub struct AgentflowRuntime {
    config: RuntimeConfig,
    nats_client: NatsClient,
    executions: Arc<RwLock<HashMap<String, AgentExecution>>>,
    wasm_sandbox: Arc<WasmSandbox>,
}

impl AgentflowRuntime {
    pub async fn new(config: RuntimeConfig) -> Result<Self> {
        info!("🔧 Initializing AgentFlow Runtime...");

        // Connect to NATS
        let nats_client = async_nats::connect(&config.nats_url)
            .await
            .context("Failed to connect to NATS")?;
        
        info!("✅ Connected to NATS at {}", config.nats_url);

        // Initialize WASM sandbox
        let wasm_sandbox = Arc::new(WasmSandbox::new()?);
        info!("✅ WASM sandbox initialized");

        let runtime = Self {
            config,
            nats_client,
            executions: Arc::new(RwLock::new(HashMap::new())),
            wasm_sandbox,
        };

        // Subscribe to agent execution requests
        runtime.subscribe_to_agent_requests().await?;

        Ok(runtime)
    }

    async fn subscribe_to_agent_requests(&self) -> Result<()> {
        let mut subscriber = self.nats_client
            .subscribe("agentflow.agent.execute")
            .await
            .context("Failed to subscribe to agent execution requests")?;

        let executions = Arc::clone(&self.executions);
        let wasm_sandbox = Arc::clone(&self.wasm_sandbox);
        
        tokio::spawn(async move {
            while let Some(message) = subscriber.next().await {
                if let Err(e) = Self::handle_agent_request(
                    message,
                    Arc::clone(&executions),
                    Arc::clone(&wasm_sandbox),
                ).await {
                    error!("Failed to handle agent request: {}", e);
                }
            }
        });

        info!("📡 Subscribed to agent execution requests");
        Ok(())
    }

    async fn handle_agent_request(
        message: async_nats::Message,
        executions: Arc<RwLock<HashMap<String, AgentExecution>>>,
        wasm_sandbox: Arc<WasmSandbox>,
    ) -> Result<()> {
        let request: AgentRequest = serde_json::from_slice(&message.payload)
            .context("Failed to deserialize agent request")?;

        debug!("🤖 Received agent execution request: {}", request.agent_id);

        // Create agent execution
        let execution = AgentExecution::new(request.clone());
        let execution_id = execution.id.clone();
        
        // Store execution
        executions.write().await.insert(execution_id.clone(), execution.clone());

        // Execute agent in WASM sandbox
        let result = wasm_sandbox.execute_agent(request).await;
        
        // Update execution with result
        {
            let mut executions_write = executions.write().await;
            if let Some(exec) = executions_write.get_mut(&execution_id) {
                match result {
                    Ok(response) => exec.complete_success(response),
                    Err(e) => exec.complete_error(e.to_string()),
                }
            }
        }

        // Send response back via NATS if reply subject is provided
        if let Some(reply) = message.reply {
            let execution_read = executions.read().await;
            if let Some(exec) = execution_read.get(&execution_id) {
                let response = AgentResponse {
                    execution_id: exec.id.clone(),
                    agent_id: exec.request.agent_id.clone(),
                    status: exec.status.clone(),
                    result: exec.result.clone(),
                    error: exec.error.clone(),
                    started_at: exec.started_at,
                    completed_at: exec.completed_at,
                };

                let response_bytes = serde_json::to_vec(&response)
                    .context("Failed to serialize agent response")?;
                
                if let Err(e) = message.respond(response_bytes.into()).await {
                    warn!("Failed to send agent response: {}", e);
                }
            }
        }

        Ok(())
    }

    pub async fn start_server(&self) -> Result<()> {
        let server = RuntimeServer::new(
            self.config.port,
            Arc::clone(&self.executions),
        );
        
        server.start().await
    }

    pub async fn shutdown(&self) -> Result<()> {
        info!("🛑 Shutting down AgentFlow Runtime...");
        
        // Cancel all running executions
        let executions = self.executions.read().await;
        for execution in executions.values() {
            execution.cancel().await;
        }
        
        Ok(())
    }

    pub async fn get_execution(&self, id: &str) -> Option<AgentExecution> {
        self.executions.read().await.get(id).cloned()
    }

    pub async fn list_executions(&self) -> Vec<AgentExecution> {
        self.executions.read().await.values().cloned().collect()
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ExecutionStatus {
    Pending,
    Running,
    Completed,
    Failed,
    Cancelled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutionMetrics {
    pub memory_usage: u64,
    pub cpu_time: u64,
    pub llm_calls: u32,
    pub execution_time_ms: u64,
}