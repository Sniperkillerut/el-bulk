package models

import (
	"time"
)

type JobStatus string

const (
	JobPending   JobStatus = "pending"
	JobRunning   JobStatus = "running"
	JobCompleted JobStatus = "completed"
	JobFailed    JobStatus = "failed"
)

type Job struct {
	ID          string     `db:"id" json:"id"`
	Type        string     `db:"type" json:"type"`
	Status      JobStatus  `db:"status" json:"status"`
	Progress    int        `db:"progress" json:"progress"`
	Payload     JSONB      `db:"payload" json:"payload,omitempty"`
	Result      JSONB      `db:"result" json:"result,omitempty"`
	Error       *string    `db:"error" json:"error,omitempty"`
	AdminID     *string    `db:"admin_id" json:"admin_id,omitempty"`
	StartedAt   *time.Time `db:"started_at" json:"started_at,omitempty"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
}
