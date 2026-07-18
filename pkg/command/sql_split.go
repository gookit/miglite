package command

import "strings"

const (
	sqlNormal byte = iota
	sqlSingleQuote
	sqlDoubleQuote
	sqlBacktick
	sqlLineComment
	sqlBlockComment
)

// splitSQLStatements splits common SQL without breaking quoted or commented semicolons.
func splitSQLStatements(sqlText string) []string {
	var statements []string
	var buf strings.Builder
	state := sqlNormal

	flush := func() {
		if statement := strings.TrimSpace(buf.String()); statement != "" {
			statements = append(statements, statement)
		}
		buf.Reset()
	}

	for i := 0; i < len(sqlText); i++ {
		ch := sqlText[i]
		next := byte(0)
		if i+1 < len(sqlText) {
			next = sqlText[i+1]
		}

		if state == sqlNormal && ch == ';' {
			flush()
			continue
		}

		buf.WriteByte(ch)
		switch state {
		case sqlNormal:
			switch {
			case ch == '\'':
				state = sqlSingleQuote
			case ch == '"':
				state = sqlDoubleQuote
			case ch == '`':
				state = sqlBacktick
			case ch == '-' && next == '-', ch == '/' && next == '*':
				buf.WriteByte(next)
				i++
				if ch == '-' {
					state = sqlLineComment
				} else {
					state = sqlBlockComment
				}
			case ch == '#':
				state = sqlLineComment
			}
		case sqlSingleQuote, sqlDoubleQuote, sqlBacktick:
			quote := byte('\'')
			if state == sqlDoubleQuote {
				quote = '"'
			} else if state == sqlBacktick {
				quote = '`'
			}
			if ch == '\\' && next != 0 {
				buf.WriteByte(next)
				i++
			} else if ch == quote && next == quote {
				buf.WriteByte(next)
				i++
			} else if ch == quote {
				state = sqlNormal
			}
		case sqlLineComment:
			if ch == '\n' {
				state = sqlNormal
			}
		case sqlBlockComment:
			if ch == '*' && next == '/' {
				buf.WriteByte(next)
				i++
				state = sqlNormal
			}
		}
	}

	flush()
	return statements
}

func isQuerySQL(sqlText string) bool {
	sqlText = strings.ToLower(trimLeadingSQLComments(sqlText))
	for _, keyword := range []string{"select", "describe", "pragma", "show"} {
		if strings.HasPrefix(sqlText, keyword) &&
			(len(sqlText) == len(keyword) || !isSQLWordChar(sqlText[len(keyword)])) {
			return true
		}
	}
	return false
}

func trimLeadingSQLComments(sqlText string) string {
	for {
		sqlText = strings.TrimSpace(sqlText)
		switch {
		case strings.HasPrefix(sqlText, "--"), strings.HasPrefix(sqlText, "#"):
			if end := strings.IndexByte(sqlText, '\n'); end >= 0 {
				sqlText = sqlText[end+1:]
				continue
			}
			return ""
		case strings.HasPrefix(sqlText, "/*"):
			if end := strings.Index(sqlText[2:], "*/"); end >= 0 {
				sqlText = sqlText[end+4:]
				continue
			}
			return ""
		default:
			return sqlText
		}
	}
}

func isSQLWordChar(ch byte) bool {
	return ch == '_' || ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9'
}
