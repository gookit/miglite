package migration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

func TestParseFilename(t *testing.T) {
	fi, err := parseFilename("20251105-102430-add-age-index.sql")
	assert.Nil(t, err)
	assert.NotNil(t, fi)
	assert.Eq(t, "20251105-102430", fi.Date)
	assert.Eq(t, "2025-11-05 10:24:30", fi.Time.Format("2006-01-02 15:04:05"))
	assert.Eq(t, "add-age-index", fi.Name)

	// invalid format
	fi, err = parseFilename("20251105-add-age-index")
	assert.Err(t, err)
	assert.Nil(t, fi)
}

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