package commands

import (
	"fmt"

	"github.com/gookit/goutil/cflag"
	"github.com/gookit/miglite/pkg/config"
	"github.com/gookit/miglite/pkg/migration"
)

// CreateCommand creates a new migration file
func CreateCommand() func(c *cflag.Cmd) error {
	return func(c *cflag.Command) error {
		// Get the migration name from arguments
		if len(c.Args()) == 0 {
			return fmt.Errorf("migration name is required")
		}
		name := c.Args()[0]
		
		// Get configuration file path
		configPath := c.StringOpt("config", "./miglite.yaml")
		
		// Load configuration
		cfg, err := config.Load(configPath)
		if err != nil {
			// If config file doesn't exist, use defaults
			cfg = &config.Config{
				Migrations: struct {
					Path string `yaml:"path"`
				}{
					Path: "./migrations",
				},
			}
		}

		// Create the migration
		filePath, err := migration.CreateMigration(cfg.Migrations.Path, name)
		if err != nil {
			return fmt.Errorf("failed to create migration: %v", err)
		}

		fmt.Printf("Created migration: %s\n", filePath)
		return nil
	}
}