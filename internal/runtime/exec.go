package runtime

import (
	"fmt"
	"github.com/gookit/miglite/internal/migutil"
	"os"
	"strings"
)

type ExecOption struct {
	SQLOrFile string
	Yes       bool
}

func (r *Runtime) Exec(opt ExecOption) error {
	return r.ExecWithHooks(opt, ExecHooks{})
}

func (r *Runtime) ExecWithHooks(opt ExecOption, hooks ExecHooks) error {
	if err := r.ensureDB(); err != nil {
		return err
	}
	sqlText := strings.TrimSpace(opt.SQLOrFile)
	if sqlText == "" {
		return fmt.Errorf("either SQL or sql-file must be provided")
	}
	if len(sqlText) < 128 && strings.HasSuffix(sqlText, ".sql") {
		data, err := os.ReadFile(sqlText)
		if err != nil {
			return err
		}
		sqlText = string(data)
	}
	statements := migutil.SplitSQL(sqlText)
	if len(statements) == 0 {
		return fmt.Errorf("no SQL statements to execute")
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	for i, statement := range statements {
		if hooks.BeforeStatement != nil {
			if err = hooks.BeforeStatement(i, len(statements), statement); err != nil {
				return err
			}
		}
		result, execErr := tx.Exec(statement)
		if execErr != nil {
			err = execErr
			return fmt.Errorf("failed to execute SQL statement %d: %w", i+1, err)
		}
		if hooks.AfterStatement != nil {
			hooks.AfterStatement(i, len(statements), statement, result)
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}
