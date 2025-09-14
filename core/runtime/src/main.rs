use clap::{Arg, Command};
use std::sync::Arc;
use tokio::signal;
use tracing::{info, error};
use agentflow_runtime::{RuntimeConfig, AgentflowRuntime};

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // Initialize tracing
    tracing_subscriber::init();

    let matches = Command::new("agentflow-runtime")
        .version("0.1.0")
        .about("AgentFlow WASM Runtime for secure agent execution")
        .arg(
            Arg::new("port")
                .short('p')
                .long("port")
                .value_name("PORT")
                .help("Port to run the runtime server on")
                .default_value("8081")
        )
        .arg(
            Arg::new("nats-url")
                .short('n')
                .long("nats-url")
                .value_name("URL")
                .help("NATS server URL")
                .default_value("nats://localhost:4222")
        )
        .arg(
            Arg::new("orchestrator-url")
                .short('o')
                .long("orchestrator-url")
                .value_name("URL")
                .help("Orchestrator service URL")
                .default_value("http://localhost:8080")
        )
        .get_matches();

    let config = RuntimeConfig {
        port: matches.get_one::<String>("port").unwrap().parse()?,
        nats_url: matches.get_one::<String>("nats-url").unwrap().clone(),
        orchestrator_url: matches.get_one::<String>("orchestrator-url").unwrap().clone(),
    };

    info!("🦀 Starting AgentFlow Runtime on port {}", config.port);

    let runtime = Arc::new(AgentflowRuntime::new(config).await?);
    
    // Start the runtime server
    let runtime_clone = Arc::clone(&runtime);
    let server_handle = tokio::spawn(async move {
        if let Err(e) = runtime_clone.start_server().await {
            error!("Runtime server error: {}", e);
        }
    });

    // Wait for shutdown signal
    match signal::ctrl_c().await {
        Ok(()) => {
            info!("🛑 Shutdown signal received, stopping runtime...");
        }
        Err(err) => {
            error!("Unable to listen for shutdown signal: {}", err);
        }
    }

    // Stop the server
    server_handle.abort();
    runtime.shutdown().await?;
    
    info!("✅ Runtime shutdown complete");
    Ok(())
}