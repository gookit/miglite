package migration

import (
	"database/sql"
	"fmt"

	"github.com/gookit/miglite/pkg/database"
)

// Status represents the status of a migration
type Status struct {
	Version string
	Status  string // up, down, pending
}

// GetMigrationStatus retrieves the status of all migrations
func GetMigrationStatus(db *database.DB, allMigrations []*Migration) ([]Status, error) {
	// Query the database for applied migrations
	rows, err := db.Query("SELECT version, status FROM db_schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to query migration status: %v", err)
	}
	defer rows.Close()

	// Create a map of applied migrations
	appliedMigrations := make(map[string]string)
	for rows.Next() {
		var version, status string
		if err := rows.Scan(&version, &status); err != nil {
			return nil, fmt.Errorf("failed to scan migration status: %v", err)
		}
		appliedMigrations[version] = status
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating migration status rows: %v", err)
	}

	// Create status list for all migrations
	var statuses []Status
	for _, migration := range allMigrations {
		if status, exists := appliedMigrations[migration.Version]; exists {
			statuses = append(statuses, Status{
				Version: migration.Version,
				Status:  status,
			})
		} else {
			statuses = append(statuses, Status{
				Version: migration.Version,
				Status:  "pending",
			})
		}
	}

	return statuses, nil
}

// GetAppliedMigrations retrieves only the applied migrations
func GetAppliedMigrations(db *database.DB) ([]MigrationRecord, error) {
	rows, err := db.Query("SELECT version, applied_at, status FROM db_schema_migrations ORDER BY applied_at")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %v", err)
	}
	defer rows.Close()

	var records []MigrationRecord
	for rows.Next() {
		var record MigrationRecord
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

// IsMigrationApplied checks if a specific migration has been applied
func IsMigrationApplied(db *database.DB, version string) (bool, string, error) {
	var status string
	err := db.QueryRow("SELECT status FROM db_schema_migrations WHERE version = ?", version).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, "", nil
		}
		return false, "", fmt.Errorf("failed to check migration status: %v", err)
	}
	return true, status, nil
}

// GetPendingMigrations returns migrations that have not been applied yet
func GetPendingMigrations(db *database.DB, allMigrations []*Migration) ([]*Migration, error) {
	appliedMigrations, err := GetAppliedMigrations(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %v", err)
	}

	// Create a set of applied migration versions
	appliedSet := make(map[string]bool)
	for _, record := range appliedMigrations {
		if record.Status == "up" {
			appliedSet[record.Version] = true
		}
	}

	// Find non-applied migrations
	var pending []*Migration
	for _, migration := range allMigrations {
		if !appliedSet[migration.Version] {
			pending = append(pending, migration)
		}
	}

	return pending, nil
}

// GetAppliedMigrationsSortedByDate returns applied migrations sorted by application date (most recent first)
func GetAppliedMigrationsSortedByDate(db *database.DB) ([]MigrationRecord, error) {
	rows, err := db.Query("SELECT version, applied_at, status FROM db_schema_migrations ORDER BY applied_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %v", err)
	}
	defer rows.Close()

	var records []MigrationRecord
	for rows.Next() {
		var record MigrationRecord
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