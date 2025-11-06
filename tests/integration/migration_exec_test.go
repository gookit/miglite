package integration

import (
	"os"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/miglite/pkg/config"
	"github.com/gookit/miglite/pkg/database"
	"github.com/gookit/miglite/pkg/migration"
)

func TestMigrationExecution(t *testing.T) {
	// For this test, we'll use a SQLite database
	dbPath := "./test_migration.db"

	// Set up environment variables for test
	os.Setenv("DATABASE_URL", "sqlite://"+dbPath)
	defer os.Remove(dbPath) // Clean up after the test

	// Load configuration
	cfg, err := config.Load("./miglite_test.yaml") // This might not exist, which is okay for this test
	assert.NoError(t, err)

	// Connect to database
	db, err := database.Connect(cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize schema
	if err := db.InitSchema(); err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}

	// Discover migrations
	migrations, err := migration.FindMigrations(cfg.Migrations.Path)
	assert.NoError(t, err)

	// Execute migrations
	executor := migration.NewExecutor(db)
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

	// Verify migrations were applied
	statuses, err := migration.GetMigrationsStatus(db, migrations)
	if err != nil {
		t.Fatalf("Failed to get migration status: %v", err)
	}

	for _, status := range statuses {
		if status.Status != "up" {
			t.Errorf("Migration %s has status %s, expected 'up'", status.Version, status.Status)
		}
	}
}
