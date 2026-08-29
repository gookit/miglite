package runtime

import (
	"github.com/gookit/goutil/x/assert"
	"testing"
)

func TestSplitSQLStatementsKeepsQuotedSemicolons(t *testing.T) {
	got := splitSQLStatements("INSERT INTO t VALUES ('a;b'); SELECT 1;")
	assert.Eq(t, 2, len(got))
	assert.Contains(t, got[0], "a;b")
}
