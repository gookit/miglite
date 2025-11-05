package migration

import (
	"time"
)

// MigrationRecord represents a record in the database migrations table
type MigrationRecord struct {
	Version   string    `db:"version"`
	AppliedAt time.Time `db:"applied_at"`
	Status    string    `db:"status"` // up, skip, down
	Hash      string    `db:"hash"`   // optional hash of migration content
}

// NewMigrationRecord creates a new migration record
func NewMigrationRecord(version string, status string) *MigrationRecord {
	return &MigrationRecord{
		Version:   version,
		AppliedAt: time.Now(),
		Status:    status,
	}
}

// SetStatus updates the status of the migration record
func (mr *MigrationRecord) SetStatus(status string) {
	mr.Status = status
	mr.AppliedAt = time.Now()
}