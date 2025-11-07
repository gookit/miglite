package migration

import (
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

func TestParseFilename(t *testing.T) {
	fi, err := parseFilename("20251105-102430-add-age-index.sql")
	assert.Nil(t, err)
	assert.NotNil(t, fi)
	assert.Eq(t, "20251105-102430", fi.Date)
	assert.Eq(t, "2025-11-05 10:24:30", fi.Time.Format("2006-01-02 15:04:05"))
	assert.Eq(t, "add-age-index", fi.Name)

	// invalid format
	fi, err = parseFilename("20251105-add-age-index")
	assert.Err(t, err)
	assert.Nil(t, fi)
}
