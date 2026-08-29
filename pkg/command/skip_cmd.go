package command

import (
	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/miglite/internal/runtime"
)

// SkipOption skip migration file option
type SkipOption struct {
	FileNames []string
}

// SkipCommand skips one or multi migration file(s)
func SkipCommand() *capp.Cmd {
	c := capp.NewCmd("skip", "Manual skip one or multi migration file(s)", func(c *capp.Cmd) error {
		return HandleSkip(SkipOption{
			FileNames: c.Arg("files").Strings(),
		})
	})
	c.WithConfigFn(capp.WithAliases("ignore"))

	bindCommonFlags(c)
	c.AddArg("files", "Migration filename(s) to skip, allow multi", true, nil)

	return c
}

// HandleSkip skips one or multi migration file(s)
func HandleSkip(opt SkipOption) error {
	r, cleanup, err := legacyRuntime()
	if err != nil {
		return err
	}
	defer cleanup()
	return r.Skip(runtime.SkipOption{FileNames: opt.FileNames})
	/*
	   	if err := initConfigAndDB(); err != nil {
	   		return err
	   	}

	   defer cleanupDB()

	   migFiles, err := migration.MigrationsFrom(cfg.Migrations.Path, opt.FileNames)

	   	if err != nil {
	   		return err
	   	}

	   // get migration status from database
	   records, err := migration.GetMigrationsStatus(db, migFiles)

	   	if err != nil {
	   		return err
	   	}

	   	recordMap := arrutil.ToMap(records, func(item migration.Record) (string, migration.Record) {
	   		return item.Version, item
	   	})

	   ccolor.Magentaf("🚀  Start ignore %d migrations:\n\n", len(migFiles))

	   	for _, migFile := range migFiles {
	   		if record, ok := recordMap[migFile.Version]; ok {
	   			if record.Status == migration.StatusUp {
	   				ccolor.Warnf("Migration %s already skipped", migFile.Version)
	   				continue
	   			}
	   		}

	   		// update migration status to skipped
	   		err = migration.SaveRecord(db, migFile.Version, migration.StatusSkip, nil)
	   		if err != nil {
	   			return err
	   		}
	   		ccolor.Printf("- Migration <green>%s</> skipped", migFile.Version)
	   	}

	   return nil
	*/
}
