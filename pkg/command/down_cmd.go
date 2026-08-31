package command

import (
	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/cliutil"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/internal/runtime"
	"github.com/gookit/miglite/pkg/migration"
)

type DownOption struct {
	Number int
	Yes    bool
}

func DownCommand() *capp.Cmd {
	var o = DownOption{Number: 1}
	c := capp.NewCmd("down", "Rollback the most recent migration", func(*capp.Cmd) error { return HandleDown(o) })
	c.WithConfigFn(capp.WithAliases("rollback"))
	bindCommonFlags(c)
	c.BoolVar(&o.Yes, "yes", false, "Skip confirmation prompt;;y")
	c.IntVar(&o.Number, "number", 1, "Number of migrations to roll back;;n")
	return c
}
func HandleDown(o DownOption) error {
	r, cl, e := legacyRuntime()
	if e != nil {
		return e
	}
	defer cl()
	rolledBack := 0
	h := runtime.MigrationHooks{
		Start: func(total int) {
			if total == 0 {
				ccolor.Println("🔎  No applied migrations to rollback")
				return
			}
			ccolor.Printf("🚀  Will roll back recent %d migrations:\n\n", total)
		},
		Before: func(i, _ int, m *migration.Migration) error {
			ccolor.Printf("%d. Rolling back migration: <ylw>%s</> (appliedAt %s)\n", i+1, m.FileName, formatTime(m.AppliedAt))
			if !o.Yes && !cliutil.Confirm("Are you sure you want to roll back the migration?") {
				ccolor.Warnln("Skipping rollback the migration!")
				return runtime.ErrCancelled
			}
			return nil
		},
		After: func(_, _ int, m *migration.Migration) {
			rolledBack++
			ccolor.Printf("✅ Success rolled back migration: %s\n", m.FileName)
		},
		Skip: func(_, _ int, m *migration.Migration, status string) {
			if status == "empty_down" {
				ccolor.Warnf("Skipping empty DOWN migration! %s\n", m.FileName)
			}
		},
		Error: func(_, _ int, m *migration.Migration, err error) {
			ccolor.Errorf("Failed to roll back migration %s: %v\n", m.FileName, err)
		},
	}
	err := r.DownWithHooks(runtime.DownOption{Number: o.Number, Yes: true}, h)
	if err == nil && rolledBack > 0 {
		ccolor.Printf("\n🎉  Successfully rolled back %d migration(s)\n", rolledBack)
	}
	return err
}
