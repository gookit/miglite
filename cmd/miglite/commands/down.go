package commands

import (
	"fmt"

	"github.com/gookit/goutil/cflag"
	"github.com/gookit/miglite/pkg/config"
	"github.com/gookit/miglite/pkg/database"
	"github.com/gookit/miglite/pkg/migration"
)

// DownCommand rolls back the last migration or a specific one
func DownCommand() func(c *cflag.Command) error {
	return func(c *cflag.Command) error {
		// Get configuration file path
		configPath := c.StringOpt("config", "./miglite.yaml")
		
		// Load configuration
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}

		// Connect to database
		db, err := database.Connect(cfg.Database.Driver, cfg.Database.DSN)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %v", err)
		}
		defer db.Close()

		// Discover migrations
		migrations, err := migration.DiscoverMigrations(cfg.Migrations.Path)
		if err != nil {
			return fmt.Errorf("failed to discover migrations: %v", err)
		}

		// Get the target number of migrations to rollback (default 1)
		count := c.IntOpt("count", 1)
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
}