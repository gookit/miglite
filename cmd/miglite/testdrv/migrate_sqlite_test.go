package testdrv

import (
	"path/filepath"
	"testing"

	"github.com/gookit/goutil/x/assert"
	"github.com/gookit/miglite/internal/database"
	"github.com/gookit/miglite/pkg/migcom"
	"github.com/gookit/miglite/pkg/migration"
)

func TestMigration_sqlite(t *testing.T) {
	db, err := database.NewDB(migcom.DriverSQLite, "sqlite", filepath.Join(t.TempDir(), "migration.db"))
	assert.Require(t, assert.NoErr(t, err))
	defer db.SilentClose()
	assert.Require(t, assert.NoErr(t, db.InitSchema()))

	// Discover migrations
	migrations, err := migration.FindMigrations("../../../testdata/migrations/sqlite", true)
	assert.Require(t, assert.NoErr(t, err))

	// Execute migrations
	executor := migration.NewExecutor(db, true)
	for _, mig := range migrations {
		applied, status, err := migration.IsApplied(db, mig.FileName)
		if err != nil {
			t.Fatalf("Failed to check migration status: %v", err)
		}

		if !applied || status == "down" {
			t.Logf("Executing migration: %s", mig.FileName)
			if err := executor.ExecuteUp(mig); err != nil {
				t.Fatalf("Failed to execute migration %s: %v", mig.FileName, err)
			}
		} else {
			t.Logf("Skipping already applied migration: %s", mig.FileName)
		}
	}
}
