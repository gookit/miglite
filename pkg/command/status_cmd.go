package command

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/pkg/migration"
	"github.com/gookit/miglite/pkg/migutil"
)

// StatusOption status command option
type StatusOption struct {
}

// StatusCommand shows the status of migrations
func StatusCommand() *capp.Cmd {
	opt := StatusOption{}

	// List applied and pending migrations
	c := capp.NewCmd("status", "Show the status of migrations", func(c *capp.Cmd) error {
		return HandleStatus(opt)
	})

	c.Aliases = []string{"st"} // , "list", "ls"
	bindCommonFlags(c)

	return c
}

// HandleStatus display migration status
func HandleStatus(_ StatusOption) error {
	// Load configuration and connect to database
	if err := initConfigAndDB(); err != nil {
		return err
	}
	defer db.SilentClose()

	// Discover migrations
	migrations, err := findMigrations()
	if err != nil {
		return fmt.Errorf("failed to discover migrations: %v", err)
	}

	// Get migration statuses
	statuses, err := migration.GetMigrationsStatus(db, migrations)
	if err != nil {
		if migutil.IsTableNotExists(db.Driver(), err.Error()) {
			err = errors.New("migration table does not exist. please run `miglite init` to create it")
		}
		return err
	}

	// Print status table
	ccolor.Cyanf("\n📊  Migrations Status:(total=%d)\n", len(statuses))
	fmt.Println(strings.Repeat("==", 44))
	ccolor.Printf("  <b>Status</>  | %13s<b>Version(migration file)</>%13s    |   <b>Operate Time</> \n", "", "")
	fmt.Println(strings.Repeat("--", 44))

	for _, st := range statuses {
		statusIcon := "<mga>pending</>" // ⏳  pending
		if st.Status == "up" {
			statusIcon = "<green>applied</>" // ✅ applied
		} else if st.Status == "down" {
			statusIcon = "<ylw>rolled</> " // ↪️ rolled back
		} else if st.Status == "skip" {
			statusIcon = "<gray>skipped</>" // ⏭️ skipped
		}
		ccolor.Printf("  %s | %-52s | %s\n", statusIcon, st.Version, formatTime(st.AppliedAt))
	}

	return nil
}
