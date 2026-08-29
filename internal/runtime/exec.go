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
		if migutil.IsQuerySQL(statement) {
			rows, queryErr := tx.Query(statement)
			if queryErr != nil {
				return fmt.Errorf("failed to execute SQL statement %d: %w", i+1, queryErr)
			}
			qr := QueryResult{}
			qr.Columns, err = rows.Columns()
			if err != nil {
				_ = rows.Close()
				return err
			}
			for rows.Next() {
				vals := make([]any, len(qr.Columns))
				ptrs := make([]any, len(vals))
				for j := range vals {
					ptrs[j] = &vals[j]
				}
				if scanErr := rows.Scan(ptrs...); scanErr != nil {
					_ = rows.Close()
					return scanErr
				}
				for j, v := range vals {
					if b, ok := v.([]byte); ok {
						vals[j] = string(b)
					}
				}
				qr.Rows = append(qr.Rows, vals)
			}
			if err = rows.Err(); err != nil {
				_ = rows.Close()
				return err
			}
			_ = rows.Close()
			if hooks.QueryResult != nil {
				hooks.QueryResult(i, len(statements), statement, qr)
			}
			continue
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
