package miglite_test

import (
	"bytes"
	"database/sql"
	"embed"
	"io"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/gookit/goutil/x/assert"
	"github.com/gookit/miglite"
	"github.com/gookit/miglite/pkg/command"
	_ "modernc.org/sqlite"
)

// The test uses the same embedded migration shape as an external application.
//
//go:embed testdata/migrations/sqlite/*.sql
var embeddedMigrations embed.FS

func TestExternalConfigAndEmbeddedSQLite(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "embedded.db")
	sqlDB, err := sql.Open("sqlite", dbPath)
	assert.Require(t, assert.NoErr(t, err))
	t.Cleanup(func() { assert.NoErr(t, sqlDB.Close()) })

	migrationFS, err := fs.Sub(embeddedMigrations, "testdata/migrations/sqlite")
	assert.Require(t, assert.NoErr(t, err))

	output := new(bytes.Buffer)
	cfg := &miglite.Config{
		Database:   miglite.Database{Driver: "sqlite", SqlDriver: "sqlite", DSN: dbPath},
		Migrations: miglite.Migrations{Path: "."},
		Output:     output,
	}
	migrator := miglite.NewWithConfigAndFS(cfg, migrationFS).SetSqlDB(sqlDB)

	assert.NoErr(t, migrator.Init(command.InitOption{}))
	assert.NoErr(t, migrator.Up(command.UpOption{Yes: true, Number: 1}))
	assert.True(t, output.Len() > 0)

	outputLen := output.Len()
	cfg.Output = io.Discard
	assert.NoErr(t, migrator.Status(command.StatusOption{}))
	assert.Eq(t, outputLen, output.Len())

	var tableName string
	assert.NoErr(t, sqlDB.QueryRow("SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'users'").Scan(&tableName))
	assert.Eq(t, "users", tableName)
}
