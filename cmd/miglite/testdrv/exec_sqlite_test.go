package testdrv

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/gookit/goutil/x/assert"
	"github.com/gookit/miglite/internal/database"
	"github.com/gookit/miglite/pkg/command"
	"github.com/gookit/miglite/pkg/migcom"
)

func TestExecMultiSQL(t *testing.T) {
	t.Run("commit all statements", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "commit.db")
		setCommandSQLiteDB(t, dbPath)

		err := command.HandleExec(command.ExecOption{
			SQLOrFile: `CREATE TABLE items(id INTEGER PRIMARY KEY, name TEXT);
				INSERT INTO items(name) VALUES ('first');
				SELECT id, name FROM items;
				UPDATE items SET name = 'updated' WHERE id = 1;`,
			Yes: true,
		})
		assert.NoErr(t, err)

		db, err := sql.Open("sqlite", dbPath)
		assert.Require(t, assert.NoErr(t, err))
		defer db.Close()

		var name string
		assert.NoErr(t, db.QueryRow("SELECT name FROM items WHERE id = 1").Scan(&name))
		assert.Eq(t, "updated", name)
	})

	t.Run("rollback all statements", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "rollback.db")
		setCommandSQLiteDB(t, dbPath)

		err := command.HandleExec(command.ExecOption{
			SQLOrFile: `CREATE TABLE items(id INTEGER PRIMARY KEY, name TEXT);
				INSERT INTO items(name) VALUES ('first');
				INSERT INTO missing_table(name) VALUES ('fail');`,
			Yes: true,
		})
		assert.Err(t, err)

		db, err := sql.Open("sqlite", dbPath)
		assert.Require(t, assert.NoErr(t, err))
		defer db.Close()

		var count int
		assert.Err(t, db.QueryRow("SELECT COUNT(*) FROM items").Scan(&count))
	})
}

func setCommandSQLiteDB(t *testing.T, dbPath string) {
	t.Helper()
	db, err := database.NewDB(migcom.DriverSQLite, "sqlite", dbPath)
	assert.Require(t, assert.NoErr(t, err))
	command.SetDB(db)
	t.Cleanup(func() { command.SetDB(nil) })
}
