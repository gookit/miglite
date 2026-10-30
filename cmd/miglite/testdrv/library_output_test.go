package testdrv_test

import (
	"bytes"
	"database/sql"
	"embed"
	"fmt"
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

// TestLibraryOutputConcurrentMigrators runs two migrators with their own Output
// at the same time: ccolor's writer is process global, so the redirect is
// serialized and each run's progress text lands in its own buffer.
func TestLibraryOutputConcurrentMigrators(t *testing.T) {
	migrationFS, err := fs.Sub(libraryMigrations, "embedfixtures/mig_one")
	assert.Require(t, assert.NoErr(t, err))

	run := func(name string, output *bytes.Buffer) error {
		dbPath := filepath.Join(t.TempDir(), name+".db")
		sqlDB, err := sql.Open("sqlite", dbPath)
		if err != nil {
			return err
		}
		defer sqlDB.Close()
		cfg := &miglite.Config{
			Database:   miglite.Database{Driver: "sqlite", SqlDriver: "sqlite", DSN: dbPath},
			Migrations: miglite.Migrations{Path: "."},
			Output:     output,
		}
		migrator := miglite.NewWithConfigAndFS(cfg, migrationFS).SetSqlDB(sqlDB)
		if err := migrator.Init(command.InitOption{}); err != nil {
			return err
		}
		return migrator.Up(command.UpOption{Yes: true, Number: 1})
	}

	outputs := []*bytes.Buffer{new(bytes.Buffer), new(bytes.Buffer)}
	errs := make(chan error, len(outputs))
	for i, output := range outputs {
		go func(i int, output *bytes.Buffer) { errs <- run(fmt.Sprintf("db%d", i), output) }(i, output)
	}
	for range outputs {
		assert.NoErr(t, <-errs)
	}
	for _, output := range outputs {
		assert.StrContains(t, output.String(), "Migration schema initialized")
		assert.StrContains(t, output.String(), "create-embed-widgets.sql")
	}
}
