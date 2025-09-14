use anyhow::Result;
use serde_json::json;
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use tracing::{info, warn};
use crate::agent::AgentExecution;

pub struct RuntimeServer {
    port: u16,
    executions: Arc<RwLock<HashMap<String, AgentExecution>>>,
}

impl RuntimeServer {
    pub fn new(
        port: u16,
        executions: Arc<RwLock<HashMap<String, AgentExecution>>>,
    ) -> Self {
        Self { port, executions }
    }

    pub async fn start(&self) -> Result<()> {
        use warp::Filter;

        info!("🌐 Starting Runtime HTTP server on port {}", self.port);

        let executions = Arc::clone(&self.executions);

        // GET /health - Health check
        let health = warp::path("health")
            .and(warp::get())
            .map(|| {
                warp::reply::json(&json!({
                    "status": "healthy",
                    "service": "agentflow-runtime"
                }))
            });

        // GET /executions - List all executions
        let executions_clone = Arc::clone(&executions);
        let list_executions = warp::path("executions")
            .and(warp::get())
            .and_then(move || {
                let executions = Arc::clone(&executions_clone);
                async move {
                    let executions_read = executions.read().await;
                    let executions_list: Vec<_> = executions_read.values().cloned().collect();
                    
                    let response = json!({
                        "executions": executions_list,
                        "count": executions_list.len()
                    });
                    
                    Ok::<_, warp::Rejection>(warp::reply::json(&response))
                }
            });

        // GET /executions/{id} - Get specific execution
        let executions_clone = Arc::clone(&executions);
        let get_execution = warp::path!("executions" / String)
            .and(warp::get())
            .and_then(move |id: String| {
                let executions = Arc::clone(&executions_clone);
                async move {
                    let executions_read = executions.read().await;
                    
                    match executions_read.get(&id) {
                        Some(execution) => {
                            Ok(warp::reply::json(execution))
                        }
                        None => {
                            let response = json!({
                                "error": "Execution not found",
                                "id": id
                            });
                            Ok(warp::reply::json(&response))
                        }
                    }
                }
            });

        // GET /metrics - Runtime metrics
        let executions_clone = Arc::clone(&executions);
        let metrics = warp::path("metrics")
            .and(warp::get())
            .and_then(move || {
                let executions = Arc::clone(&executions_clone);
                async move {
                    let executions_read = executions.read().await;
                    
                    let total_executions = executions_read.len();
                    let running_executions = executions_read
                        .values()
                        .filter(|e| e.is_running())
                        .count();
                    let completed_executions = executions_read
                        .values()
                        .filter(|e| e.is_completed())
                        .count();

                    let response = json!({
                        "runtime_metrics": {
                            "total_executions": total_executions,
                            "running_executions": running_executions,
                            "completed_executions": completed_executions,
                            "memory_usage_mb": 256, // TODO: Get actual memory usage
                            "uptime_seconds": 3600  // TODO: Track actual uptime
                        }
                    });
                    
                    Ok::<_, warp::Rejection>(warp::reply::json(&response))
                }
            });

        // Combine all routes
        let routes = health
            .or(list_executions)
            .or(get_execution)
            .or(metrics)
            .with(warp::cors().allow_any_origin());

        // Start server
        warp::serve(routes)
            .run(([0, 0, 0, 0], self.port))
            .await;

        Ok(())
    }
}

// Note: This implementation uses warp, but since it's not in our Cargo.toml,
// let's create a simpler HTTP server using standard library for now
impl RuntimeServer {
    pub async fn start_simple(&self) -> Result<()> {
        use std::io::prelude::*;
        use std::net::{TcpListener, TcpStream};

        info!("🌐 Starting simple HTTP server on port {}", self.port);

        let listener = TcpListener::bind(format!("0.0.0.0:{}", self.port))?;
        
        for stream in listener.incoming() {
            match stream {
                Ok(stream) => {
                    let executions = Arc::clone(&self.executions);
                    tokio::spawn(async move {
                        if let Err(e) = Self::handle_connection(stream, executions).await {
                            warn!("Failed to handle connection: {}", e);
                        }
                    });
                }
                Err(e) => {
                    warn!("Connection failed: {}", e);
                }
            }
        }

        Ok(())
    }

    async fn handle_connection(
        mut stream: TcpStream,
        executions: Arc<RwLock<HashMap<String, AgentExecution>>>,
    ) -> Result<()> {
        let mut buffer = [0; 1024];
        stream.read(&mut buffer)?;

        let request = String::from_utf8_lossy(&buffer[..]);
        
        let response = if request.starts_with("GET /health") {
            "HTTP/1.1 200 OK\r\n\r\n{\"status\": \"healthy\"}"
        } else if request.starts_with("GET /executions") {
            let executions_read = executions.read().await;
            let count = executions_read.len();
            drop(executions_read);
            
            &format!("HTTP/1.1 200 OK\r\n\r\n{{\"count\": {}}}", count)
        } else {
            "HTTP/1.1 404 NOT FOUND\r\n\r\n{\"error\": \"Not found\"}"
        };

        stream.write_all(response.as_bytes())?;
        stream.flush()?;

        Ok(())
    }
}