package migration

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gookit/miglite/pkg/database"
)

// Tracker handles tracking of applied migrations
type Tracker struct {
	db *database.DB
	// verbose flag
	verbose bool
}

// NewTracker creates a new migration tracker
func NewTracker(db *database.DB, verbose bool) *Tracker {
	return &Tracker{db: db, verbose: verbose}
}

// SaveRecord records a migration in the database
//  - status=up: insert a new record
//  - status=down: update the record
func (mt *Tracker) SaveRecord(version, status string) error {
	builder, err := mt.db.SqlBuilder()
	if err != nil {
		return err
	}

	// Check if the record already exists
	var exists bool
	err = mt.db.QueryRow(builder.QueryExists(), version).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if migration exists: %v", err)
	}

	var aSql string
	var args = []any{version, status}
	if exists {
		// Update the existing record. eg: up -> down
		aSql = "UPDATE db_schema_migrations SET applied_at = CURRENT_TIMESTAMP, status = ? WHERE version = ?"
		args = []any{status, version} // parameter order must be same as query
	} else {
		// Insert a new record
		aSql = "INSERT INTO db_schema_migrations (version, status) VALUES (?, ?)"
	}

	_, err = mt.db.Exec(aSql, args...)
	if err != nil {
		return fmt.Errorf("failed to record migration: %v", err)
	}
	return nil
}

// GetMigrationsStatus retrieves the status of all migrations
func GetMigrationsStatus(db *database.DB, allMigrations []*Migration) ([]Record, error) {
	builder, err := db.SqlBuilder()
	if err != nil {
		return nil, err
	}

	// Query the database for applied migrations
	rows, err := db.Query(builder.QueryAll())
	if err != nil {
		return nil, fmt.Errorf("failed to query migration status: %v", err)
	}
	defer rows.Close()

	// Create status list for all migrations
	var statuses []Record

	// Create a map of applied migrations
	appliedMigrations := make(map[string]Record)
	for rows.Next() {
		var appliedAt time.Time
		var version, status string
		if err := rows.Scan(&version, &status, &appliedAt); err != nil {
			return nil, fmt.Errorf("failed to scan migration status: %v", err)
		}
		appliedMigrations[version] = Record{
			Version:   version,
			Status:    status,
			AppliedAt: appliedAt,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating migration status rows: %v", err)
	}

	for _, migration := range allMigrations {
		if status, exists := appliedMigrations[migration.Version]; exists {
			statuses = append(statuses, status)
		} else {
			statuses = append(statuses, Record{
				Version: migration.FileName,
				Status:  StatusPending,
			})
		}
	}

	return statuses, nil
}

// GetAppliedMigrations retrieves only the applied migrations
func GetAppliedMigrations(db *database.DB) ([]Record, error) {
	rows, err := db.Query("SELECT version, applied_at, status FROM db_schema_migrations ORDER BY applied_at")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %v", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		err := rows.Scan(&record.Version, &record.AppliedAt, &record.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan migration record: %v", err)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating migration record rows: %v", err)
	}
	return records, nil
}

// IsApplied checks if a specific migration has been applied(status=up)
func IsApplied(db *database.DB, version string) (bool, string, error) {
	var status string
	err := db.QueryRow("SELECT status FROM db_schema_migrations WHERE version = ?", version).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, "", nil
		}
		return false, "", fmt.Errorf("failed to check migration status: %v", err)
	}
	return status == StatusUp, status, nil
}

// GetAppliedSortedByDate returns applied migrations sorted by application date (most recent first)
func GetAppliedSortedByDate(db *database.DB, limit int) ([]Record, error) {
	rows, err := db.Query(
		"SELECT version, applied_at FROM db_schema_migrations WHERE status=? ORDER BY applied_at DESC LIMIT ?",
		StatusUp,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %v", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		err := rows.Scan(&record.Version, &record.AppliedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan migration record: %v", err)
		}
		record.Status = StatusUp
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating migration record rows: %v", err)
	}

	return records, nil
}
