package command

import (
	"fmt"

	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/pkg/migration"
)

// CreateCommand creates a new migration file
func CreateCommand() *capp.Cmd {
	c := capp.NewCmd("create", "Create new migration SQL files", handleCreate)
	c.Aliases = []string{"new"}

	c.BoolVar(&showVerbose, "verbose", false, "Enable verbose output;;v")
	c.StringVar(&configFile, "config", "./miglite.yaml", "Path to the configuration file;;c")

	c.AddArg("name", "Migration name...", true, nil)

	return c
}

func handleCreate(c *capp.Cmd) error {
	// Load configuration
	cfg, err := initLoadConfig()
	if err != nil {
		return err
	}

	// Get the migration name from arguments
	names := c.Args()

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
