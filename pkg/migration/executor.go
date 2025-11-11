package migration

import (
	"fmt"
	"log"

	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/pkg/database"
)

// Executor handles the execution of migrations
type Executor struct {
	db *database.DB
	// verbose flag
	verbose bool
	tracker *Tracker
}

// NewExecutor creates a new migration executor
func NewExecutor(db *database.DB, verbose bool) *Executor {
	return &Executor{
		db:      db,
		verbose: verbose,
		tracker: NewTracker(db, verbose),
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
			if err1 := tx.Rollback(); err1 != nil {
				log.Printf("[ERROR] Failed to rollback transaction: %v", err1)
			}
		}
	}()

	if e.verbose {
		ccolor.Printf("Executing migration UP Section: %s", migration.UpSection)
	}

	// Execute the UP migration
	if _, err := tx.Exec(migration.UpSection); err != nil {
		return fmt.Errorf("failed to execute UP migration: %v", err)
	}

	// Record the migration in the database
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Now record the migration status after successful commit
	if err := e.tracker.SaveRecord(migration.Version, StatusUp); err != nil {
		return err
	}
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
			if err1 := tx.Rollback(); err1 != nil {
				log.Printf("[ERROR] Failed to rollback transaction: %v", err1)
			}
		}
	}()

	if e.verbose {
		ccolor.Printf("Executing migration DOWN Section: %s", migration.DownSection)
	}

	// Execute the DOWN migration
	if _, err := tx.Exec(migration.DownSection); err != nil {
		return fmt.Errorf("failed to execute DOWN migration: %v", err)
	}

	// Record the migration rollback in the database
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Now record the migration status after successful commit
	if err := e.tracker.SaveRecord(migration.Version, StatusDown); err != nil {
		return err
	}

	log.Printf("Successfully rolled back migration: %s", migration.FileName)
	return nil
}
