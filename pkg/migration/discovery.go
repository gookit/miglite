package migration

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DiscoverMigrations finds and parses all migration files in the specified directory
func DiscoverMigrations(migrationsPath string) ([]*Migration, error) {
	var migrations []*Migration

	// Walk through the migrations directory
	err := filepath.Walk(migrationsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process .sql files
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".sql") {
			migration, err := ParseMigration(path)
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