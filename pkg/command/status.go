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

	c.Aliases = []string{"st"} // , "list", "ls"
	bindCommonFlags(c)

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
	ccolor.Cyanf("\n📊  Migrations Status:(total=%d)\n", len(statuses))
	fmt.Println(strings.Repeat("==", 45))
	ccolor.Printf("    <b>Status</>   | %13s<b>Version(migration file)</>%13s    |   <b>Operate Time</> \n", "", "")
	fmt.Println(strings.Repeat("--", 45))

	for _, st := range statuses {
		statusIcon := "⏳  <mga>pending</>" // pending
		if st.Status == "up" {
			statusIcon = "✅  <green>applied</>" // applied
		} else if st.Status == "down" {
			statusIcon = "↪️  <ylw>rolled</> " // rolled back
		} else if st.Status == "skip" {
			statusIcon = "⏭️  <ylw>skipped</>" // skipped
		}
		ccolor.Printf("  %s | %-52s | %s\n", statusIcon, st.Version, formatTime(st.AppliedAt))
	}

	return nil
}
