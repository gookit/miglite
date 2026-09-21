package migration

import (
	"testing"
	"testing/fstest"

	"github.com/gookit/goutil/x/assert"
)

const fsUpDownOne = `-- Migrate:UP
CREATE TABLE one (id INTEGER PRIMARY KEY);
-- Migrate:DOWN
DROP TABLE one;
`

const fsUpDownTwo = `-- Migrate:UP
CREATE TABLE two (id INTEGER PRIMARY KEY);
-- Migrate:DOWN
DROP TABLE two;
`

const fsUpThree = `-- Migrate:UP
CREATE TABLE three (id INTEGER PRIMARY KEY);
-- Migrate:DOWN
DROP TABLE three;
`

// embedMigFS holds the fixture tree shared by the fs.FS tests
func embedMigFS() fstest.MapFS {
	return fstest.MapFS{
		"migrations/20260101-100000-one.sql":         {Data: []byte(fsUpDownOne)},
		"migrations/20260101-100100-two.sql":         {Data: []byte(fsUpDownTwo)},
		"migrations/readme.md":                       {Data: []byte("not a migration")},
		"migrations/_ignored.sql":                    {Data: []byte(fsUpDownTwo)},
		"migrations/_backup/20251231-090000-old.sql": {Data: []byte(fsUpDownTwo)},
		"migrations/sub/20260102-100000-sub.sql":     {Data: []byte(fsUpThree)},
		"other/20260103-100000-other.sql":            {Data: []byte(fsUpThree)},
	}
}

func TestParseFS(t *testing.T) {
	fsys := embedMigFS()

	t.Run("parse up and down sections", func(t *testing.T) {
		mig, err := ParseFS(fsys, "migrations/20260101-100000-one.sql")
		assert.Require(t, assert.NoErr(t, err))
		assert.Eq(t, "20260101-100000-one.sql", mig.FileName)
		assert.Eq(t, "20260101-100000-one.sql", mig.Version)
		assert.Eq(t, "20260101-100000", mig.SortKey)
		assert.StrContains(t, mig.UpSection, "CREATE TABLE one")
		assert.StrContains(t, mig.DownSection, "DROP TABLE one")
	})

	t.Run("keep fs logical path", func(t *testing.T) {
		mig, err := ParseFS(fsys, "migrations/sub/20260102-100000-sub.sql")
		assert.Require(t, assert.NoErr(t, err))
		assert.Eq(t, "migrations/sub/20260102-100000-sub.sql", mig.FilePath)
		assert.Eq(t, "20260102-100000-sub.sql", mig.FileName)
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := ParseFS(fsys, "migrations/20260101-100000-nope.sql")
		assert.ErrMsgContains(t, err, "failed to read migration file")
	})

	t.Run("invalid file name", func(t *testing.T) {
		_, err := ParseFS(fsys, "migrations/readme.md")
		assert.Err(t, err)
	})

	t.Run("missing up section", func(t *testing.T) {
		fsys := fstest.MapFS{"migrations/20260101-100000-bad.sql": {Data: []byte("-- nothing here\n")}}
		_, err := ParseFS(fsys, "migrations/20260101-100000-bad.sql")
		assert.ErrMsgContains(t, err, "'-- Migrate:UP' section")
	})
}

func TestFindMigrationsFS(t *testing.T) {
	fsys := embedMigFS()

	t.Run("not recursive only scans the dir", func(t *testing.T) {
		list, err := FindMigrationsFS(fsys, "migrations", false)
		assert.Require(t, assert.NoErr(t, err))
		assert.Eq(t, 2, len(list))
		assert.Eq(t, "20260101-100000-one.sql", list[0].Version)
		assert.Eq(t, "20260101-100100-two.sql", list[1].Version)
		assert.Eq(t, "migrations/20260101-100000-one.sql", list[0].FilePath)
	})

	t.Run("recursive scans sub dirs and sorts", func(t *testing.T) {
		list, err := FindMigrationsFS(fsys, "migrations", true)
		assert.Require(t, assert.NoErr(t, err))
		assert.Eq(t, 3, len(list))
		assert.Eq(t, "20260101-100000-one.sql", list[0].Version)
		assert.Eq(t, "20260101-100100-two.sql", list[1].Version)
		assert.Eq(t, "20260102-100000-sub.sql", list[2].Version)
	})

	t.Run("ignore underscore prefixed and non sql files", func(t *testing.T) {
		list, err := FindMigrationsFS(fsys, "migrations", true)
		assert.Require(t, assert.NoErr(t, err))
		for _, mig := range list {
			assert.False(t, mig.FileName[0] == '_', "ignored file leaked: %s", mig.FilePath)
			assert.False(t, mig.FilePath == "migrations/_ignored.sql", "ignored file leaked")
		}
	})

	t.Run("multiple dirs", func(t *testing.T) {
		list, err := FindMigrationsFS(fsys, "migrations,other", true)
		assert.Require(t, assert.NoErr(t, err))
		assert.Eq(t, 4, len(list))
		assert.Eq(t, "20260103-100000-other.sql", list[3].Version)
	})

	t.Run("missing dir", func(t *testing.T) {
		_, err := FindMigrationsFS(fsys, "nope", false)
		assert.Err(t, err)
	})
}

func TestMigrationsFromFS(t *testing.T) {
	fsys := embedMigFS()

	t.Run("append sql ext and multiple dirs", func(t *testing.T) {
		list, err := MigrationsFromFS(fsys, "migrations,other", []string{
			"20260101-100000-one",
			"20260103-100000-other.sql",
		})
		assert.Require(t, assert.NoErr(t, err))
		assert.Eq(t, 2, len(list))
		assert.Eq(t, "migrations/20260101-100000-one.sql", list[0].FilePath)
		assert.Eq(t, "other/20260103-100000-other.sql", list[1].FilePath)
	})

	t.Run("file not exists", func(t *testing.T) {
		_, err := MigrationsFromFS(fsys, "migrations", []string{"nope"})
		assert.ErrMsgContains(t, err, "migration file not exists: nope.sql")
	})

	t.Run("skipped file must be a valid migration", func(t *testing.T) {
		_, err := MigrationsFromFS(fsys, "migrations", []string{"_ignored"})
		assert.Err(t, err)
	})
}
