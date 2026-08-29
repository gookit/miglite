package migutil

import "strings"

// SplitSQL splits statements while preserving quoted and commented semicolons.
func SplitSQL(input string) []string {
	var out []string
	start, quote := 0, byte(0)
	lineComment, blockComment := false, false
	for i := 0; i < len(input); i++ {
		c := input[i]
		if lineComment {
			if c == '\n' {
				lineComment = false
			}
			continue
		}
		if blockComment {
			if c == '*' && i+1 < len(input) && input[i+1] == '/' {
				blockComment = false
				i++
			}
			continue
		}
		if quote != 0 {
			if c == '\\' {
				i++
			} else if c == quote {
				if i+1 < len(input) && input[i+1] == quote {
					i++
				} else {
					quote = 0
				}
			}
			continue
		}
		switch c {
		case '\'', '"', '`':
			quote = c
		case '#':
			lineComment = true
		case '-':
			if i+1 < len(input) && input[i+1] == '-' {
				lineComment = true
				i++
			}
		case '/':
			if i+1 < len(input) && input[i+1] == '*' {
				blockComment = true
				i++
			}
		case ';':
			if s := strings.TrimSpace(input[start:i]); s != "" {
				out = append(out, s)
			}
			start = i + 1
		}
	}
	if s := strings.TrimSpace(input[start:]); s != "" {
		out = append(out, s)
	}
	return out
}
