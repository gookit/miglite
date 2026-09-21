package command

import (
	"github.com/gookit/goutil/cflag/capp"
)

type InitOption struct {
	Drop bool
}

// InitCommand initializes the migration schema on db
func InitCommand() *capp.Cmd {
	var initOpt = InitOption{}
	c := capp.NewCmd("init", "Initialize the migration schema on database")

	bindCommonFlags(c)
	c.BoolVar(&initOpt.Drop, "drop", false, "Drop existing schema before create")

	c.Func = func(c *capp.Cmd) error {
		return HandleInit(initOpt)
	}
	return c
}

// HandleInit initializes the migration schema on db
func HandleInit(opt InitOption) error {
	r, cleanup, err := legacyRuntime()
	if err != nil {
		return err
	}
	defer cleanup()
	return RunInit(r, opt)
}
