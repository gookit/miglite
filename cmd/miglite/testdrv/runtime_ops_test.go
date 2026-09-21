package testdrv

import (
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/gookit/goutil/x/assert"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/internal/config"
	"github.com/gookit/miglite/internal/database"
	"github.com/gookit/miglite/internal/runtime"
	"github.com/gookit/miglite/pkg/command"
	"github.com/gookit/miglite/pkg/migcom"
	"github.com/gookit/miglite/pkg/migration"
	_ "modernc.org/sqlite"
)

const sqlOneTable = `-- Migrate:UP
CREATE TABLE tbl_one (id INTEGER PRIMARY KEY);
-- Migrate:DOWN
DROP TABLE tbl_one;
`

const sqlTwoTable = `-- Migrate:UP
CREATE TABLE tbl_two (id INTEGER PRIMARY KEY);
`

const sqlTwoTableAgain = `-- Migrate:UP
CREATE TABLE tbl_two (id INTEGER PRIMARY KEY);
`

// newOpsRuntime builds a sqlite runtime over a temp migrations dir.
func newOpsRuntime(t *testing.T, files map[string]string) (*runtime.Runtime, *sql.DB) {
	t.Helper()

	dir := t.TempDir()
	migDir := filepath.Join(dir, "migrations")
	assert.Require(t, assert.NoErr(t, os.MkdirAll(migDir, 0o755)))
	for name, body := range files {
		assert.Require(t, assert.NoErr(t, os.WriteFile(filepath.Join(migDir, name), []byte(body), 0o644)))
	}

	dbPath := filepath.Join(dir, "ops.db")
	sqlDB, err := sql.Open("sqlite", dbPath)
	assert.Require(t, assert.NoErr(t, err))
	t.Cleanup(func() { _ = sqlDB.Close() })

	db := database.NewWithSqlDB(migcom.DriverSQLite, sqlDB)
	cfg := &config.Config{
		Database:   config.Database{Driver: migcom.DriverSQLite, SqlDriver: "sqlite", DSN: dbPath},
		Migrations: config.Migrations{Path: migDir, Recursive: true},
	}
	return runtime.NewWithDatabase(cfg, db, false), sqlDB
}

func hasTable(t *testing.T, sqlDB *sql.DB, table string) bool {
	t.Helper()

	var count int
	assert.Require(t, assert.NoErr(t, sqlDB.QueryRow(
		"SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)))
	return count > 0
}

// captureOutput redirects ccolor output while fn runs.
func captureOutput(fn func()) string {
	buf := new(bytes.Buffer)
	ccolor.SetOutput(buf)
	defer ccolor.SetOutput(os.Stdout)
	fn()
	return buf.String()
}

// A migration file without a DOWN section must not be reported as rolled back.
func TestRuntimeDownEmptyDownIsNotRolledBack(t *testing.T) {
	r, sqlDB := newOpsRuntime(t, map[string]string{
		"20260101-100000-one.sql": sqlOneTable,
		"20260101-100100-two.sql": sqlTwoTable,
	})
	assert.NoErr(t, r.Up(runtime.UpOption{}))

	var rolledBack, emptyDown []string
	err := r.DownWithHooks(runtime.DownOption{Number: 2}, runtime.MigrationHooks{
		Skip: func(_, _ int, mig *migration.Migration, status string) {
			if status == runtime.StatusEmptyDown {
				emptyDown = append(emptyDown, mig.FileName)
			}
		},
		After: func(_, _ int, mig *migration.Migration) { rolledBack = append(rolledBack, mig.FileName) },
	})
	assert.NoErr(t, err)
	assert.Eq(t, []string{"20260101-100100-two.sql"}, emptyDown)
	assert.Eq(t, []string{"20260101-100000-one.sql"}, rolledBack)
	assert.True(t, hasTable(t, sqlDB, "tbl_two"))
	assert.False(t, hasTable(t, sqlDB, "tbl_one"))
}

// With SkipErr the run continues after a failed file, but the failure is still
// reported to the caller, so the CLI cannot exit with success.
func TestRuntimeUpSkipErrContinuesAndReportsFailure(t *testing.T) {
	r, sqlDB := newOpsRuntime(t, map[string]string{
		"20260101-100000-one.sql": sqlOneTable,
		"20260101-100100-dup.sql": sqlTwoTable,
		"20260101-100200-dup.sql": sqlTwoTableAgain,
	})

	var result runtime.MigrationResult
	err := r.UpWithHooks(runtime.UpOption{SkipErr: true}, runtime.MigrationHooks{
		Complete: func(res runtime.MigrationResult) { result = res },
	})
	assert.ErrMsgContains(t, err, "20260101-100200-dup.sql")
	assert.Eq(t, 3, result.Total)
	assert.Eq(t, 2, result.Applied)
	assert.Eq(t, 1, result.Failed)
	assert.True(t, hasTable(t, sqlDB, "tbl_one"))
}

// Without SkipErr a failed file aborts the run and no completion is reported.
func TestRuntimeUpAbortsOnFailure(t *testing.T) {
	r, sqlDB := newOpsRuntime(t, map[string]string{
		"20260101-100000-dup.sql":  sqlTwoTable,
		"20260101-100100-dup.sql":  sqlTwoTableAgain,
		"20260101-100200-late.sql": sqlOneTable,
	})

	completed := false
	err := r.UpWithHooks(runtime.UpOption{}, runtime.MigrationHooks{
		Complete: func(runtime.MigrationResult) { completed = true },
	})
	assert.Err(t, err)
	assert.False(t, completed)
	assert.False(t, hasTable(t, sqlDB, "tbl_one"))
}

// An empty migrations dir reports the found total and never reports completion.
func TestRuntimeUpNoMigrations(t *testing.T) {
	r, _ := newOpsRuntime(t, nil)

	var totals []int
	completed := false
	assert.NoErr(t, r.UpWithHooks(runtime.UpOption{}, runtime.MigrationHooks{
		Start:    func(total int) { totals = append(totals, total) },
		Complete: func(runtime.MigrationResult) { completed = true },
	}))
	assert.Eq(t, []int{0}, totals)
	assert.False(t, completed)
}

// The shared command runner must expose the SkipErr failure to the CLI.
func TestRunUpSkipErrReturnsError(t *testing.T) {
	r, _ := newOpsRuntime(t, map[string]string{
		"20260101-100000-one.sql": sqlOneTable,
		"20260101-100100-dup.sql": sqlOneTable,
	})
	out := captureOutput(func() {
		err := command.RunUp(r, command.UpOption{Yes: true, SkipErr: true}, true)
		assert.Err(t, err)
	})
	assert.StrContains(t, out, "migration(s) failed")
}

// The down runner must not claim a rollback for a file without a DOWN section.
func TestRunDownEmptyDownOutput(t *testing.T) {
	r, _ := newOpsRuntime(t, map[string]string{
		"20260101-100000-one.sql": sqlOneTable,
		"20260101-100100-two.sql": sqlTwoTable,
	})
	assert.NoErr(t, r.Up(runtime.UpOption{}))

	out := captureOutput(func() {
		assert.NoErr(t, command.RunDown(r, command.DownOption{Number: 2, Yes: true}, true))
	})
	assert.StrContains(t, out, "Skipping empty DOWN migration!")
	assert.StrContains(t, out, "Success rolled back migration: 20260101-100000-one.sql")
	assert.StrNotContains(t, out, "Success rolled back migration: 20260101-100100-two.sql")
	assert.StrContains(t, out, "Successfully rolled back 1 migration(s)")
}

// The status and show runners print the v0.6.0 table formats.
func TestRunStatusAndShowOutput(t *testing.T) {
	r, _ := newOpsRuntime(t, map[string]string{
		"20260101-100000-one.sql": sqlOneTable,
	})
	assert.NoErr(t, r.Up(runtime.UpOption{}))

	statusOut := captureOutput(func() {
		assert.NoErr(t, command.RunStatus(r))
	})
	assert.StrContains(t, statusOut, "Migrations Status:(total=1)")
	assert.StrContains(t, statusOut, "20260101-100000-one.sql")

	showOut := captureOutput(func() {
		assert.NoErr(t, command.RunShow(r, command.ShowOption{Tables: true}))
	})
	assert.StrContains(t, showOut, "Fetching database tables...")
	assert.StrContains(t, showOut, "tbl_one")
	assert.StrNotContains(t, showOut, database.SchemaTableName)

	showErr := command.RunShow(r, command.ShowOption{})
	assert.ErrMsgContains(t, showErr, "either --tables or --schema must be provided")
}
