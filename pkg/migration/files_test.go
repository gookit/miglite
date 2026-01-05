package migration

import (
	"os"
	"testing"

	"github.com/gookit/goutil/dump"
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
	parentDir := "./testdata/migtest"
	tempDir := "./testdata/migtest/subdir"
	_ = os.RemoveAll(parentDir)
	err := os.MkdirAll(tempDir, 0755)
	assert.NoError(t, err)

	// Test creating a migration
	name := "test-migration"
	filePath, err := CreateMigration(parentDir, name)
	assert.NoError(t, err)
	assert.FileExists(t, filePath)

	// Read the file and verify the content
	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.StrContainsAll(t, string(content), []string{MarkUp, MarkDown})

	// create migration in subdir
	filePath, err = CreateMigration(tempDir, "subdir-migration")
	assert.NoError(t, err)
	assert.FileExists(t, filePath)

	// 以下划线开头的目录会被忽略
	filePath, err = CreateMigration(parentDir+"/_ignoredir", "ignore-migration")
	assert.NoError(t, err)
	assert.FileExists(t, filePath)

	t.Run("find", func(t *testing.T) {
		migrations, err1 := FindMigrations(parentDir, true)
		assert.NoError(t, err1)
		assert.Eq(t, 2, len(migrations))

		migrations, err1 = FindMigrations(parentDir, false)
		assert.NoError(t, err1)
		assert.Eq(t, 1, len(migrations))
		dump.P(migrations)
	})
}
