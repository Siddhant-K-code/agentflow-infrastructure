package aos

import (
	"context"
	"fmt"
	"time"

	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/db"
)

type EventCollector struct {
	clickhouse *db.ClickHouseDB
	batchSize  int
	flushInterval time.Duration
	eventBuffer []TraceEvent
	bufferChan chan TraceEvent
	flushChan  chan struct{}
}

func NewEventCollector(ch *db.ClickHouseDB) *EventCollector {
	collector := &EventCollector{
		clickhouse:    ch,
		batchSize:     1000,
		flushInterval: 5 * time.Second,
		eventBuffer:   make([]TraceEvent, 0, 1000),
		bufferChan:    make(chan TraceEvent, 10000),
		flushChan:     make(chan struct{}, 1),
	}

	// Start background batch processor
	go collector.processBatches()

	return collector
}

// Ingest ingests a single trace event
func (ec *EventCollector) Ingest(ctx context.Context, event *TraceEvent) error {
	select {
	case ec.bufferChan <- *event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Buffer full, try direct insert
		return ec.insertEvent(ctx, *event)
	}
}

// IngestBatch ingests multiple trace events
func (ec *EventCollector) IngestBatch(ctx context.Context, events []TraceEvent) error {
	if len(events) == 0 {
		return nil
	}

	// For large batches, insert directly
	if len(events) > ec.batchSize {
		return ec.insertEvents(ctx, events)
	}

	// For smaller batches, add to buffer
	for _, event := range events {
		select {
		case ec.bufferChan <- event:
			continue
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Buffer full, flush and try again
			ec.triggerFlush()
			select {
			case ec.bufferChan <- event:
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return nil
}

func (ec *EventCollector) processBatches() {
	ticker := time.NewTicker(ec.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case event := <-ec.bufferChan:
			ec.eventBuffer = append(ec.eventBuffer, event)
			
			// Flush if buffer is full
			if len(ec.eventBuffer) >= ec.batchSize {
				ec.flush()
			}

		case <-ticker.C:
			// Periodic flush
			if len(ec.eventBuffer) > 0 {
				ec.flush()
			}

		case <-ec.flushChan:
			// Manual flush trigger
			if len(ec.eventBuffer) > 0 {
				ec.flush()
			}
		}
	}
}

func (ec *EventCollector) flush() {
	if len(ec.eventBuffer) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := ec.insertEvents(ctx, ec.eventBuffer); err != nil {
		// Log error but don't block - in production would have proper error handling
		fmt.Printf("Failed to flush events: %v\n", err)
	}

	// Clear buffer
	ec.eventBuffer = ec.eventBuffer[:0]
}

func (ec *EventCollector) triggerFlush() {
	select {
	case ec.flushChan <- struct{}{}:
	default:
		// Flush already triggered
	}
}

func (ec *EventCollector) insertEvent(ctx context.Context, event TraceEvent) error {
	return ec.insertEvents(ctx, []TraceEvent{event})
}

func (ec *EventCollector) insertEvents(ctx context.Context, events []TraceEvent) error {
	if len(events) == 0 {
		return nil
	}

	// Prepare batch insert
	batch, err := ec.clickhouse.PrepareBatch(ctx, `
		INSERT INTO trace_event (
			org_id, run_id, step_id, ts, event_type, payload,
			cost_cents, tokens_prompt, tokens_completion,
			provider, model, quality_tier, latency_ms
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	// Add events to batch
	for _, event := range events {
		payloadJSON, err := marshalPayload(event.Payload)
		if err != nil {
			continue // Skip events with invalid payload
		}

		err = batch.Append(
			event.OrgID,
			event.RunID,
			event.StepID,
			event.Timestamp,
			event.EventType,
			payloadJSON,
			event.CostCents,
			event.TokensPrompt,
			event.TokensCompletion,
			event.Provider,
			event.Model,
			event.QualityTier,
			event.LatencyMs,
		)
		if err != nil {
			continue // Skip events that fail to append
		}
	}

	// Execute batch
	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}

func marshalPayload(payload map[string]interface{}) (string, error) {
	if payload == nil {
		return "{}", nil
	}

	// Convert to JSON string for ClickHouse JSON column
	// In production, would use proper JSON marshaling
	result := "{"
	first := true
	for k, v := range payload {
		if !first {
			result += ","
		}
		result += fmt.Sprintf(`"%s":"%v"`, k, v)
		first = false
	}
	result += "}"

	return result, nil
}

// Shutdown gracefully shuts down the collector
func (ec *EventCollector) Shutdown(ctx context.Context) error {
	// Flush remaining events
	ec.triggerFlush()
	
	// Wait for flush to complete or timeout
	select {
	case <-time.After(10 * time.Second):
		// Force flush
		ec.flush()
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}