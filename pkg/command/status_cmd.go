package command

import (
	"github.com/gookit/goutil/cflag/capp"
)

type StatusOption struct{}

func StatusCommand() *capp.Cmd {
	c := capp.NewCmd("status", "Show the status of migrations", func(*capp.Cmd) error { return HandleStatus(StatusOption{}) })
	c.Aliases = []string{"st"}
	bindCommonFlags(c)
	return c
}

// HandleStatus shows the status of migrations
func HandleStatus(_ StatusOption) error {
	r, cleanup, err := legacyRuntime()
	if err != nil {
		return err
	}
	defer cleanup()
	return RunStatus(r)
}
