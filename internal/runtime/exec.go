package runtime

import (
	"fmt"
	"os"
	"strings"
)

type ExecOption struct {
	SQLOrFile string
	Yes       bool
}

func (r *Runtime) Exec(opt ExecOption) error {
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
	statements := splitSQLStatements(sqlText)
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
		if _, err = tx.Exec(statement); err != nil {
			return fmt.Errorf("failed to execute SQL statement %d: %w", i+1, err)
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}
