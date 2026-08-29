package migutil

import (
	"testing"

	"github.com/gookit/goutil/x/assert"
)

func TestSplitSQL(t *testing.T) {
	got := SplitSQL("INSERT INTO t VALUES ('a;b'); SELECT 1;")
	assert.Eq(t, 2, len(got))
	assert.Contains(t, got[0], "a;b")
}

func TestSplitSQL2(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want []string
	}{
		{
			name: "two statements",
			sql:  "CREATE TABLE users(id int); INSERT INTO users VALUES (1);",
			want: []string{"CREATE TABLE users(id int)", "INSERT INTO users VALUES (1)"},
		},
		{
			name: "quoted semicolons",
			sql:  "INSERT INTO logs VALUES ('a;b', \"c;d\", `e;f`); SELECT 1;",
			want: []string{"INSERT INTO logs VALUES ('a;b', \"c;d\", `e;f`)", "SELECT 1"},
		},
		{
			name: "comment semicolons",
			sql:  "-- keep ; here\nSELECT 1; /* keep ; here */ SELECT 2;",
			want: []string{"-- keep ; here\nSELECT 1", "/* keep ; here */ SELECT 2"},
		},
		{name: "empty statements", sql: "; SELECT 1;;", want: []string{"SELECT 1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Eq(t, tt.want, SplitSQL(tt.sql))
		})
	}
}

func TestIsQuerySQL(t *testing.T) {
	tests := map[string]bool{
		" SELECT 1":                  true,
		"-- comment\nDESCRIBE users": true,
		"/* comment */ PRAGMA info":  true,
		"# comment\nSHOW TABLES":     true,
		"SELECTED value":             false,
		"UPDATE users SET id = 1":    false,
	}

	for sqlText, want := range tests {
		t.Run(sqlText, func(t *testing.T) {
			assert.Eq(t, want, IsQuerySQL(sqlText))
		})
	}
}
