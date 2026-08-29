package runtime

import (
	"fmt"
)

func (r *Runtime) Show(opt ShowOption) (any, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}
	if !opt.Tables && opt.Schema == "" {
		return nil, fmt.Errorf("either --tables or --schema must be provided")
	}
	if opt.Tables && opt.Schema != "" {
		return nil, fmt.Errorf("--tables and --schema cannot be used together")
	}
	if opt.Tables {
		return r.db.ShowTables()
	}
	return r.db.QueryTableSchema(opt.Schema)
}
