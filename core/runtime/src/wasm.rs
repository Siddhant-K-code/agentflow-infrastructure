use anyhow::{Context, Result};
use serde::{Deserialize, Serialize};
use std::time::{Duration, SystemTime, UNIX_EPOCH};
use tracing::{debug, error, info, warn};
use uuid::Uuid;
use wasmtime::*;

pub struct WasmSandbox {
    engine: Engine,
}

impl WasmSandbox {
    pub fn new() -> Result<Self> {
        let mut config = Config::new();
        
        // Enable WebAssembly features for sandboxing
        config.wasm_backtrace_details(WasmBacktraceDetails::Enable);
        config.wasm_multi_memory(true);
        config.wasm_memory64(false);
        
        // Set resource limits for security
        config.epoch_interruption(true);
        config.max_wasm_stack(1024 * 1024); // 1MB stack limit
        
        let engine = Engine::new(&config)
            .context("Failed to create WASM engine")?;
        
        Ok(Self { engine })
    }

    pub async fn execute_agent(&self, request: crate::agent::AgentRequest) -> Result<AgentResult> {
        info!("🔒 Executing agent {} in WASM sandbox", request.agent_id);
        
        // Create a new store for this execution
        let mut store = Store::new(&self.engine, WasmContext::new());
        
        // Set resource limits
        store.limiter(|ctx| &mut ctx.limiter);
        store.epoch_deadline_trap();
        
        // Load the agent WASM module
        let wasm_bytes = self.load_agent_wasm(&request.image).await?;
        let module = Module::new(&self.engine, &wasm_bytes)
            .context("Failed to compile WASM module")?;
        
        // Create instance
        let instance = Instance::new(&mut store, &module, &[])
            .context("Failed to instantiate WASM module")?;
        
        // Get the main execution function
        let execute_func = instance
            .get_typed_func::<(i32, i32), i32>(&mut store, "execute")
            .context("Failed to get execute function from WASM module")?;
        
        // Prepare input data
        let input_json = serde_json::to_string(&request.input)
            .context("Failed to serialize agent input")?;
        
        let input_ptr = self.allocate_string(&mut store, &instance, &input_json)?;
        let output_ptr = 0; // Will be set by the WASM function
        
        debug!("🚀 Calling WASM execute function");
        
        // Execute with timeout
        let start_time = SystemTime::now();
        
        // Set epoch deadline for timeout
        store.set_epoch_deadline(1);
        
        let result = execute_func.call(&mut store, (input_ptr, output_ptr));
        
        let execution_time = start_time.elapsed().unwrap_or(Duration::ZERO);
        
        match result {
            Ok(result_ptr) => {
                debug!("✅ WASM execution completed successfully");
                
                // Read result from WASM memory
                let output_json = self.read_string(&mut store, &instance, result_ptr)?;
                let agent_output: AgentOutput = serde_json::from_str(&output_json)
                    .context("Failed to deserialize agent output")?;
                
                Ok(AgentResult {
                    success: true,
                    output: Some(agent_output),
                    error: None,
                    metrics: ExecutionMetrics {
                        execution_time_ms: execution_time.as_millis() as u64,
                        memory_usage: self.get_memory_usage(&store, &instance)?,
                        cpu_time: execution_time.as_millis() as u64,
                        llm_calls: agent_output.llm_calls.unwrap_or(0),
                    },
                })
            }
            Err(e) => {
                error!("❌ WASM execution failed: {}", e);
                
                Ok(AgentResult {
                    success: false,
                    output: None,
                    error: Some(e.to_string()),
                    metrics: ExecutionMetrics {
                        execution_time_ms: execution_time.as_millis() as u64,
                        memory_usage: 0,
                        cpu_time: execution_time.as_millis() as u64,
                        llm_calls: 0,
                    },
                })
            }
        }
    }

    async fn load_agent_wasm(&self, image: &str) -> Result<Vec<u8>> {
        // TODO: Load WASM from container registry or local storage
        // For now, return a minimal WASM module
        info!("📦 Loading WASM module for image: {}", image);
        
        // This would typically download from a registry
        // For demo purposes, we'll generate a minimal valid WASM module
        Ok(self.generate_demo_wasm_module())
    }

    fn generate_demo_wasm_module(&self) -> Vec<u8> {
        // Minimal WASM module with execute function
        // This is a placeholder - in production, agents would be compiled to WASM
        vec![
            0x00, 0x61, 0x73, 0x6d, // WASM magic number
            0x01, 0x00, 0x00, 0x00, // Version
            // Type section
            0x01, 0x07, 0x01, 0x60, 0x02, 0x7f, 0x7f, 0x01, 0x7f,
            // Function section  
            0x03, 0x02, 0x01, 0x00,
            // Export section
            0x07, 0x0b, 0x01, 0x07, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x00, 0x00,
            // Code section
            0x0a, 0x09, 0x01, 0x07, 0x00, 0x20, 0x00, 0x20, 0x01, 0x6a, 0x0b,
        ]
    }

    fn allocate_string(&self, store: &mut Store<WasmContext>, instance: &Instance, s: &str) -> Result<i32> {
        // TODO: Implement proper memory allocation in WASM
        // For now, return a mock pointer
        debug!("📝 Allocating string in WASM memory: {} bytes", s.len());
        Ok(1000) // Mock pointer
    }

    fn read_string(&self, store: &mut Store<WasmContext>, instance: &Instance, ptr: i32) -> Result<String> {
        // TODO: Implement proper memory reading from WASM
        // For now, return mock output
        debug!("📖 Reading string from WASM memory at pointer: {}", ptr);
        
        Ok(r#"{"result": "Agent execution completed", "llm_calls": 1}"#.to_string())
    }

    fn get_memory_usage(&self, store: &Store<WasmContext>, instance: &Instance) -> Result<u64> {
        // TODO: Get actual memory usage from WASM instance
        Ok(1024 * 1024) // Mock 1MB usage
    }
}

#[derive(Debug, Clone)]
struct WasmContext {
    limiter: StoreLimitsBuilder,
}

impl WasmContext {
    fn new() -> Self {
        let limiter = StoreLimitsBuilder::new()
            .memory_size(10 * 1024 * 1024) // 10MB memory limit
            .table_elements(1000)
            .instances(1)
            .tables(1)
            .memories(1);
        
        Self { limiter }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentResult {
    pub success: bool,
    pub output: Option<AgentOutput>,
    pub error: Option<String>,
    pub metrics: crate::ExecutionMetrics,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentOutput {
    pub result: String,
    pub llm_calls: Option<u32>,
    pub artifacts: Option<Vec<String>>,
}