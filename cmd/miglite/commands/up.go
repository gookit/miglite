package commands

import (
	"fmt"

	"github.com/gookit/goutil/cflag"
	"github.com/gookit/miglite/pkg/config"
	"github.com/gookit/miglite/pkg/database"
	"github.com/gookit/miglite/pkg/migration"
)

// UpCommand executes pending migrations
func UpCommand() func(c *cflag.Command) error {
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

		// Initialize schema if needed
		if err := db.InitSchema(); err != nil {
			return fmt.Errorf("failed to initialize schema: %v", err)
		}

		// Discover migrations
		migrations, err := migration.DiscoverMigrations(cfg.Migrations.Path)
		if err != nil {
			return fmt.Errorf("failed to discover migrations: %v", err)
		}

		// Get executor
		executor := migration.NewExecutor(db)

		// Execute pending migrations
		for _, mig := range migrations {
			// Check if migration is already applied
			applied, status, err := migration.IsMigrationApplied(db, mig.Version)
			if err != nil {
				return fmt.Errorf("failed to check migration status: %v", err)
			}
			
			if !applied || status == "down" {
				fmt.Printf("Executing migration: %s\n", mig.FileName)
				if err := executor.ExecuteUp(mig); err != nil {
					return fmt.Errorf("failed to execute migration %s: %v", mig.FileName, err)
				}
			} else {
				fmt.Printf("Skipping already applied migration: %s\n", mig.FileName)
			}
		}

		fmt.Println("All migrations applied successfully")
		return nil
	}
}