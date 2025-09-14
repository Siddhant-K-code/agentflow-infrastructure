package common

import (
	"time"

	"github.com/google/uuid"
)

// Organization represents a tenant organization
type Organization struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string    `json:"name" gorm:"unique;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"default:now()"`
}

// Project represents a project within an organization
type Project struct {
	ID     uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrgID  uuid.UUID `json:"org_id" gorm:"type:uuid;not null"`
	Name   string    `json:"name" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"default:now()"`
}

// Status represents the execution status of workflows and steps
type Status string

const (
	StatusQueued         Status = "queued"
	StatusRunning        Status = "running"
	StatusSucceeded      Status = "succeeded"
	StatusFailed         Status = "failed"
	StatusCanceled       Status = "canceled"
	StatusPartialSuccess Status = "partial-success"
	StatusRetrying       Status = "retrying"
)

// QualityTier represents the quality tier for model selection
type QualityTier string

const (
	QualityGold   QualityTier = "Gold"
	QualitySilver QualityTier = "Silver"
	QualityBronze QualityTier = "Bronze"
)

// BudgetPeriod represents budget enforcement periods
type BudgetPeriod string

const (
	BudgetDaily   BudgetPeriod = "daily"
	BudgetWeekly  BudgetPeriod = "weekly"
	BudgetMonthly BudgetPeriod = "monthly"
)

// EventType represents trace event types for observability
type EventType string

const (
	EventStarted      EventType = "started"
	EventCompleted    EventType = "completed"
	EventRetry        EventType = "retry"
	EventLog          EventType = "log"
	EventToolCall     EventType = "tool_call"
	EventModelIO      EventType = "model_io"
	EventError        EventType = "error"
	EventCanceled     EventType = "canceled"
)