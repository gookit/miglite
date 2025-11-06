package command

import (
	"fmt"

	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/pkg/migration"
)

var upCmdOpt = struct {
	// 默认每执行一个都需要确认
	yes bool
	// 跳过错误迁移并继续执行
	skipErr bool
}{}

// NewUpCommand executes pending migrations
func NewUpCommand() *capp.Cmd {
	c := capp.NewCmd("up", "Execute pending migrations", func(c *capp.Cmd) error {
		return handleUp()
	})
	c.Aliases = []string{"migrate", "run"}

	c.BoolVar(&showVerbose, "verbose", false, "Enable verbose output;;v")
	c.StringVar(&configFile, "config", "./miglite.yaml", "Path to the configuration file;;c")
	c.BoolVar(&upCmdOpt.yes, "yes", false, "Skip confirmation prompt;;y")
	c.BoolVar(&upCmdOpt.skipErr, "skip-err", false, "Skip the error migration and continue with the execution;;s")

	return c
}

func handleUp() error {
	// Load configuration and connect to database
	cfg, db, err := initConfigAndDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Initialize schema if needed
	if err := db.InitSchema(); err != nil {
		return fmt.Errorf("failed to initialize schema: %v", err)
	}

	// Discover migrations
	migrations, err := migration.DiscoverMigrations(cfg.Migrations.Path)
	if err != nil {
		return fmt.Errorf("failed to discover migrations: %v", err)
	}

	if len(migrations) == 0 {
		ccolor.Infoln("No migrations found")
		return nil
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
