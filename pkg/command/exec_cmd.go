package command

import (
	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/cliutil"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/internal/runtime"
	"strings"
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
func HandleExec(o ExecOption) error {
	r, cl, e := legacyRuntime()
	if e != nil {
		return e
	}
	defer cl()
	ccolor.Printf("📄 Input SQL: %s\n", o.SQLOrFile)
	if !o.Yes && !cliutil.Confirm("Continue?") {
		return nil
	}
	return r.ExecWithHooks(runtime.ExecOption{SQLOrFile: o.SQLOrFile, Yes: true}, runtime.ExecHooks{BeforeStatement: func(i, total int, _ string) error {
		ccolor.Printf("🚀 Executing SQL statement %d/%d...\n", i+1, total)
		return nil
	}, QueryResult: func(_, _ int, _ string, qr runtime.QueryResult) {
		ccolor.Printf("📘 Query Results(size=%d):\n", len(qr.Rows))
		ccolor.Printf("  %s\n", strings.Join(qr.Columns, "  | "))
		for _, row := range qr.Rows {
			ccolor.Printf("  %v\n", row)
		}
	}})
}
