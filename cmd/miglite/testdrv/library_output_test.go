package testdrv_test

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

// The migrations are embedded the way an application embeds its own. The test
// lives here, next to the other driver tests, so the miglite module itself does
// not depend on a database driver.
//
//go:embed embedfixtures/mig_one/*.sql
var libraryMigrations embed.FS

// TestLibraryConfigAndOutput uses only the exported API, as an application
// would: it builds the config, runs embedded SQLite migrations, and checks that
// Output receives the progress text and that io.Discard silences it.
func TestLibraryConfigAndOutput(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "embedded.db")
	sqlDB, err := sql.Open("sqlite", dbPath)
	assert.Require(t, assert.NoErr(t, err))
	t.Cleanup(func() { assert.NoErr(t, sqlDB.Close()) })

	migrationFS, err := fs.Sub(libraryMigrations, "embedfixtures/mig_one")
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
	assert.NoErr(t, sqlDB.QueryRow("SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'embed_widgets'").Scan(&tableName))
	assert.Eq(t, "embed_widgets", tableName)
}
