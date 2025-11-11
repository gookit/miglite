package migutil

import (
	"fmt"
	"strings"
)

// ResolveDriver resolves the database driver name
func ResolveDriver(driver string) (string, error) {
	driver = strings.ToLower(driver)
	switch driver {
	case "mysql", "mariadb", "mysql2":
		return "mysql", nil
	case "postgres", "pg", "pgx", "pgsql", "postgresql":
		return "postgres", nil
	case "sqlite", "sqlite3":
		return "sqlite", nil
	case "mssql", "sqlserver":
		return "mssql", nil
	default:
		return "", fmt.Errorf("unsupported database driver: %s", driver)
	}
}
