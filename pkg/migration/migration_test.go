package migration

import (
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

func TestMigration_ParseContents(t *testing.T) {
	m := &Migration{}
	contents := []byte(`
-- Migrate:UP --
ALTER TABLE users
    ADD COLUMN password_hash TEXT;

-- Migrate:DOWN --
-- ALTER TABLE users DROP COLUMN password_hash;

`)
	m.Contents = string(contents)

	err := m.ParseContents()
	assert.NoErr(t, err)
	assert.Eq(t, "ALTER TABLE users\n    ADD COLUMN password_hash TEXT;", m.UpSection)
	assert.Empty(t, m.DownSection)
}