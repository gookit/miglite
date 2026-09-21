package testdrv

import (
	"database/sql"
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/gookit/goutil/x/assert"
	"github.com/gookit/miglite"
	"github.com/gookit/miglite/internal/config"
	"github.com/gookit/miglite/pkg/command"
	"github.com/gookit/miglite/pkg/migcom"
	_ "modernc.org/sqlite"
)

//go:embed embedfixtures
var embedRawFS embed.FS

// embedMigFS exposes embedfixtures as the fs root, so the logical migration dirs
// (mig_one, mig_two) do not exist on disk relative to the working directory: a
// migration read would fail if any code fell back to the local filesystem.
func embedMigFS(t *testing.T) fs.FS {
	t.Helper()

	sub, err := fs.Sub(embedRawFS, "embedfixtures")
	assert.Require(t, assert.NoErr(t, err))
	return sub
}

// newEmbedMigrator prepares a migrator over an embedded migrations dir.
func newEmbedMigrator(t *testing.T) (*miglite.Migrator, *sql.DB, *config.Config) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "embed.db")
	sqlDB, err := sql.Open("sqlite", dbPath)
	assert.Require(t, assert.NoErr(t, err))
	t.Cleanup(func() { _ = sqlDB.Close() })

	cfg := &config.Config{
		Database:   config.Database{Driver: migcom.DriverSQLite, SqlDriver: "sqlite", DSN: dbPath},
		Migrations: config.Migrations{Path: "mig_one"},
	}
	m := miglite.NewWithConfigAndFS(cfg, embedMigFS(t)).SetSqlDB(sqlDB)
	return m, sqlDB, cfg
}

func hasEmbedTable(t *testing.T, sqlDB *sql.DB, table string) bool {
	t.Helper()

	var count int
	assert.Require(t, assert.NoErr(t, sqlDB.QueryRow(
		"SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)))
	return count > 0
}

func recordStatus(t *testing.T, sqlDB *sql.DB, version string) string {
	t.Helper()

	var status string
	err := sqlDB.QueryRow("SELECT status FROM z_schema_migrations WHERE version = ?", version).Scan(&status)
	if err == sql.ErrNoRows {
		return ""
	}
	assert.NoErr(t, err)
	return status
}

func TestMigratorEmbedFS(t *testing.T) {
	const upFile = "20260101-100000-create-embed-widgets.sql"
	const skipFile = "20260101-100100-embed-skipped.sql"

	m, sqlDB, cfg := newEmbedMigrator(t)
	assert.NoErr(t, m.Init(command.InitOption{}))

	// the logical dirs only exist inside the embedded fs
	_, statErr := os.Stat("mig_one")
	assert.Err(t, statErr, "fixture leaked to disk, the test would not prove fs reading")

	t.Run("up reads the embedded fs", func(t *testing.T) {
		assert.NoErr(t, m.Up(command.UpOption{Yes: true}))
		assert.True(t, hasEmbedTable(t, sqlDB, "embed_widgets"))
		assert.Eq(t, "up", recordStatus(t, sqlDB, upFile))
	})

	t.Run("status reads the embedded fs", func(t *testing.T) {
		assert.NoErr(t, m.Status(command.StatusOption{}))
	})

	t.Run("skip reads the embedded fs", func(t *testing.T) {
		// the skipped file lives in a second embedded dir, the skip op resolves
		// it through the multi dir path of MigrationsFromFS
		cfg.Migrations.Path = "mig_one,mig_two"
		defer func() { cfg.Migrations.Path = "mig_one" }()

		assert.NoErr(t, m.Skip(command.SkipOption{FileNames: []string{skipFile}}))
		assert.Eq(t, "skip", recordStatus(t, sqlDB, skipFile))
	})

	t.Run("down reads the embedded fs", func(t *testing.T) {
		assert.NoErr(t, m.Down(command.DownOption{Number: 1, Yes: true}))
		assert.False(t, hasEmbedTable(t, sqlDB, "embed_widgets"))
	})

	t.Run("set fs nil restores the disk mode", func(t *testing.T) {
		diskDir := t.TempDir()
		assert.NoErr(t, os.WriteFile(filepath.Join(diskDir, "20260101-100200-disk-only.sql"), []byte(
			"-- Migrate:UP\nCREATE TABLE disk_only (id INTEGER PRIMARY KEY);\n-- Migrate:DOWN\nDROP TABLE disk_only;\n"), 0o644))

		cfg.Migrations.Path = diskDir
		m.SetFS(nil)

		assert.NoErr(t, m.Up(command.UpOption{Yes: true}))
		assert.True(t, hasEmbedTable(t, sqlDB, "disk_only"))
		assert.Eq(t, "up", recordStatus(t, sqlDB, "20260101-100200-disk-only.sql"))
	})
}
