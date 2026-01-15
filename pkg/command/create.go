package command

import (
	"fmt"
	"strconv"

	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/cliutil"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/pkg/migration"
)

// CreateCommand creates a new migration file
func CreateCommand() *capp.Cmd {
	c := capp.NewCmd("create", "Create new migration SQL files", func(c *capp.Cmd) error {
		// Get the migration name from arguments
		return HandleCreate(c.Args())
	})

	c.Aliases = []string{"new"}
	bindCommonFlags(c)

	c.AddArg("name", "Migration name...", true, nil)

	return c
}

// HandleCreate creates migration files
func HandleCreate(names []string) error {
	if len(names) == 0 {
		return fmt.Errorf("no migration name provided")
	}

	// Load configuration
	if err := initLoadConfig(); err != nil {
		return err
	}

	migPaths := cfg.Migrations.GetPaths()
	migPath := migPaths[0]
	if ln := len(migPaths); ln > 1 {
		ccolor.Infof("📢 Multiple migration paths found: %v\n", migPaths)
		for i, p := range migPaths {
			ccolor.Cyanf(" %d) %s\n", i+1, p)
		}
		str, err := cliutil.ReadLine("which one do you want to use? (default: 1) ")
		if err != nil {
			return err
		}

		// convert to int value
		if intVal, err1 := strconv.Atoi(str); err1 == nil && intVal > 1 {
			index := intVal - 1
			if index < ln {
				migPath = migPaths[index]
			} else {
				ccolor.Warnf("Invalid index: %d, Exit!\n", intVal)
				return nil
			}
		}
	}

	// Create the migration
	filePaths, err := migration.CreateMigrations(migPath, names)
	if err != nil {
		return fmt.Errorf("failed to create migration: %v", err)
	}

	ccolor.Successln("Created migrations:")
	for _, filePath := range filePaths {
		ccolor.Printf("  - %s\n", filePath)
	}
	return nil

}
