package db

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/config"
)

type ClickHouseDB struct {
	driver.Conn
}

func NewClickHouseDB(cfg *config.ClickHouseConfig) (*ClickHouseDB, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.User,
			Password: cfg.Password,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	return &ClickHouseDB{Conn: conn}, nil
}

func (db *ClickHouseDB) InitSchema(ctx context.Context, schemaPath string) error {
	// This would read and execute the schema file
	// For now, we'll implement a basic version
	queries := []string{
		`CREATE DATABASE IF NOT EXISTS agentflow`,
		`USE agentflow`,
		`CREATE TABLE IF NOT EXISTS trace_event (
			org_id UUID,
			run_id UUID,
			step_id UUID,
			ts DateTime64(3),
			event_type LowCardinality(String),
			payload JSON,
			cost_cents Int64 DEFAULT 0,
			tokens_prompt Int32 DEFAULT 0,
			tokens_completion Int32 DEFAULT 0,
			provider LowCardinality(String) DEFAULT '',
			model LowCardinality(String) DEFAULT '',
			quality_tier LowCardinality(String) DEFAULT '',
			latency_ms Int32 DEFAULT 0
		) ENGINE = MergeTree()
		PARTITION BY toDate(ts)
		ORDER BY (org_id, run_id, ts)
		TTL ts + INTERVAL 90 DAY`,
	}

	for _, query := range queries {
		if err := db.Conn.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to execute query %q: %w", query, err)
		}
	}

	return nil
}
