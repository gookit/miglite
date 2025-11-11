package migration

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/x/ccolor"
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
	timestamp := time.Now().Format(DateLayout)
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
-- created_at: %s
--

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

// FindMigrations finds all migration files in the specified directory, and returns them sorted by timestamp
func FindMigrations(migrationsDir string) ([]*Migration, error) {
	var migrations []*Migration
	ccolor.Printf("🔎  Discovering migrations from <green>%s</>\n", migrationsDir)

	err := filepath.Walk(migrationsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() { // TODO 支持date子目录
			return nil
		}

		// Only process .sql files
		if strings.HasSuffix(info.Name(), ".sql") {
			migration, err := NewMigration(path)
			if err != nil {
				return err
			}
			migrations = append(migrations, migration)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort migrations by timestamp
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Timestamp.Before(migrations[j].Timestamp)
	})

	return migrations, nil
}

// FilenameInfo represents the information extracted from a migration filename
//
//	eg: 20251105-102430-add-age-index.sql
//	=> {Time: time.Time{}, Date: "20251105-102430", Name: "add-age-index"}
type FilenameInfo struct {
	Time time.Time // parsed time from Date field
	Date string    // eg: 20251105-102430
	Name string
}

// parseFilename extracts the time,name from a migration filename
func parseFilename(filename string) (*FilenameInfo, error) {
	matches := regexFilename.FindStringSubmatch(filename)
	if len(matches) < 3 {
		return nil, fmt.Errorf("invalid filename format: %s, expected %s-{name}.sql", filename, PrefixFormat)
	}

	dateStr := matches[1]
	createTime, err := time.Parse(DateLayout, dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid datetime in filename: %s", filename)
	}

	return &FilenameInfo{
		Time: createTime,
		Date: dateStr,
		Name: matches[2],
	}, nil
}
