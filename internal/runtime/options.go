package runtime

import (
	"database/sql"
	"errors"
	"github.com/gookit/miglite/pkg/migration"
)

type MigrationHooks struct {
	Start    func(total int)
	Complete func(result MigrationResult)
	Before   func(index, total int, mig *migration.Migration) error
	After    func(index, total int, mig *migration.Migration)
	Skip     func(index, total int, mig *migration.Migration, status string)
	Error    func(index, total int, mig *migration.Migration, err error)
	Confirm  func(message string) bool
}

var ErrCancelled = errors.New("operation cancelled")

type MigrationResult struct{ Total, Applied, Skipped, Failed int }

type ExecHooks struct {
	BeforeStatement func(index, total int, statement string) error
	AfterStatement  func(index, total int, statement string, result sql.Result)
	QueryResult     func(index, total int, statement string, result QueryResult)
	Confirm         func(message string) bool
}

type QueryResult struct {
	Columns []string
	Rows    [][]any
}

type InitOption struct{ Drop bool }
type StatusOption struct{}
type ShowOption struct {
	Tables bool
	Schema string
}
