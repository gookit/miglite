package command

import (
	"fmt"
	"strings"

	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/pkg/migration"
)

// StatusCommand shows the status of migrations
func StatusCommand() *capp.Cmd {
	// List applied and pending migrations
	c := capp.NewCmd("status", "Show the status of migrations", func(c *capp.Cmd) error {
		return HandleStatus()
	})
	c.Aliases = []string{"st", "list", "ls"}

	c.BoolVar(&ShowVerbose, "verbose", false, "Enable verbose output;;v")
	c.StringVar(&ConfigFile, "config", "./miglite.yaml", "Path to the configuration file;;c")

	return c
}

// HandleStatus display migration status
func HandleStatus() error {
	// Load configuration and connect to database
	cfg, db, err := initConfigAndDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Discover migrations
	migrations, err := migration.FindMigrations(cfg.Migrations.Path)
	if err != nil {
		return fmt.Errorf("failed to discover migrations: %v", err)
	}

	// Get migration statuses
	statuses, err := migration.GetMigrationsStatus(db, migrations)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %v", err)
	}

	// Print status table
	ccolor.Cyanf("\nMigration Status:(total=%d)\n", len(statuses))
	fmt.Println(strings.Repeat("==", 41))
	ccolor.Printf("   <b>Status</>   | %12s<b>Version(migration file)</>%12s    |   <b>Operate Time</> \n", "", "")
	fmt.Println(strings.Repeat("--", 41))

	for _, status := range statuses {
		statusIcon := "[<mga>pending</>]" // pending
		if status.Status == "up" {
			statusIcon = "[<green>applied</>]" // applied
		} else if status.Status == "down" {
			statusIcon = "[<ylw>rolled</>]" // rolled back
		}
		ccolor.Printf("  %s | %-50s | \n", statusIcon, status.Version)
	}

	return nil
}
