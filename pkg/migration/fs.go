package migration

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/gookit/goutil/x/ccolor"
)

// This file provides the explicit fs.FS based API, so migrations can be embedded
// with Go's embed.FS. All paths are io/fs logical paths: slash separated and
// relative to the fs.FS root, eg: migrations/20260829-120000-create-users.sql.
//
// The disk API (ParseFile, FindMigrations, MigrationsFrom) is unchanged. A parsed
// Migration keeps the logical path in FilePath; no extra source field is added to
// the Migration struct.

// ParseFS parses one migration file from the given fs.FS.
func ParseFS(fsys fs.FS, filePath string) (*Migration, error) {
	mig, err := newMigration(path.Base(filePath), filePath)
	if err != nil {
		return nil, err
	}

	if err = mig.LoadFS(fsys); err != nil {
		return nil, err
	}
	return mig, nil
}

// LoadFS reads the migration contents from fsys and parses the UP/DOWN sections,
// keeping the current FilePath. It is the fs.FS counterpart of Parse, for the
// case where the file name is already resolved (eg. discovered migrations).
//
// NOTE: Migration does not remember its source, the caller decides between
// LoadFS and Parse.
func (m *Migration) LoadFS(fsys fs.FS) error {
	contents, err := fs.ReadFile(fsys, m.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %s", err)
	}

	m.Contents = string(contents)
	return m.ParseContents()
}

// FindMigrationsFS finds migration files in the given fs.FS directory, sorted by
// filename prefix.
//
//   - migrationsDir: allow multiple directories separated by comma
//   - recursive: walk the sub directories, ignoring "_" prefixed files/directories
func FindMigrationsFS(fsys fs.FS, migrationsDir string, recursive bool) ([]*Migration, error) {
	var migrations []*Migration
	ccolor.Printf("🔎  Discovering migrations from <green>%s</>\n", migrationsDir)

	for _, dirPath := range splitMigrationDirs(migrationsDir) {
		migList, err := findMigrationsFS(fsys, dirPath, recursive)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migList...)
	}

	// Sort migrations by timestamp
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].IsBefore(migrations[j])
	})

	return migrations, nil
}

// MigrationsFromFS creates migrations from a list of file names and fs.FS dirs.
//
// It is the fs.FS counterpart of MigrationsFrom, used by the skip command.
func MigrationsFromFS(fsys fs.FS, migPath string, files []string) ([]*Migration, error) {
	migrations := make([]*Migration, 0, len(files))
	migPaths := splitMigrationDirs(migPath)

	for _, file := range files {
		// check and auto append .sql
		if !strings.HasSuffix(file, ".sql") {
			file += ".sql"
		}

		var filePath string
		for _, dirPath := range migPaths {
			candidate := path.Join(dirPath, file)
			if _, err := fs.Stat(fsys, candidate); err == nil {
				filePath = candidate
				break
			}
		}

		if filePath == "" {
			return nil, fmt.Errorf("migration file not exists: %s", file)
		}

		mig, err := ParseFS(fsys, filePath)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, mig)
	}

	return migrations, nil
}

func findMigrationsFS(fsys fs.FS, dirPath string, recursive bool) ([]*Migration, error) {
	var migrations []*Migration

	// not recursive: only the .sql files in the current dir
	if !recursive {
		entries, err := fs.ReadDir(fsys, dirPath)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() || !isMigrationSQLFile(entry.Name()) {
				continue
			}

			mig, err := ParseFS(fsys, path.Join(dirPath, entry.Name()))
			if err != nil {
				return nil, err
			}
			migrations = append(migrations, mig)
		}
		return migrations, nil
	}

	// recursive: ".. /_backup/xx.sql" like "_" prefixed files/dirs are ignored
	err := fs.WalkDir(fsys, dirPath, func(filePath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filePath != dirPath && strings.HasPrefix(path.Base(filePath), "_") {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if entry.IsDir() || !isMigrationSQLFile(entry.Name()) {
			return nil
		}

		mig, err := ParseFS(fsys, filePath)
		if err != nil {
			return err
		}
		migrations = append(migrations, mig)
		return nil
	})

	return migrations, err
}

// isMigrationSQLFile reports whether the file name is a migration SQL file
func isMigrationSQLFile(fileName string) bool {
	return fileName != "" && fileName[0] != '_' && strings.HasSuffix(fileName, ".sql")
}

func splitMigrationDirs(dirPaths string) []string {
	return strings.Split(dirPaths, ",")
}
