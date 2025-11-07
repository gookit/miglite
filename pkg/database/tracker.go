package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// MigrationTracker handles tracking of applied migrations
type MigrationTracker struct {
	db *DB
	// verbose flag
	verbose bool
}

// NewMigrationTracker creates a new migration tracker
func NewMigrationTracker(db *DB, verbose bool) *MigrationTracker {
	return &MigrationTracker{db: db, verbose: verbose}
}

// SaveRecord records a migration in the database
//  - status=up: insert a new record
//  - status=down: update the record
func (mt *MigrationTracker) SaveRecord(version, status string) error {
	// Check if the record already exists
	var exists bool
	err := mt.db.QueryRow("SELECT EXISTS(SELECT 1 FROM db_schema_migrations WHERE version = ?)", version).Scan(&exists)
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

// GetStatus retrieves the status of a specific migration
func (mt *MigrationTracker) GetStatus(version string) (string, error) {
	var status string
	err := mt.db.QueryRow("SELECT status FROM db_schema_migrations WHERE version = ?", version).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil // Migration not applied yet
		}
		return "", fmt.Errorf("failed to get migration status: %v", err)
	}
	return status, nil
}

// GetAllStatus retrieves the status of all migrations
func (mt *MigrationTracker) GetAllStatus() (map[string]string, error) {
	rows, err := mt.db.Query("SELECT version, status FROM db_schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to query migration status: %v", err)
	}
	defer rows.Close()

	statuses := make(map[string]string)
	for rows.Next() {
		var version, status string
		if err := rows.Scan(&version, &status); err != nil {
			return nil, fmt.Errorf("failed to scan migration status: %v", err)
		}
		statuses[version] = status
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating migration status rows: %v", err)
	}
	return statuses, nil
}

// RemoveMigration removes a migration record from the database
func (mt *MigrationTracker) RemoveMigration(version string) error {
	_, err := mt.db.Exec("DELETE FROM db_schema_migrations WHERE version = ?", version)
	if err != nil {
		return fmt.Errorf("failed to remove migration record: %v", err)
	}
	return nil
}