package migutil

import (
	"github.com/gookit/goutil/x/assert"
	"testing"
)

func TestSplitSQL(t *testing.T) {
	got := SplitSQL("INSERT INTO t VALUES ('a;b'); SELECT 1;")
	assert.Eq(t, 2, len(got))
	assert.Contains(t, got[0], "a;b")
}
