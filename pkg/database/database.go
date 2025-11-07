package database

import (
	"database/sql"
	"fmt"
)

const (
	DriverMySQL    = "mysql"
	DriverPostgres = "postgres"
	DriverSQLite   = "sqlite"
)

// DB represents a database connection
type DB struct {
	*sql.DB
	debug bool
	driver string
}

// Connect establishes a database connection
//
//  driver: mysql, postgres, sqlite
//  dsn:
//   - mysql: username:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
//   - postgres: host=localhost port=5432 user=username password=password dbname=dbname sslmode=disable
//   - sqlite: filepath
func Connect(driver, dsn string) (*DB, error) {
	// get the register driver name
	regName := sqlDriver(driver)
	db, err := sql.Open(regName, dsn)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err1 := db.Ping(); err1 != nil {
		return nil, err1
	}
	return &DB{DB: db, driver: driver}, nil
}

func sqlDriver(driver string) string {
	if driver == DriverSQLite {
		// returns a list of supported SQL drivers
		registered := sql.Drivers()
		for _, name := range registered {
			if name == "sqlite3" { // for github.com/mattn/go-sqlite3
				driver = "sqlite3"
				break
			}
		}
	}
	return driver
}

// SetDebug sets the debug mode
func (db *DB) SetDebug(debug bool) { db.debug = debug }

// Driver returns the database driver name
func (db *DB) Driver() string { return db.driver }

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// InitSchema creates the migrations table if it doesn't exist
func (db *DB) InitSchema() error {
	var sqlStmt string
	switch db.driver {
	case DriverMySQL:
		sqlStmt = `
CREATE TABLE IF NOT EXISTS db_schema_migrations (
    version VARCHAR(160) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(24) -- up,skip,down
);`
	case DriverPostgres:
		sqlStmt = `
CREATE TABLE IF NOT EXISTS db_schema_migrations (
    version VARCHAR(160) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(24) -- up,skip,down
);`
	case DriverSQLite:
		sqlStmt = `
CREATE TABLE IF NOT EXISTS db_schema_migrations (
    version VARCHAR(160) PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(24) -- up,skip,down
);`
	default:
		return fmt.Errorf("unsupported database driver: %s", db.driver)
	}

	if db.debug {
		fmt.Println("[DEBUG] database.InitSchema:", sqlStmt)
	}
	_, err := db.Exec(sqlStmt)
	return err
}
