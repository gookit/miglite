package command

import (
	"fmt"
	"strings"

	"github.com/gookit/goutil/cflag/capp"
)

type ExecOption struct {
	SQLOrFile string
	Yes       bool
}

func NewExecCommand() *capp.Cmd {
	var o ExecOption
	c := capp.NewCmd("exec", "Execute SQL statement or SQL file directly", func(c *capp.Cmd) error { o.SQLOrFile = c.Arg("sql-or-file").String(); return HandleExec(o) })
	c.Aliases = []string{"execute", "run-sql"}
	bindCommonFlags(c)
	c.BoolVar(&o.Yes, "yes", false, "Skip confirmation prompt;;y")
	c.AddArg("sql-or-file", "SQL statement/file to execute", true, nil)
	return c
}

// HandleExec executes SQL statements or a SQL file
func HandleExec(o ExecOption) error {
	if strings.TrimSpace(o.SQLOrFile) == "" {
		return fmt.Errorf("either SQL or sql-file must be provided")
	}

	r, cl, e := legacyRuntime()
	if e != nil {
		return e
	}
	defer cl()
	return RunExec(r, o, true)
}
