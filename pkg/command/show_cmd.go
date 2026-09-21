package command

import (
	"github.com/gookit/goutil/cflag/capp"
)

type ShowOption struct {
	Tables bool
	Schema string
}

func NewShowCommand() *capp.Cmd {
	var opt ShowOption
	c := capp.NewCmd("show", "Show database information like tables or table schema", func(*capp.Cmd) error { return HandleShow(opt) })
	c.Aliases = []string{"info", "describe"}
	bindCommonFlags(c)
	c.BoolVar(&opt.Tables, "tables", false, "Show database tables;;t")
	c.StringVar(&opt.Schema, "schema", "", "Show table schema;;s")
	return c
}

// HandleShow shows database tables or a table schema
func HandleShow(opt ShowOption) error {
	// validate before connecting, so bad usage fails without a database connection
	if err := validateShowOption(opt); err != nil {
		return err
	}

	r, cleanup, err := legacyRuntime()
	if err != nil {
		return err
	}
	defer cleanup()
	return RunShow(r, opt)
}
