package command

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/cliutil"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/internal/runtime"
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
	input := strings.TrimSpace(o.SQLOrFile)
	if input == "" {
		return fmt.Errorf("either SQL or sql-file must be provided")
	}
	r, cl, e := legacyRuntime()
	if e != nil {
		return e
	}
	defer cl()
	err := r.ExecWithHooks(runtime.ExecOption{SQLOrFile: o.SQLOrFile, Yes: o.Yes}, runtime.ExecHooks{
		Input: func(sqlText string, _ bool) {
			ccolor.Infop("📄  Input SQL: ")
			ccolor.Println(sqlText)
		},
		BeforeStatement: func(i, total int, _ string) error {
			ccolor.Printf("🚀 Executing SQL statement %d/%d...\n", i+1, total)
			return nil
		},
		AfterStatement: func(_, _ int, _ string, result sql.Result) {
			rows, err := result.RowsAffected()
			if err != nil {
				ccolor.Printf("✅  SQL executed successfully (result info not available)\n")
				return
			}
			ccolor.Printf("✅  SQL executed successfully, rows affected: <green>%d</>\n", rows)
		},
		Confirm: func(message string) bool {
			ccolor.Warnf("⚠️  %s\n", message)
			ok := cliutil.Confirm("Continue?")
			if !ok {
				ccolor.Magentaln("Exiting SQL execution!")
			}
			return ok
		},
		QueryResult: func(_, _ int, _ string, qr runtime.QueryResult) {
			ccolor.Printf("📘 Query Results(size=%d):\n", len(qr.Rows))
			ccolor.Printf("  %s\n", strings.Join(qr.Columns, "  | "))
			ccolor.Println("  ----------------------------------------------")
			for _, row := range qr.Rows {
				values := make([]string, len(row))
				for i, value := range row {
					values[i] = fmt.Sprint(value)
				}
				ccolor.Printf("  %s\n", strings.Join(values, "  | "))
			}
		},
	})
	if errors.Is(err, runtime.ErrCancelled) {
		return nil
	}
	return err
}
