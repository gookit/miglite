package migutil

import (
	"strings"

	"github.com/gookit/miglite/pkg/migcom"
)

// FmtDriverName format the database driver name to standard
func FmtDriverName(driver string) string {
	driver = strings.ToLower(driver)
	switch driver {
	case "mysql", "my", "mariadb", "mysql2":
		return migcom.DriverMySQL
	case "postgres", "pg", "pgx", "pgsql", "postgresql":
		return migcom.DriverPostgres
	case "sqlite", "sqlite3":
		return migcom.DriverSQLite
	case "mssql", "ms", "sqlserver":
		return migcom.DriverMSSQL
	default:
		return driver
	}
}

// IsTableNotExists check the error message is table not exists
func IsTableNotExists(driver, errMsg string) bool {
	switch driver {
	case migcom.DriverMySQL:
		// Error 1146 (42S02): Table 'dbname.schema_migrations' doesn't exist
		return strings.Contains(errMsg, "1146") && strings.Contains(errMsg, "doesn't exist")
	case migcom.DriverPostgres:
		return strings.Contains(errMsg, "does not exist")
	case migcom.DriverSQLite:
		// sqlite: "no such table: xxx"
		return strings.Contains(errMsg, "no such table")
	case migcom.DriverMSSQL:
		return strings.Contains(errMsg, "does not exist")
	default:
		return false
	}
}
