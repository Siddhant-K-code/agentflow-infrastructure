package aor

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

type Monitor struct {
	cp       *ControlPlane
	mu       sync.RWMutex
	running  bool
	shutdown chan struct{}
}

func NewMonitor(cp *ControlPlane) *Monitor {
	return &Monitor{
		cp:       cp,
		shutdown: make(chan struct{}),
	}
}

func (m *Monitor) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return nil
	}

	m.running = true

	// Subscribe to results
	_, err := m.cp.js.Subscribe("agentflow.results", m.handleResult, nats.Durable("monitor-results"))
	if err != nil {
		return err
	}

	// Subscribe to heartbeats
	_, err = m.cp.js.Subscribe("agentflow.heartbeats", m.handleHeartbeat, nats.Durable("monitor-heartbeats"))
	if err != nil {
		return err
	}

	// Start monitoring loops
	go m.monitoringLoop(ctx)

	log.Println("Monitor started")
	return nil
}

func (m *Monitor) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	close(m.shutdown)
	m.running = false

	log.Println("Monitor shutdown")
	return nil
}

func (m *Monitor) handleResult(msg *nats.Msg) {
	var result TaskResult
	if err := json.Unmarshal(msg.Data, &result); err != nil {
		log.Printf("Failed to unmarshal result: %v", err)
		return
	}

	log.Printf("Received result for task %s: %s", result.TaskID, result.Status)

	// Results are already processed by the scheduler
	// This is mainly for monitoring and metrics
	_ = msg.Ack() // Ignore error for monitoring ack
}

func (m *Monitor) handleHeartbeat(msg *nats.Msg) {
	var heartbeat map[string]interface{}
	if err := json.Unmarshal(msg.Data, &heartbeat); err != nil {
		log.Printf("Failed to unmarshal heartbeat: %v", err)
		return
	}

	workerID, _ := heartbeat["worker_id"].(string)
	log.Printf("Received heartbeat from worker %s", workerID)

	// Store worker status in Redis for health monitoring
	key := "worker:" + workerID
	m.cp.redis.Set(context.Background(), key, string(msg.Data), 2*time.Minute)

	_ = msg.Ack() // Ignore error for heartbeat ack
}

func (m *Monitor) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.shutdown:
			return
		case <-ticker.C:
			m.checkStuckTasks(ctx)
			m.checkWorkerHealth(ctx)
		}
	}
}

func (m *Monitor) checkStuckTasks(ctx context.Context) {
	// Find tasks that have been running too long
	query := `SELECT id, workflow_run_id, node_id, started_at 
			  FROM step_run 
			  WHERE status = 'running' 
			  AND started_at < NOW() - INTERVAL '1 hour'`

	rows, err := m.cp.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Failed to query stuck tasks: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var stepID, runID string
		var nodeID string
		var startedAt time.Time

		if err := rows.Scan(&stepID, &runID, &nodeID, &startedAt); err != nil {
			continue
		}

		log.Printf("Found stuck task: step=%s, run=%s, node=%s, started=%v",
			stepID, runID, nodeID, startedAt)

		// Could implement automatic retry or cancellation here
	}
}

func (m *Monitor) checkWorkerHealth(ctx context.Context) {
	// Get all worker keys from Redis
	keys, err := m.cp.redis.Keys(ctx, "worker:*").Result()
	if err != nil {
		log.Printf("Failed to get worker keys: %v", err)
		return
	}

	activeWorkers := len(keys)
	log.Printf("Active workers: %d", activeWorkers)

	// Could implement alerts for low worker count, etc.
}
