package migration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateMigration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := "./test_migrations"
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	// Test creating a migration
	name := "test-migration"
	filePath, err := CreateMigration(tempDir, name)
	if err != nil {
		t.Fatalf("Failed to create migration: %v", err)
	}

	// Verify that the file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Migration file was not created: %s", filePath)
	}

	// Read the file and verify the content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}

	expectedContent := "-- Migrate:UP --\n-- Add your migration SQL here\n\n\n-- Migrate:DOWN --\n-- Add your rollback SQL here (optional)\n"
	if string(content) != expectedContent {
		t.Errorf("Migration file content is incorrect.\nExpected: %s\nGot: %s", expectedContent, string(content))
	}

	// Verify the filename format
	expectedPrefix := "20" // All valid dates start with "20" in this century
	if len(filepath.Base(filePath)) < 10 || filepath.Base(filePath)[:2] != expectedPrefix {
		t.Errorf("Migration filename does not follow YYYYMMDD format: %s", filePath)
	}
}