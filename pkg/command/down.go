package command

import (
	"fmt"

	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/miglite/pkg/migration"
)

var downCmdOpt = struct {
	count int
}{}

// DownCommand rolls back the last migration or a specific one
func DownCommand() *capp.Cmd {
	c := capp.NewCmd("down", "Rollback the most recent migration", handleDown)

	c.BoolVar(&showVerbose, "verbose", false, "Enable verbose output;;v")
	c.StringVar(&configFile, "config", "./miglite.yaml", "Path to the configuration file;;c")

	c.IntVar(&downCmdOpt.count, "count", 1, "Number of migrations to roll back;;c")
	return c
}

func handleDown(c *capp.Cmd) error {
	// Load configuration and connect to database
	cfg, db, err := initConfigAndDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Discover migrations
	migrations, err := migration.DiscoverMigrations(cfg.Migrations.Path)
	if err != nil {
		return fmt.Errorf("failed to discover migrations: %v", err)
	}

	// Get the target number of migrations to rollback (default 1)
	count := downCmdOpt.count
	if count <= 0 {
		return fmt.Errorf("count must be greater than 0")
	}

	// Get applied migrations sorted by date (most recent first)
	appliedMigrations, err := migration.GetAppliedMigrationsSortedByDate(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %v", err)
	}

	if len(appliedMigrations) == 0 {
		fmt.Println("No applied migrations to rollback")
		return nil
	}

	// Limit the number of rollbacks to the available applied migrations
	if count > len(appliedMigrations) {
		count = len(appliedMigrations)
	}

	// Get executor
	executor := migration.NewExecutor(db)

	// Roll back the specified number of migrations
	for i := 0; i < count; i++ {
		// Find the corresponding migration file
		var targetMigration *migration.Migration
		for _, mig := range migrations {
			if mig.Version == appliedMigrations[i].Version {
				targetMigration = mig
				break
			}
		}

		if targetMigration == nil {
			return fmt.Errorf("migration file not found for version: %s", appliedMigrations[i].Version)
		}

		fmt.Printf("Rolling back migration: %s\n", targetMigration.FileName)
		if err := executor.ExecuteDown(targetMigration); err != nil {
			return fmt.Errorf("failed to execute rollback for migration %s: %v", targetMigration.FileName, err)
		}
	}

	fmt.Printf("Successfully rolled back %d migration(s)\n", count)
	return nil

}
