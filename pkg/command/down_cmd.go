package command

import (
	"github.com/gookit/goutil/cflag/capp"
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

// HandleDown rolls back applied migrations
func HandleDown(o DownOption) error {
	r, cl, e := legacyRuntime()
	if e != nil {
		return e
	}
	defer cl()
	return RunDown(r, o, true)
}
