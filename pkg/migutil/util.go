package migutil

import (
	"strings"
)

// FmtDriverName format the database driver name to standard
func FmtDriverName(driver string) string {
	driver = strings.ToLower(driver)
	switch driver {
	case "mysql", "my", "mariadb", "mysql2":
		return "mysql"
	case "postgres", "pg", "pgx", "pgsql", "postgresql":
		return "postgres"
	case "sqlite", "sqlite3":
		return "sqlite"
	case "mssql", "ms", "sqlserver":
		return "mssql"
	default:
		return driver
	}
}
