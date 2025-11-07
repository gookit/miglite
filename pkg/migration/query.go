package migration

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gookit/miglite/pkg/database"
)

// Status represents the status of a migration
type Status struct {
	Version   string
	Status    string // up, down, pending, skip
	AppliedAt time.Time
}

// GetMigrationsStatus retrieves the status of all migrations
func GetMigrationsStatus(db *database.DB, allMigrations []*Migration) ([]Status, error) {
	// Query the database for applied migrations
	rows, err := db.Query("SELECT version, status, applied_at FROM db_schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to query migration status: %v", err)
	}
	defer rows.Close()

	// Create status list for all migrations
	var statuses []Status

	// Create a map of applied migrations
	appliedMigrations := make(map[string]Status)
	for rows.Next() {
		var appliedAt time.Time
		var version, status string
		if err := rows.Scan(&version, &status, &appliedAt); err != nil {
			return nil, fmt.Errorf("failed to scan migration status: %v", err)
		}
		appliedMigrations[version] = Status{
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
			statuses = append(statuses, Status{
				Version: migration.FileName,
				Status: StatusPending,
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

// IsApplied checks if a specific migration has been applied
func IsApplied(db *database.DB, version string) (bool, string, error) {
	var status string
	err := db.QueryRow("SELECT status FROM db_schema_migrations WHERE version = ?", version).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, "", nil
		}
		return false, "", fmt.Errorf("failed to check migration status: %v", err)
	}
	return true, status, nil
}

// GetAppliedSortedByDate returns applied migrations sorted by application date (most recent first)
func GetAppliedSortedByDate(db *database.DB) ([]Record, error) {
	rows, err := db.Query("SELECT version, applied_at, status FROM db_schema_migrations ORDER BY applied_at DESC")
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
