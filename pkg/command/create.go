package command

import (
	"fmt"

	"github.com/gookit/goutil/cflag/capp"
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

	// Create the migration
	filePaths, err := migration.CreateMigrations(cfg.Migrations.Path, names)
	if err != nil {
		return fmt.Errorf("failed to create migration: %v", err)
	}

	ccolor.Successln("Created migrations:")
	for _, filePath := range filePaths {
		ccolor.Printf("  - %s\n", filePath)
	}
	return nil

}
