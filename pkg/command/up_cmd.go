package command

import (
	"fmt"

	"github.com/gookit/goutil/cflag/capp"
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
	c.BoolVar(&o.SkipErr, "skip-err", false, "Skip the error migration and continue with the execution (still exits with an error);;s")
	return c
}

// HandleUp executes pending migrations
func HandleUp(o UpOption) error {
	r, cl, e := legacyRuntime()
	if e != nil {
		return fmt.Errorf("failed to connect to database: %v", e)
	}
	defer cl()
	return RunUp(r, o, true)
}
