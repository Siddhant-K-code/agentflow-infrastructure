package aor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/config"
	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/db"
	nats "github.com/nats-io/nats.go"
	redis "github.com/redis/go-redis/v9"
)

type Worker struct {
	id    string
	cfg   *config.Config
	db    *db.PostgresDB
	redis *redis.Client
	nats  *nats.Conn
	js    nats.JetStreamContext

	executors map[ExecutorType]Executor

	mu       sync.RWMutex
	running  bool
	shutdown chan struct{}
}

// Remove duplicate Executor interface - it's already defined in types.go

func NewWorker(cfg *config.Config) (*Worker, error) {
	// Initialize database
	pgDB, err := db.NewPostgresDB(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres: %w", err)
	}

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Initialize NATS
	nc, err := nats.Connect(cfg.NATS.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to get JetStream context: %w", err)
	}

	worker := &Worker{
		id:        uuid.New().String(),
		cfg:       cfg,
		db:        pgDB,
		redis:     redisClient,
		nats:      nc,
		js:        js,
		shutdown:  make(chan struct{}),
		executors: make(map[ExecutorType]Executor),
	}

	// Initialize executors
	worker.executors[ExecutorTypeLLM] = NewLLMExecutor(worker)
	worker.executors[ExecutorTypeHTTP] = NewHTTPExecutor(worker)
	worker.executors[ExecutorTypeScript] = NewScriptExecutor(worker)

	return worker, nil
}

func (w *Worker) Start(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return fmt.Errorf("worker already running")
	}

	// Start worker execution loop
	go w.run(ctx)

	w.running = true
	log.Printf("Worker %s started", w.id)

	return nil
}

func (w *Worker) Shutdown(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return nil
	}

	close(w.shutdown)

	// Close connections
	if w.nats != nil {
		w.nats.Close()
	}
	if w.redis != nil {
		_ = w.redis.Close()
	}
	if w.db != nil {
		_ = w.db.Close()
	}

	w.running = false
	log.Printf("Worker %s shutdown complete", w.id)

	return nil
}

func (w *Worker) run(ctx context.Context) {
	log.Printf("Worker %s: Starting execution loop", w.id)

	// For demo purposes, just log that worker is running
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %s: Context cancelled", w.id)
			return
		case <-w.shutdown:
			log.Printf("Worker %s: Shutdown signal received", w.id)
			return
		case <-ticker.C:
			log.Printf("Worker %s: Still running...", w.id)
		}
	}
}

func (w *Worker) handleTask(msg *nats.Msg) {
	var task Task
	if err := json.Unmarshal(msg.Data, &task); err != nil {
		log.Printf("Failed to unmarshal task: %v", err)
		_ = msg.Nak() // Ignore nak error
		return
	}

	deadline := time.Now().Add(30 * time.Minute)
	if task.DeadlineAt != nil {
		deadline = *task.DeadlineAt
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Until(deadline))
	defer cancel()

	// Update step status to running
	if err := w.updateStepStatus(ctx, task.ID, StepStatusRunning, w.id); err != nil {
		log.Printf("Failed to update step status to running: %v", err)
		_ = msg.Nak() // Ignore nak error
		return
	}

	// Execute task
	result, err := w.executeTask(ctx, &task)
	if err != nil {
		log.Printf("Failed to execute task %s: %v", task.ID, err)
		result = &TaskResult{
			TaskID: task.ID,
			Status: TaskStatusFailed,
			Error:  err.Error(),
		}
	}

	// Update step with result
	if err := w.updateStepWithResult(ctx, result); err != nil {
		log.Printf("Failed to update step with result: %v", err)
		_ = msg.Nak() // Ignore nak error
		return
	}

	// Publish result
	if err := w.publishResult(ctx, result); err != nil {
		log.Printf("Failed to publish result: %v", err)
		_ = msg.Nak() // Ignore nak error
		return
	}

	_ = msg.Ack() // Ignore ack error
}

func (w *Worker) executeTask(ctx context.Context, task *Task) (*TaskResult, error) {
	executor, exists := w.executors[ExecutorType(task.Node.Type)]
	if !exists {
		return nil, fmt.Errorf("no executor for node type %s", task.Node.Type)
	}

	// Add retry logic
	maxRetries := 3 // Default retry count

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		result, err := executor.Execute(ctx, task)
		if err == nil {
			return result, nil
		}

		lastErr = err
		if attempt < maxRetries {
			// Exponential backoff
			backoff := time.Duration(attempt*attempt) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				continue
			}
		}
	}

	return nil, fmt.Errorf("task failed after %d attempts: %w", maxRetries, lastErr)
}

func (w *Worker) updateStepStatus(ctx context.Context, stepID uuid.UUID, status StepStatus, workerID string) error {
	var startedAt *time.Time
	if status == StepStatusRunning {
		now := time.Now()
		startedAt = &now
	}

	query := `UPDATE step_run SET status = $1, worker_id = $2, started_at = $3 WHERE id = $4`
	_, err := w.db.ExecContext(ctx, query, status, workerID, startedAt, stepID)
	return err
}

func (w *Worker) updateStepWithResult(ctx context.Context, result *TaskResult) error {
	now := time.Now()

	query := `UPDATE step_run SET
			  status = $1, ended_at = $2, error = $3, cost_cents = $4,
			  tokens_prompt = $5, tokens_completion = $6
			  WHERE id = $7`

	_, err := w.db.ExecContext(ctx, query,
		result.Status, now, result.Error, result.CostCents,
		result.TokensPrompt, result.TokensCompletion, result.TaskID,
	)
	return err
}

func (w *Worker) publishResult(ctx context.Context, result *TaskResult) error {
	resultData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	_, err = w.js.Publish("agentflow.results", resultData)
	return err
}

func (w *Worker) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.shutdown:
			return
		case <-ticker.C:
			w.sendHeartbeat(ctx)
		}
	}
}

func (w *Worker) sendHeartbeat(ctx context.Context) {
	heartbeat := map[string]interface{}{
		"worker_id": w.id,
		"timestamp": time.Now(),
		"status":    "healthy",
	}

	data, _ := json.Marshal(heartbeat)
	_, _ = w.js.Publish("agentflow.heartbeats", data) // Ignore publish error for heartbeat
}
