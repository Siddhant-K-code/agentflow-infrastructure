package aor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/config"
	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/db"
	"github.com/google/uuid"
	nats "github.com/nats-io/nats.go"
	redis "github.com/redis/go-redis/v9"
)

type ControlPlane struct {
	cfg   *config.Config
	db    *db.PostgresDB
	redis *redis.Client
	nats  *nats.Conn
	js    nats.JetStreamContext

	scheduler *Scheduler
	monitor   *Monitor

	mu       sync.RWMutex
	running  bool
	shutdown chan struct{}
}

func NewControlPlane(cfg *config.Config) (*ControlPlane, error) {
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

	// Initialize NATS (optional for demo)
	var nc *nats.Conn
	var js nats.JetStreamContext
	nc, err = nats.Connect(cfg.NATS.URL)
	if err != nil {
		log.Printf("Warning: failed to connect to NATS: %v (continuing without NATS for demo)", err)
		nc = nil
		js = nil
	} else {
		js, err = nc.JetStream()
		if err != nil {
			log.Printf("Warning: failed to get JetStream context: %v (continuing without NATS for demo)", err)
			nc = nil
			js = nil
		}
	}

	cp := &ControlPlane{
		cfg:      cfg,
		db:       pgDB,
		redis:    redisClient,
		nats:     nc,
		js:       js,
		shutdown: make(chan struct{}),
	}

	// Initialize scheduler and monitor
	cp.scheduler = NewScheduler(pgDB, redisClient, nc, js)
	cp.monitor = NewMonitor(cp)

	return cp, nil
}

func (cp *ControlPlane) Start(ctx context.Context) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if cp.running {
		return fmt.Errorf("control plane already running")
	}

	// Run database migrations
	if err := cp.db.RunMigrations("./migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize NATS streams
	if err := cp.initStreams(); err != nil {
		log.Printf("Warning: failed to initialize streams: %v", err)
		// Continue without streams for demo purposes
	}

	// Scheduler doesn't need explicit start in this implementation
	log.Printf("Scheduler initialized")

	// Start monitor
	if err := cp.monitor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start monitor: %w", err)
	}

	cp.running = true
	log.Println("Control plane started")

	return nil
}

func (cp *ControlPlane) Shutdown(ctx context.Context) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if !cp.running {
		return nil
	}

	close(cp.shutdown)

	// Scheduler doesn't need explicit shutdown in this implementation
	log.Printf("Scheduler shutdown")
	if cp.monitor != nil {
		_ = cp.monitor.Shutdown(ctx) // Ignore shutdown errors
	}

	// Close connections
	if cp.nats != nil {
		cp.nats.Close()
	}
	if cp.redis != nil {
		_ = cp.redis.Close() // Ignore close errors
	}
	if cp.db != nil {
		_ = cp.db.Close() // Ignore close errors
	}

	cp.running = false
	log.Println("Control plane shutdown complete")

	return nil
}

func (cp *ControlPlane) SubmitWorkflow(ctx context.Context, req *RunRequest) (*WorkflowRun, error) {
	// Get workflow spec
	spec, err := cp.getWorkflowSpec(ctx, req.WorkflowName, req.WorkflowVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow spec: %w", err)
	}

	// Create workflow run
	run := &WorkflowRun{
		ID:             uuid.New(),
		WorkflowSpecID: spec.ID,
		Status:         RunStatusQueued,
		Metadata: map[string]interface{}{
			"inputs":       req.Inputs,
			"tags":         req.Tags,
			"budget_cents": req.BudgetCents,
		},
		CreatedAt: time.Now(),
	}

	// Save to database
	if err := cp.saveWorkflowRun(ctx, run); err != nil {
		return nil, fmt.Errorf("failed to save workflow run: %w", err)
	}

	// Submit to scheduler
	if err := cp.scheduler.ScheduleWorkflow(ctx, run); err != nil {
		return nil, fmt.Errorf("failed to schedule workflow: %w", err)
	}

	return run, nil
}

func (cp *ControlPlane) GetWorkflowRun(ctx context.Context, runID uuid.UUID) (*WorkflowRun, error) {
	query := `SELECT id, workflow_spec_id, status, started_at, ended_at, cost_cents, metadata, created_at
			  FROM workflow_run WHERE id = $1`

	var run WorkflowRun
	var metadataJSON []byte

	err := cp.db.QueryRowContext(ctx, query, runID).Scan(
		&run.ID, &run.WorkflowSpecID, &run.Status, &run.StartedAt, &run.EndedAt,
		&run.CostCents, &metadataJSON, &run.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow run: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &run.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &run, nil
}

func (cp *ControlPlane) CancelWorkflowRun(ctx context.Context, runID uuid.UUID) error {
	// Update workflow run status to canceled
	query := `UPDATE workflow_run SET status = 'canceled', ended_at = NOW() WHERE id = $1`
	_, err := cp.db.ExecContext(ctx, query, runID)
	if err != nil {
		return fmt.Errorf("failed to cancel workflow run: %w", err)
	}
	return nil
}

func (cp *ControlPlane) ListWorkflowRuns(ctx context.Context, limit, offset int) ([]WorkflowRun, error) {
	query := `SELECT id, workflow_spec_id, status, started_at, ended_at, cost_cents, metadata, created_at
			  FROM workflow_run
			  ORDER BY created_at DESC
			  LIMIT $1 OFFSET $2`

	rows, err := cp.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflow runs: %w", err)
	}
	defer rows.Close()

	var runs []WorkflowRun
	for rows.Next() {
		var run WorkflowRun
		var metadataJSON []byte

		err := rows.Scan(
			&run.ID, &run.WorkflowSpecID, &run.Status, &run.StartedAt, &run.EndedAt,
			&run.CostCents, &metadataJSON, &run.CreatedAt,
		)
		if err != nil {
			continue
		}

		if err := json.Unmarshal(metadataJSON, &run.Metadata); err != nil {
			run.Metadata = make(map[string]interface{})
		}

		runs = append(runs, run)
	}

	return runs, nil
}

func (cp *ControlPlane) initStreams() error {
	// Initialize NATS JetStream streams for workflow processing
	streams := []string{
		"WORKFLOW_QUEUE",
		"WORKFLOW_EVENTS",
		"STEP_QUEUE",
		"STEP_EVENTS",
	}

	for _, streamName := range streams {
		_, err := cp.js.AddStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{streamName + ".*"},
			Storage:  nats.FileStorage,
		})
		if err != nil && !strings.Contains(err.Error(), "stream name already in use") {
			return fmt.Errorf("failed to create stream %s: %w", streamName, err)
		}
	}

	return nil
}

func (cp *ControlPlane) getWorkflowSpec(ctx context.Context, name string, version int) (*WorkflowSpec, error) {
	// For demo purposes, return a mock workflow spec
	return &WorkflowSpec{
		ID:      uuid.New(),
		OrgID:   uuid.New(),
		Name:    name,
		Version: version,
		DAG: DAG{
			Steps: []Step{
				{
					ID:   "step1",
					Type: "llm",
					Name: "Process Input",
					Config: map[string]interface{}{
						"prompt": "Process the input: {{input}}",
					},
				},
			},
		},
		Created: time.Now(),
		Updated: time.Now(),
	}, nil
}

func (cp *ControlPlane) saveWorkflowRun(ctx context.Context, run *WorkflowRun) error {
	query := `INSERT INTO workflow_run (id, workflow_spec_id, status, created_at, cost_cents, metadata)
			  VALUES ($1, $2, $3, $4, $5, $6)`

	metadataJSON, err := json.Marshal(run.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = cp.db.ExecContext(ctx, query,
		run.ID, run.WorkflowSpecID, run.Status, run.CreatedAt, run.CostCents, metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to insert workflow run: %w", err)
	}

	log.Printf("Saved workflow run: %s", run.ID)
	return nil
}
