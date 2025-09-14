package cas

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/agentflow/infrastructure/internal/db"
	"github.com/google/uuid"
)

type BudgetManager struct {
	postgres *db.PostgresDB
}

func NewBudgetManager(pg *db.PostgresDB) *BudgetManager {
	return &BudgetManager{
		postgres: pg,
	}
}

// CreateBudget creates a new budget
func (bm *BudgetManager) CreateBudget(ctx context.Context, budget *Budget) (*Budget, error) {
	query := `INSERT INTO budget (id, org_id, project_id, period_type, limit_cents, spent_cents, period_start, period_end, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			  RETURNING id`

	err := bm.postgres.QueryRowContext(ctx, query,
		budget.ID, budget.OrgID, budget.ProjectID, budget.PeriodType,
		budget.LimitCents, budget.SpentCents, budget.PeriodStart, budget.PeriodEnd, budget.CreatedAt,
	).Scan(&budget.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create budget: %w", err)
	}

	return budget, nil
}

// GetBudget retrieves a budget by ID
func (bm *BudgetManager) GetBudget(ctx context.Context, budgetID uuid.UUID) (*Budget, error) {
	query := `SELECT id, org_id, project_id, period_type, limit_cents, spent_cents, period_start, period_end, created_at
			  FROM budget WHERE id = $1`

	var budget Budget
	err := bm.postgres.QueryRowContext(ctx, query, budgetID).Scan(
		&budget.ID, &budget.OrgID, &budget.ProjectID, &budget.PeriodType,
		&budget.LimitCents, &budget.SpentCents, &budget.PeriodStart, &budget.PeriodEnd, &budget.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get budget: %w", err)
	}

	return &budget, nil
}

// GetCurrentBudget retrieves the current active budget for an organization
func (bm *BudgetManager) GetCurrentBudget(ctx context.Context, orgID uuid.UUID, projectID *uuid.UUID) (*Budget, error) {
	var query string
	var args []interface{}

	if projectID != nil {
		query = `SELECT id, org_id, project_id, period_type, limit_cents, spent_cents, period_start, period_end, created_at
				 FROM budget 
				 WHERE org_id = $1 AND project_id = $2 AND period_start <= NOW() AND period_end > NOW()
				 ORDER BY created_at DESC LIMIT 1`
		args = []interface{}{orgID, *projectID}
	} else {
		query = `SELECT id, org_id, project_id, period_type, limit_cents, spent_cents, period_start, period_end, created_at
				 FROM budget 
				 WHERE org_id = $1 AND project_id IS NULL AND period_start <= NOW() AND period_end > NOW()
				 ORDER BY created_at DESC LIMIT 1`
		args = []interface{}{orgID}
	}

	var budget Budget
	err := bm.postgres.QueryRowContext(ctx, query, args...).Scan(
		&budget.ID, &budget.OrgID, &budget.ProjectID, &budget.PeriodType,
		&budget.LimitCents, &budget.SpentCents, &budget.PeriodStart, &budget.PeriodEnd, &budget.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get current budget: %w", err)
	}

	return &budget, nil
}

// CheckBudget checks budget status and returns current state
func (bm *BudgetManager) CheckBudget(ctx context.Context, orgID uuid.UUID, requestedAmount int64) (*BudgetStatus, error) {
	budget, err := bm.GetCurrentBudget(ctx, orgID, nil)
	if err != nil {
		// If no budget exists, create a default one
		budget, err = bm.createDefaultBudget(ctx, orgID)
		if err != nil {
			return nil, fmt.Errorf("failed to create default budget: %w", err)
		}
	}

	status := &BudgetStatus{
		BudgetID:       budget.ID,
		LimitCents:     budget.LimitCents,
		SpentCents:     budget.SpentCents,
		RemainingCents: budget.LimitCents - budget.SpentCents,
		PeriodStart:    budget.PeriodStart,
		PeriodEnd:      budget.PeriodEnd,
	}

	// Calculate utilization percentage
	if budget.LimitCents > 0 {
		status.UtilizationPct = float64(budget.SpentCents) / float64(budget.LimitCents) * 100
	}

	// Determine status
	if status.RemainingCents <= 0 {
		status.Status = BudgetStatusExceeded
	} else if status.UtilizationPct >= 90 {
		status.Status = BudgetStatusCritical
	} else if status.UtilizationPct >= 75 {
		status.Status = BudgetStatusWarning
	} else {
		status.Status = BudgetStatusHealthy
	}

	// Check if requested amount would exceed budget
	if requestedAmount > 0 && status.RemainingCents < requestedAmount {
		status.Status = BudgetStatusExceeded
	}

	return status, nil
}

// GetStatus retrieves budget status for an organization
func (bm *BudgetManager) GetStatus(ctx context.Context, orgID uuid.UUID) (*BudgetStatus, error) {
	return bm.CheckBudget(ctx, orgID, 0)
}

// RecordSpending records spending against the budget
func (bm *BudgetManager) RecordSpending(ctx context.Context, orgID uuid.UUID, amountCents int64) error {
	budget, err := bm.GetCurrentBudget(ctx, orgID, nil)
	if err != nil {
		return fmt.Errorf("failed to get current budget: %w", err)
	}

	// Update spent amount
	query := `UPDATE budget SET spent_cents = spent_cents + $1 WHERE id = $2`
	_, err = bm.postgres.ExecContext(ctx, query, amountCents, budget.ID)
	if err != nil {
		return fmt.Errorf("failed to record spending: %w", err)
	}

	// Check if budget is exceeded and send alerts if needed
	newSpent := budget.SpentCents + amountCents
	if newSpent > budget.LimitCents {
		bm.sendBudgetAlert(ctx, budget, newSpent)
	}

	return nil
}

// UpdateBudget updates an existing budget
func (bm *BudgetManager) UpdateBudget(ctx context.Context, budgetID uuid.UUID, limitCents int64) error {
	query := `UPDATE budget SET limit_cents = $1 WHERE id = $2`
	_, err := bm.postgres.ExecContext(ctx, query, limitCents, budgetID)
	if err != nil {
		return fmt.Errorf("failed to update budget: %w", err)
	}

	return nil
}

// ListBudgets lists all budgets for an organization
func (bm *BudgetManager) ListBudgets(ctx context.Context, orgID uuid.UUID) ([]Budget, error) {
	query := `SELECT id, org_id, project_id, period_type, limit_cents, spent_cents, period_start, period_end, created_at
			  FROM budget 
			  WHERE org_id = $1 
			  ORDER BY created_at DESC`

	rows, err := bm.postgres.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list budgets: %w", err)
	}
	defer rows.Close()

	budgets := make([]Budget, 0)
	for rows.Next() {
		var budget Budget
		err := rows.Scan(
			&budget.ID, &budget.OrgID, &budget.ProjectID, &budget.PeriodType,
			&budget.LimitCents, &budget.SpentCents, &budget.PeriodStart, &budget.PeriodEnd, &budget.CreatedAt,
		)
		if err != nil {
			continue
		}
		budgets = append(budgets, budget)
	}

	return budgets, nil
}

// DeleteBudget deletes a budget
func (bm *BudgetManager) DeleteBudget(ctx context.Context, budgetID uuid.UUID) error {
	query := `DELETE FROM budget WHERE id = $1`
	_, err := bm.postgres.ExecContext(ctx, query, budgetID)
	if err != nil {
		return fmt.Errorf("failed to delete budget: %w", err)
	}

	return nil
}

// GetBudgetHistory retrieves spending history for a budget
func (bm *BudgetManager) GetBudgetHistory(ctx context.Context, budgetID uuid.UUID, days int) ([]BudgetHistoryEntry, error) {
	// This would query spending history from AOS trace events
	// For now, return mock data
	history := make([]BudgetHistoryEntry, 0)
	
	startDate := time.Now().AddDate(0, 0, -days)
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		entry := BudgetHistoryEntry{
			Date:       date,
			SpentCents: int64(1000 + i*100), // Mock spending
			EventCount: 50 + i*5,            // Mock event count
		}
		history = append(history, entry)
	}

	return history, nil
}

// Helper methods

func (bm *BudgetManager) createDefaultBudget(ctx context.Context, orgID uuid.UUID) (*Budget, error) {
	now := time.Now()
	budget := &Budget{
		ID:          uuid.New(),
		OrgID:       orgID,
		PeriodType:  PeriodMonthly,
		LimitCents:  1000000, // $10,000 default limit
		SpentCents:  0,
		PeriodStart: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()),
		PeriodEnd:   time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, 1, 0),
		CreatedAt:   now,
	}

	return bm.CreateBudget(ctx, budget)
}

func (bm *BudgetManager) sendBudgetAlert(ctx context.Context, budget *Budget, newSpent int64) {
	// Mock alert sending - in production would integrate with notification system
	utilizationPct := float64(newSpent) / float64(budget.LimitCents) * 100
	
	alert := BudgetAlert{
		BudgetID:       budget.ID,
		OrgID:          budget.OrgID,
		AlertType:      "budget_exceeded",
		Message:        fmt.Sprintf("Budget exceeded: %.1f%% utilized", utilizationPct),
		UtilizationPct: utilizationPct,
		SpentCents:     newSpent,
		LimitCents:     budget.LimitCents,
		Timestamp:      time.Now(),
	}

	// Would send to notification service
	alertJSON, _ := json.Marshal(alert)
	fmt.Printf("Budget Alert: %s\n", string(alertJSON))
}

// Supporting types

type BudgetHistoryEntry struct {
	Date       time.Time `json:"date"`
	SpentCents int64     `json:"spent_cents"`
	EventCount int       `json:"event_count"`
}

type BudgetAlert struct {
	BudgetID       uuid.UUID `json:"budget_id"`
	OrgID          uuid.UUID `json:"org_id"`
	AlertType      string    `json:"alert_type"`
	Message        string    `json:"message"`
	UtilizationPct float64   `json:"utilization_pct"`
	SpentCents     int64     `json:"spent_cents"`
	LimitCents     int64     `json:"limit_cents"`
	Timestamp      time.Time `json:"timestamp"`
}