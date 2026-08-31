package command

import (
	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/internal/runtime"
	"github.com/gookit/miglite/pkg/migration"
)

type SkipOption struct { FileNames []string }

func SkipCommand() *capp.Cmd {
	c := capp.NewCmd("skip", "Manual skip one or multi migration file(s)", func(c *capp.Cmd) error {
		return HandleSkip(SkipOption{FileNames: c.Arg("files").Strings()})
	})
	c.WithConfigFn(capp.WithAliases("ignore"))
	bindCommonFlags(c)
	c.AddArg("files", "Migration filename(s) to skip, allow multi", true, nil)
	return c
}

func HandleSkip(opt SkipOption) error {
	r, cleanup, err := legacyRuntime()
	if err != nil { return err }
	defer cleanup()
	ccolor.Magentaf("🚀  Start ignore %d migrations:\n\n", len(opt.FileNames))
	return r.SkipWithHooks(runtime.SkipOption{FileNames: opt.FileNames}, runtime.MigrationHooks{
		After: func(_, _ int, m *migration.Migration) { ccolor.Printf("- Migration <green>%s</> skipped\n", m.Version) },
		Skip: func(_, _ int, m *migration.Migration, status string) {
			if status == migration.StatusUp { ccolor.Warnf("Migration %s already applied\n", m.Version) }
		},
	})
}
