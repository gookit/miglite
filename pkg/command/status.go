package command

import (
	"fmt"

	"github.com/gookit/goutil/cflag"
	"github.com/gookit/miglite/pkg/config"
	"github.com/gookit/miglite/pkg/database"
	"github.com/gookit/miglite/pkg/migration"
)

// StatusCommand shows the status of migrations
func StatusCommand() *cflag.Cmd {
	// List applied and pending migrations
	c := cflag.NewCmd("status", "Show the status of migrations", handleStatus)

	c.BoolVar(&showVerbose, "verbose", false, "Enable verbose output;;v")
	c.StringVar(&configFile, "config", "./miglite.yaml", "Path to the configuration file;;c")

	return c
}

func handleStatus(c *cflag.Cmd) error {
	// Load configuration
	cfg, err := config.Load(configFile)
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

	// Get migration statuses
	statuses, err := migration.GetMigrationStatus(db, migrations)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %v", err)
	}

	// Print status table
	fmt.Println("Migration Status:")
	fmt.Println("=================")
	for _, status := range statuses {
		statusIcon := "[ ]" // pending
		if status.Status == "up" {
			statusIcon = "[X]" // applied
		} else if status.Status == "down" {
			statusIcon = "[R]" // rolled back
		}
		fmt.Printf("%s %s\n", statusIcon, status.Version)
	}

	return nil
}
