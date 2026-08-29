package runtime

import (
	"database/sql"
	"github.com/gookit/miglite/pkg/migration"
)

type MigrationHooks struct {
	Before  func(index, total int, mig *migration.Migration) error
	After   func(index, total int, mig *migration.Migration)
	Skip    func(index, total int, mig *migration.Migration, status string)
	Error   func(index, total int, mig *migration.Migration, err error)
	Confirm func(message string) bool
}

type ExecHooks struct {
	BeforeStatement func(index, total int, statement string) error
	AfterStatement  func(index, total int, statement string, result sql.Result)
	Confirm         func(message string) bool
}

type InitOption struct{ Drop bool }
type StatusOption struct{}
type ShowOption struct {
	Tables bool
	Schema string
}
