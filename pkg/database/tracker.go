package database

import (
	"database/sql"
	"fmt"
)

// MigrationTracker handles tracking of applied migrations
type MigrationTracker struct {
	db *DB
}

// NewMigrationTracker creates a new migration tracker
func NewMigrationTracker(db *DB) *MigrationTracker {
	return &MigrationTracker{db: db}
}

// RecordMigration records a migration in the database
func (mt *MigrationTracker) RecordMigration(version, status string) error {
	// Check if the record already exists
	var exists bool
	err := mt.db.QueryRow("SELECT EXISTS(SELECT 1 FROM db_schema_migrations WHERE version = ?)", version).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if migration exists: %v", err)
	}

	var query string
	if exists {
		// Update the existing record
		query = "UPDATE db_schema_migrations SET applied_at = CURRENT_TIMESTAMP, status = ? WHERE version = ?"
	} else {
		// Insert a new record
		query = "INSERT INTO db_schema_migrations (version, status) VALUES (?, ?)"
	}

	_, err = mt.db.Exec(query, status, version)
	if err != nil {
		return fmt.Errorf("failed to record migration: %v", err)
	}

	return nil
}

// GetMigrationStatus retrieves the status of a specific migration
func (mt *MigrationTracker) GetMigrationStatus(version string) (string, error) {
	var status string
	err := mt.db.QueryRow("SELECT status FROM db_schema_migrations WHERE version = ?", version).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // Migration not applied yet
		}
		return "", fmt.Errorf("failed to get migration status: %v", err)
	}
	return status, nil
}

// GetAllMigrationStatus retrieves the status of all migrations
func (mt *MigrationTracker) GetAllMigrationStatus() (map[string]string, error) {
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