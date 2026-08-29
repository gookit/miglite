package command

import "github.com/gookit/miglite/internal/migutil"

func splitSQLStatements(s string) []string { return migutil.SplitSQL(s) }
func isQuerySQL(s string) bool             { return migutil.IsQuerySQL(s) }
