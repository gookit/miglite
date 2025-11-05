package migration

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Migration represents a single migration file
type Migration struct {
	FileName    string
	Content     string
	UpMigration string
	DownMigration string
	Timestamp   time.Time
	Version     string
}

// ParseMigration parses a migration file to extract UP and DOWN sections
func ParseMigration(filePath string) (*Migration, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	text := string(content)

	// Extract UP section
	upStart := strings.Index(text, "-- Migrate:UP --")
	if upStart == -1 {
		return nil, fmt.Errorf("migration file %s does not contain -- Migrate:UP -- marker", filePath)
	}
	
	upEnd := strings.Index(text[upStart:], "-- Migrate:DOWN --")
	var upContent string
	if upEnd == -1 {
		// No DOWN section, take everything after UP marker
		upContent = strings.TrimSpace(text[upStart+15:]) // 15 is len("-- Migrate:UP --")
	} else {
		// Take content between UP and DOWN markers
		upContent = strings.TrimSpace(text[upStart+15 : upStart+upEnd])
	}

	// Extract DOWN section if it exists
	var downContent string
	if upEnd != -1 {
		downStart := upStart + upEnd + 18 // 18 is len("-- Migrate:DOWN --")
		downContent = strings.TrimSpace(text[downStart:])
	}

	// Extract timestamp from filename
	fileName := filepath.Base(filePath)
	timestamp, err := extractTimestamp(fileName)
	if err != nil {
		return nil, err
	}

	return &Migration{
		FileName:    fileName,
		Content:     text,
		UpMigration: upContent,
		DownMigration: downContent,
		Timestamp:   timestamp,
		Version:     fileName, // Use filename as version
	}, nil
}

// extractTimestamp extracts the timestamp from a filename in YYYYMMDD format
func extractTimestamp(filename string) (time.Time, error) {
	// Match YYYYMMDD part from filename
	re := regexp.MustCompile(`^(\d{8})`)
	matches := re.FindStringSubmatch(filename)
	if len(matches) < 2 {
		return time.Time{}, fmt.Errorf("invalid migration filename format: %s, expected YYYYMMDD-...", filename)
	}

	dateStr := matches[1]
	timestamp, err := time.Parse("20060102", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date in filename: %s", filename)
	}

	return timestamp, nil
}

// FindMigrations finds all migration files in the specified directory
func FindMigrations(migrationsDir string) ([]*Migration, error) {
	var migrations []*Migration

	err := filepath.Walk(migrationsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Only process .sql files
		if strings.HasSuffix(strings.ToLower(info.Name()), ".sql") {
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

// DiscoverMigrations discovers all migration files, parses them, and returns them sorted by timestamp
func DiscoverMigrations(migrationsDir string) ([]*Migration, error) {
	migrations, err := FindMigrations(migrationsDir)
	if err != nil {
		return nil, err
	}

	// Sort migrations by timestamp
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Timestamp.Before(migrations[j].Timestamp)
	})

	return migrations, nil
}