package migration

import (
	"time"
)

const (
	// StatusUp represents an up migration status
	StatusUp = "up"
	// StatusDown represents a down migration status
	StatusDown = "down"
	// StatusSkip represents a skipped migration status
	StatusSkip = "skip"
	// StatusPending represents a pending migration status
	StatusPending = "pending"
)

const (
	MarkUp   = "-- Migrate:UP"
	MarkDown = "-- Migrate:DOWN"
	// DateLayout defines the layout for migration filename
	DateLayout   = "20060102-150405"
	PrefixFormat = "YYYYMMDD-HHMMSS"
)

// StatusText returns the text representation of a migration status
func StatusText(status string) string {
	switch status {
	case StatusUp:
		return "applied"
	case StatusDown:
		return "rolled"
	case StatusSkip:
		return "skipped"
	case StatusPending:
		return "pending"
	default:
		return "unknown"
	}
}

// Record represents a record in the database migrations table
type Record struct {
	// is migration filename
	Version   string    `db:"version"`
	AppliedAt time.Time `db:"applied_at"`
	// up, skip, down.
	Status string `db:"status"`
}

// NewRecord creates a new migration record
func NewRecord(version string, status string) *Record {
	return &Record{
		Version:   version,
		AppliedAt: time.Now(),
		Status:    status,
	}
}

// SetStatus updates the status of the migration record
func (mr *Record) SetStatus(status string) {
	mr.Status = status
	mr.AppliedAt = time.Now()
}
