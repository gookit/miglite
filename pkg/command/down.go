package command

import (
	"fmt"

	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/pkg/migration"
)

// DownOption represents the options for the down command
type DownOption struct {
	Count int
}

// DownCommand rolls back the last migration or a specific one
func DownCommand() *capp.Cmd {
	var downOpt = DownOption{}
	c := capp.NewCmd("down", "Rollback the most recent migration", func(c *capp.Cmd) error {
		return HandleDown(downOpt)
	})

	c.BoolVar(&ShowVerbose, "verbose", false, "Enable verbose output;;v")
	c.StringVar(&ConfigFile, "config", "./miglite.yaml", "Path to the configuration file;;c")

	c.IntVar(&downOpt.Count, "count", 1, "Number of migrations to roll back;;c")
	return c
}

// HandleDown migration logic
func HandleDown(opt DownOption) error {
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

	// Get the target number of migrations to rollback (default 1)
	count := opt.Count
	if count <= 0 {
		return fmt.Errorf("count must be greater than 0")
	}

	// Get applied migrations sorted by date (most recent first)
	appliedList, err := migration.GetAppliedSortedByDate(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %v", err)
	}

	if len(appliedList) == 0 {
		fmt.Println("No applied migrations to rollback")
		return nil
	}

	// Limit the number of rollbacks to the available applied migrations
	if count > len(appliedList) {
		count = len(appliedList)
	}

	// Get executor
	executor := migration.NewExecutor(db, ShowVerbose)

	// Roll back the specified number of migrations
	for i := 0; i < count; i++ {
		// Find the corresponding migration file
		var targetMig *migration.Migration
		for _, mig := range migrations {
			if mig.Version == appliedList[i].Version {
				targetMig = mig
				break
			}
		}

		if targetMig == nil {
			return fmt.Errorf("migration file not found for version: %s", appliedList[i].Version)
		}

		fmt.Printf("Rolling back migration: %s\n", targetMig.FileName)
		if err := targetMig.Parse(); err != nil {
			return err
		}

		// if down section is empty, skip
		if targetMig.DownSection == "" {
			ccolor.Warnln("Skipping empty down migration!")
			continue
		}

		if err := executor.ExecuteDown(targetMig); err != nil {
			return fmt.Errorf("failed to execute rollback for migration %s: %v", targetMig.FileName, err)
		}
	}

	fmt.Printf("Successfully rolled back %d migration(s)\n", count)
	return nil

}
