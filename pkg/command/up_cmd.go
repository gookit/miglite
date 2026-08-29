package command

import (
	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/cliutil"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/internal/runtime"
	"github.com/gookit/miglite/pkg/migration"
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
	h := runtime.MigrationHooks{Start: func(total int) { ccolor.Printf("🚀 Starting exec migrations(founds=%d)\n", total) }, Complete: func(res runtime.MigrationResult) {
		ccolor.Printf("🎉 All migrations applied successfully! apply:%d, skip:%d\n", res.Applied, res.Skipped)
	}, Before: func(i, _ int, m *migration.Migration) error {
		ccolor.Printf("<green>%d.</> Executing migration: <green>%s</>\n", i+1, m.FileName)
		if !o.Yes && !cliutil.Confirm("Are you sure you want to execute this migration?") {
			return runtime.ErrCancelled
		}
		return nil
	}, After: func(_, _ int, m *migration.Migration) {
		ccolor.Printf("✅ Successfully executed migration: %s\n", m.FileName)
	}}
	return r.UpWithHooks(runtime.UpOption{Yes: true, SkipErr: o.SkipErr, Number: o.Number, StartTime: o.StartTime}, h)
}
