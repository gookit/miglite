package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CreateMigration creates a new migration file with the specified name
func CreateMigration(migrationsDir, name string) (string, error) {
	// Generate filename with current timestamp
	timestamp := time.Now().Format("20060102") // YYYYMMDD format
	filename := fmt.Sprintf("%s-%s.sql", timestamp, name)

	// Ensure the migrations directory exists
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create migrations directory: %v", err)
	}

	// Full path for the new migration file
	filePath := filepath.Join(migrationsDir, filename)
	
	// Create the migration template
	content := fmt.Sprintf(`%s
-- Add your migration SQL here

%s
-- Add your rollback SQL here (optional)
`, MarkUp, MarkDown)

	// Write the content to the file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write migration file: %v", err)
	}
	
	return filePath, nil
}