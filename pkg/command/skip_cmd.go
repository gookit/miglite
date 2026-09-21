package command

import (
	"github.com/gookit/goutil/cflag/capp"
)

type SkipOption struct{ FileNames []string }

func SkipCommand() *capp.Cmd {
	c := capp.NewCmd("skip", "Manual skip one or multi migration file(s)", func(c *capp.Cmd) error {
		return HandleSkip(SkipOption{FileNames: c.Arg("files").Strings()})
	})
	c.WithConfigFn(capp.WithAliases("ignore"))
	bindCommonFlags(c)
	c.AddArg("files", "Migration filename(s) to skip, allow multi", true, nil)
	return c
}

// HandleSkip marks migration files as skipped
func HandleSkip(opt SkipOption) error {
	r, cl, e := legacyRuntime()
	if e != nil {
		return e
	}
	defer cl()
	return RunSkip(r, opt)
}
