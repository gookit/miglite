package testdrv

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/gookit/goutil/x/assert"
	"github.com/gookit/miglite/internal/config"
	"github.com/gookit/miglite/internal/database"
	"github.com/gookit/miglite/internal/runtime"
	"github.com/gookit/miglite/pkg/migcom"
	_ "modernc.org/sqlite"
)

func TestRuntimeSQLiteLifecycle(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "runtime.db")
	sqlDB, err := sql.Open("sqlite", dbPath)
	assert.NoErr(t, err)
	defer sqlDB.Close()
	db := database.NewWithSqlDB(migcom.DriverSQLite, sqlDB)
	cfg := &config.Config{Database: config.Database{Driver: migcom.DriverSQLite, SqlDriver: "sqlite", DSN: dbPath}, Migrations: config.Migrations{Path: "../../../testdata/migrations/sqlite", Recursive: true}}
	r := runtime.NewWithDatabase(cfg, db, false)
	assert.NoErr(t, r.Init(runtime.InitOption{}))
	assert.NoErr(t, r.Up(runtime.UpOption{}))
	assert.NoErr(t, r.Close())
	rows, err := sqlDB.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='users'")
	assert.NoErr(t, err)
	defer rows.Close()
	assert.True(t, rows.Next())
}
