package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gookit/goutil/x/ccolor"
)

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
//  eg: 20251105-102430-add-age-index.sql
//  => {Time: time.Time{}, Date: "20251105-102430", Name: "add-age-index"}
type FilenameInfo struct {
	Time time.Time // parsed time from Date field
	Date string // eg: 20251105-102430
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
