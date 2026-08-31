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
	sqlText := strings.TrimSpace(opt.SQLOrFile)
	if sqlText == "" {
		return fmt.Errorf("either SQL or sql-file must be provided")
	}
	fromFile := false
	if len(sqlText) < 128 && strings.HasSuffix(sqlText, ".sql") {
		fromFile = true
		data, err := os.ReadFile(sqlText)
		if err != nil {
			return fmt.Errorf("failed to read SQL file: %v", err)
		}
		sqlText = string(data)
		if strings.TrimSpace(sqlText) == "" {
			return fmt.Errorf("no SQL contents in file: %s", opt.SQLOrFile)
		}
	}
	if hooks.Input != nil {
		hooks.Input(strings.TrimSpace(sqlText), fromFile)
	}
	if !opt.Yes && hooks.Confirm != nil {
		message := "Are you sure you want to execute the following SQL statement?"
		if fromFile {
			message = fmt.Sprintf("Are you sure you want to execute SQL from file: %s", opt.SQLOrFile)
		}
		if !hooks.Confirm(message) {
			return ErrCancelled
		}
	}
	statements := migutil.SplitSQL(sqlText)
	if len(statements) == 0 {
		return fmt.Errorf("no SQL statements to execute")
	}
	if err := r.ensureDB(); err != nil {
		return err
	}
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin SQL transaction: %v", err)
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
				return fmt.Errorf("failed to get query columns: %v", err)
			}
			for rows.Next() {
				vals := make([]any, len(qr.Columns))
				ptrs := make([]any, len(vals))
				for j := range vals {
					ptrs[j] = &vals[j]
				}
				if scanErr := rows.Scan(ptrs...); scanErr != nil {
					_ = rows.Close()
					return fmt.Errorf("failed to scan query row: %v", scanErr)
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
				return fmt.Errorf("failed to iterate query rows: %v", err)
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
		return fmt.Errorf("failed to commit SQL transaction: %v", err)
	}
	committed = true
	return nil
}
