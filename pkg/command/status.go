package command

import (
	"fmt"

	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/pkg/migration"
)

// StatusCommand shows the status of migrations
func StatusCommand() *capp.Cmd {
	// List applied and pending migrations
	c := capp.NewCmd("status", "Show the status of migrations", handleStatus)

	c.BoolVar(&showVerbose, "verbose", false, "Enable verbose output;;v")
	c.StringVar(&configFile, "config", "./miglite.yaml", "Path to the configuration file;;c")

	return c
}

func handleStatus(c *capp.Cmd) error {
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

	// Get migration statuses
	statuses, err := migration.GetMigrationStatus(db, migrations)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %v", err)
	}

	// Print status table
	ccolor.Cyanln("Migration Status:")
	fmt.Println("=================")
	fmt.Println(" Status  | Version | Operate Time ")
	for _, status := range statuses {
		statusIcon := "[ ]" // pending
		if status.Status == "up" {
			statusIcon = "[<green>X</>]" // applied
		} else if status.Status == "down" {
			statusIcon = "[<ylw>R</>]" // rolled back
		}
		ccolor.Printf("%s | %s\n", statusIcon, status.Version)
	}

	return nil
}
