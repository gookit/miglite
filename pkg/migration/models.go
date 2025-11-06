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

// Record represents a record in the database migrations table
type Record struct {
	Version string `db:"version"` // is migration filename
	AppliedAt time.Time `db:"applied_at"`
	Status string `db:"status"`       // up, skip, down.
	Hash      string    `db:"hash"`   // optional hash of migration content
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
