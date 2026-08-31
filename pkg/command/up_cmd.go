package command

import (
	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/cliutil"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/internal/runtime"
	"github.com/gookit/miglite/pkg/migration"
	"time"
)

type UpOption struct {
	Yes, SkipErr bool
	Number       int
	StartTime    string
}

func NewUpCommand() *capp.Cmd {
	var o UpOption
	c := capp.NewCmd("up", "Execute pending migrations", func(*capp.Cmd) error { return HandleUp(o) })
	c.Aliases = []string{"migrate", "run"}
	bindCommonFlags(c)
	c.BoolVar(&o.Yes, "yes", false, "Skip confirmation prompt;;y")
	c.IntVar(&o.Number, "number", 0, "Execute only the specified number of migrations;;n")
	c.BoolVar(&o.SkipErr, "skip-err", false, "Skip the error migration and continue with the execution;;s")
	return c
}
func HandleUp(o UpOption) error {
	r, cl, e := legacyRuntime()
	if e != nil {
		return e
	}
	defer cl()
	start := time.Now()
	h := runtime.MigrationHooks{
		Start: func(total int) {
			ccolor.Printf("🚀  Starting exec migrations(<green>founds=%d</>). Start at: %s\n\n", total, formatTime(start))
		},
		Complete: func(res runtime.MigrationResult) {
			ccolor.Printf("\n🎉  All migrations complete! 📘 apply:%d, skip:%d, failed:%d ⏱️ duration: %s\n", res.Applied, res.Skipped, res.Failed, time.Since(start))
		},
		Before: func(i, _ int, m *migration.Migration) error {
			ccolor.Printf("<green>%d.</> Executing migration: <green>%s</>\n", i+1, m.FileName)
			if !o.Yes && !cliutil.Confirm("Are you sure you want to execute this migration?") {
				ccolor.Magentaln("Exiting run migrations!")
				return runtime.ErrCancelled
			}
			return nil
		},
		Skip: func(_, _ int, m *migration.Migration, status string) {
			if !ShowVerbose {
				ccolor.Print(".")
				return
			}
			if status == migration.StatusSkip {
				ccolor.Printf("- Migration <gray>%s</> skipped\n", m.FileName)
			} else {
				ccolor.Printf("- Migration <gray>%s</> already applied\n", m.FileName)
			}
		},
		After: func(_, _ int, m *migration.Migration) {
			ccolor.Printf("✅ Successfully executed migration: %s\n", m.FileName)
		},
		Error: func(_, _ int, m *migration.Migration, err error) {
			ccolor.Errorf("❌ Failed migration %s: %v\n", m.FileName, err)
			ccolor.Printf("UpSQL:\n%s\n", m.UpSection)
		},
	}
	return r.UpWithHooks(runtime.UpOption{Yes: true, SkipErr: o.SkipErr, Number: o.Number, StartTime: o.StartTime}, h)
}
