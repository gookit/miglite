package command

import (
	"fmt"

	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/cliutil"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/pkg/database"
	"github.com/gookit/miglite/pkg/migration"
)

// DownOption represents the options for the down command
type DownOption struct {
	Number int
	// Yes 是否跳过确认
	Yes bool
}

// DownCommand rolls back the last migration or a specific one
func DownCommand() *capp.Cmd {
	var downOpt = DownOption{}
	c := capp.NewCmd("down", "Rollback the most recent migration", func(c *capp.Cmd) error {
		return HandleDown(downOpt)
	}).WithConfigFn(capp.WithAliases("rollback"))

	bindCommonFlags(c)

	c.BoolVar(&downOpt.Yes, "yes", false, "Skip confirmation prompt;;y")
	c.IntVar(&downOpt.Number, "number", 1, "Number of migrations to roll back;;n")
	return c
}

// HandleDown migration logic
func HandleDown(opt DownOption) error {
	// Load configuration and connect to database
	cfg, db, err := initConfigAndDB()
	if err != nil {
		return err
	}
	defer db.SilentClose()

	// Get applied migrations sorted by date (most recent first)
	appliedList, err := findAppliedMigrations(db, &opt)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %v", err)
	}
	if len(appliedList) == 0 {
		fmt.Println("🔎  No applied migrations to rollback")
		return nil
	}

	// Discover migrations
	count := opt.Number
	migrations, err := findMigrations()
	if err != nil {
		return fmt.Errorf("failed to discover migrations: %v", err)
	}

	// Get executor
	executor := migration.NewExecutor(db, ShowVerbose)
	confirmTip := "Are you sure you want to roll back the migration?"
	ccolor.Magentaf("🚀  Will roll back recent %d migrations:\n\n", count)

	// Roll back the specified number of migrations
	for i := 0; i < count; i++ {
		var applied = appliedList[i]
		// Find the corresponding migration file
		var targetMig *migration.Migration
		for _, mig := range migrations {
			if mig.Version == applied.Version {
				targetMig = mig
				break
			}
		}
		if targetMig == nil {
			return fmt.Errorf("migration file not found for version: %s", applied.Version)
		}

		ccolor.Printf("%d. Rolling back migration: <ylw>%s</> (appliedAt %s)\n", i+1, targetMig.FileName, formatTime(applied.AppliedAt))
		if !opt.Yes && !cliutil.Confirm(confirmTip) {
			ccolor.Warnln("Skipping rollback the migration!")
			continue
		}

		if err = targetMig.Parse(); err != nil {
			return err
		}

		// if down section is empty, skip
		if targetMig.DownSection == "" {
			ccolor.Warnln("Skipping empty DOWN migration!")
			continue
		}

		if err = executor.ExecuteDown(targetMig); err != nil {
			return fmt.Errorf(
				"failed to execute rollback for migration %s: %v.\nDownSQL:\n%s",
				targetMig.FileName, err, targetMig.DownSection,
			)
		}
		ccolor.Printf("✅  Success rolled back migration: %s\n", targetMig.FileName)
	}

	ccolor.Successf("\n🎉  Successfully rolled back %d migration(s)\n", count)
	return nil
}

func findAppliedMigrations(db *database.DB, opt *DownOption) ([]migration.Record, error) {
	// Get the target number of migrations to rollback (default 1)
	count := opt.Number
	if count <= 0 {
		return nil, fmt.Errorf("count must be greater than 0")
	}

	// Get applied migrations sorted by date (most recent first)
	appliedList, err := migration.GetAppliedSortedByDate(db, count)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %v", err)
	}

	// Limit the number of rollbacks to the available applied migrations
	if count > len(appliedList) {
		count = len(appliedList)
	}

	opt.Number = count
	return appliedList, nil
}
