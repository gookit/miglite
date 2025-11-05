package command

import (
	"fmt"

	"github.com/gookit/goutil/cflag"
	"github.com/gookit/miglite/pkg/config"
	"github.com/gookit/miglite/pkg/migration"
)

// CreateCommand creates a new migration file
func CreateCommand() *cflag.Cmd {
	c := cflag.NewCmd("create", "Create a new migration", handleCreate)

	c.BoolVar(&showVerbose, "verbose", false, "Enable verbose output;;v")
	c.StringVar(&configFile, "config", "./miglite.yaml", "Path to the configuration file;;c")

	c.AddArg("name", "Migration name", true, nil)

	return c
}

func handleCreate(c *cflag.Cmd) error {
	// Get the migration name from arguments
	name := c.Arg("name").String()

	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		cfg = config.Default()
	}

	// Create the migration
	filePath, err := migration.CreateMigration(cfg.Migrations.Path, name)
	if err != nil {
		return fmt.Errorf("failed to create migration: %v", err)
	}

	fmt.Printf("Created migration: %s\n", filePath)
	return nil

}
