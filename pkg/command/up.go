package command

import (
	"fmt"

	"github.com/gookit/goutil/cflag"
	"github.com/gookit/miglite/pkg/config"
	"github.com/gookit/miglite/pkg/database"
	"github.com/gookit/miglite/pkg/migration"
)

var upCmdOpt = struct {
}{}

// NewUpCommand executes pending migrations
func NewUpCommand() *cflag.Cmd {
	c := &cflag.Cmd{
		Name: "up",
		Desc: "Execute pending migrations",
		Func: handleUp,
	}

	c.BoolVar(&showVerbose, "verbose", false, "Enable verbose output;;v")
	c.StringVar(&configFile, "config", "./miglite.yaml", "Path to the configuration file;;c")

	return c
}

func handleUp(c *cflag.Cmd) error {
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

	// Initialize schema if needed
	if err := db.InitSchema(); err != nil {
		return fmt.Errorf("failed to initialize schema: %v", err)
	}

	// Discover migrations
	migrations, err := migration.DiscoverMigrations(cfg.Migrations.Path)
	if err != nil {
		return fmt.Errorf("failed to discover migrations: %v", err)
	}

	// Get executor
	executor := migration.NewExecutor(db)

	// Execute pending migrations
	for _, mig := range migrations {
		// Check if migration is already applied
		applied, status, err := migration.IsMigrationApplied(db, mig.Version)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %v", err)
		}

		if !applied || status == "down" {
			fmt.Printf("Executing migration: %s\n", mig.FileName)
			if err := executor.ExecuteUp(mig); err != nil {
				return fmt.Errorf("failed to execute migration %s: %v", mig.FileName, err)
			}
		} else {
			fmt.Printf("Skipping already applied migration: %s\n", mig.FileName)
		}
	}

	fmt.Println("All migrations applied successfully")
	return nil
}
