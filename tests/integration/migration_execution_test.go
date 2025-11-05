package integration

import (
	"os"
	"testing"

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
	if err != nil {
		// If config file doesn't exist, create a basic config
		cfg = &config.Config{
			Database: struct {
				Driver string `yaml:"driver"`
				DSN    string `yaml:"dsn"`
			}{
				Driver: "sqlite3",
				DSN:    dbPath,
			},
			Migrations: struct {
				Path string `yaml:"path"`
			}{
				Path: "./testdata",
			},
		}
	}

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
	migrations, err := migration.DiscoverMigrations(cfg.Migrations.Path)
	if err != nil {
		t.Fatalf("Failed to discover migrations: %v", err)
	}

	// Execute migrations
	executor := migration.NewExecutor(db)
	for _, mig := range migrations {
		applied, status, err := migration.IsMigrationApplied(db, mig.Version)
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
	statuses, err := migration.GetMigrationStatus(db, migrations)
	if err != nil {
		t.Fatalf("Failed to get migration status: %v", err)
	}

	for _, status := range statuses {
		if status.Status != "up" {
			t.Errorf("Migration %s has status %s, expected 'up'", status.Version, status.Status)
		}
	}
}