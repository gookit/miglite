package migration

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/gookit/goutil/fsutil"
)

// CreateMigrations creates multi migration file with the specified names
func CreateMigrations(migrationsDir string, names []string) ([]string, error) {
	var files []string
	for _, name := range names {
		if name == "" || name[0] == '-' {
			return nil, fmt.Errorf("invalid migration name: %s", name)
		}

		filePath, err := CreateMigration(migrationsDir, name)
		if err != nil {
			return nil, err
		}
		files = append(files, filePath)
	}
	return files, nil
}

// CreateMigration creates a new migration file with the specified name
func CreateMigration(migrationsDir, name string) (string, error) {
	// Generate filename with current timestamp. format: YYYYMMDD-HHMMSS
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s.sql", timestamp, name)

	// Full path for the new migration file
	filePath := filepath.Join(migrationsDir, filename)
	if fsutil.IsFile(filePath) {
		return "", fmt.Errorf("migration file already exists: %s", filePath)
	}

	var userLine string
	u, err := user.Current()
	if err == nil {
		userLine = fmt.Sprintf("\n-- author: %s", filepath.Base(u.Username))
	}

	// Create the migration template
	content := fmt.Sprintf(`--
-- name: %s%s
-- created at: %s

%s
-- Add your migration SQL here

%s
-- Add your rollback SQL here (optional)
`, name, userLine, timestamp, MarkUp, MarkDown)

	// Ensure the migrations directory exists
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create migrations directory: %v", err)
	}

	// Write the content to the file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write migration file: %v", err)
	}
	
	return filePath, nil
}