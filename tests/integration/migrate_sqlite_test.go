package integration

import (
	"os"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/miglite/pkg/config"
	"github.com/gookit/miglite/pkg/database"
	"github.com/gookit/miglite/pkg/migration"
)

func TestMigration_sqlite(t *testing.T) {
	// For this test, we'll use a SQLite database
	dbPath := "./test_migration.db"

	// Set up environment variables for test
	err := os.Setenv("DATABASE_URL", "sqlite://"+dbPath)
	assert.NoError(t, err)
	defer os.Remove(dbPath) // Clean up after the test

	// run command: init
	err = app.RunWithArgs([]string{"init"})
	assert.Nil(t, err)

	// Discover migrations
	migrations, err := migration.FindMigrations(config.Get().Migrations.Path)
	assert.NoError(t, err)

	// Execute migrations
	executor := migration.NewExecutor(database.GetDB(), true)
	for _, mig := range migrations {
		applied, status, err := migration.IsApplied(database.GetDB(), mig.FileName)
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

	// Verify migrations were applied
	statuses, err := migration.GetMigrationsStatus(database.GetDB(), migrations)
	if err != nil {
		t.Fatalf("Failed to get migration status: %v", err)
	}

	for _, status := range statuses {
		if status.Status != "up" {
			t.Errorf("Migration %s has status %s, expected 'up'", status.Version, status.Status)
		}
	}
}
