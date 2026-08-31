package runtime

import (
	"fmt"
)

func (r *Runtime) Show(opt ShowOption) (ShowResult, error) {
	if err := r.ensureDB(); err != nil {
		return ShowResult{}, err
	}
	if !opt.Tables && opt.Schema == "" {
		return ShowResult{}, fmt.Errorf("either --tables or --schema must be provided")
	}
	if opt.Tables && opt.Schema != "" {
		return ShowResult{}, fmt.Errorf("--tables and --schema cannot be used together")
	}
	if opt.Tables {
		tables, err := r.db.ShowTables()
		return ShowResult{Tables: tables}, err
	}
	columns, err := r.db.QueryTableSchema(opt.Schema)
	return ShowResult{Columns: columns}, err
}
