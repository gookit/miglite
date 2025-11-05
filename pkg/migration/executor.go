package migration

import (
	"fmt"
	"log"

	"github.com/gookit/miglite/pkg/database"
)

// Executor handles the execution of migrations
type Executor struct {
	db    *database.DB
	tracker *database.MigrationTracker
}

// NewExecutor creates a new migration executor
func NewExecutor(db *database.DB) *Executor {
	return &Executor{
		db:      db,
		tracker: database.NewMigrationTracker(db),
	}
}

// ExecuteUp executes the UP part of a migration
func (e *Executor) ExecuteUp(migration *Migration) error {
	// Start a transaction
	tx, err := e.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Execute the UP migration
	if _, err := tx.Exec(migration.UpMigration); err != nil {
		return fmt.Errorf("failed to execute UP migration: %v", err)
	}

	// Record the migration in the database
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	
	// Now record the migration status after successful commit
	if err := e.tracker.RecordMigration(migration.Version, "up"); err != nil {
		return fmt.Errorf("failed to record migration: %v", err)
	}

	log.Printf("Successfully executed migration: %s", migration.FileName)
	return nil
}

// ExecuteDown executes the DOWN part of a migration
func (e *Executor) ExecuteDown(migration *Migration) error {
	// Start a transaction
	tx, err := e.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Execute the DOWN migration if it exists
	if migration.DownMigration != "" {
		if _, err := tx.Exec(migration.DownMigration); err != nil {
			return fmt.Errorf("failed to execute DOWN migration: %v", err)
		}
	}

	// Record the migration rollback in the database
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	
	// Now record the migration status after successful commit
	if err := e.tracker.RecordMigration(migration.Version, "down"); err != nil {
		return fmt.Errorf("failed to record migration rollback: %v", err)
	}

	log.Printf("Successfully rolled back migration: %s", migration.FileName)
	return nil
}