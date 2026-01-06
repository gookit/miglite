package migration

import (
	"testing"

	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/testutil/assert"
)

var testContents = []byte(`
-- Migrate:UP --
ALTER TABLE users
    ADD COLUMN password_hash TEXT;

-- Migrate:DOWN --
-- ALTER TABLE users DROP COLUMN password_hash;

`)

func TestParseFile(t *testing.T) {
	sqlFile := "testdata/20251105-102430_add_password_hash.sql"
	_, err := fsutil.WriteData(sqlFile, testContents)
	assert.NoError(t, err)

	m, err := ParseFile(sqlFile)
	assert.NoError(t, err)

	assert.StrContains(t, m.FileName, "20251105-102430")
	assert.Eq(t, "ALTER TABLE users\n    ADD COLUMN password_hash TEXT;", m.UpSection)
	assert.Empty(t, m.DownSection)
}

func TestMigration_ParseContents(t *testing.T) {
	m := &Migration{}
	m.Contents = string(testContents)

	err := m.ParseContents()
	assert.NoErr(t, err)
	assert.Eq(t, "ALTER TABLE users\n    ADD COLUMN password_hash TEXT;", m.UpSection)
	assert.Empty(t, m.DownSection)

	m.ResetContents()
	assert.Empty(t, m.Contents)
	assert.Empty(t, m.UpSection)
}
