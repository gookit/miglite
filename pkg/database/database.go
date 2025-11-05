package database

import (
	"database/sql"
	"fmt"
	"log"

	// Register database drivers
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// DB represents a database connection
type DB struct {
	*sql.DB
	driver string
}

// Connect establishes a database connection
func Connect(driver, dsn string) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{DB: db, driver: driver}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// GetDriver returns the database driver
func (db *DB) GetDriver() string {
	return db.driver
}

// InitSchema creates the migrations table if it doesn't exist
func (db *DB) InitSchema() error {
	var sqlStmt string
	switch db.driver {
	case "mysql":
		sqlStmt = `
CREATE TABLE IF NOT EXISTS db_schema_migrations (
    version VARCHAR(160) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(24) -- up,skip,down
);`
	case "postgres", "postgresql":
		sqlStmt = `
CREATE TABLE IF NOT EXISTS db_schema_migrations (
    version VARCHAR(160) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(24) -- up,skip,down
);`
	case "sqlite3":
		sqlStmt = `
CREATE TABLE IF NOT EXISTS db_schema_migrations (
    version VARCHAR(160) PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(24) -- up,skip,down
);`
	default:
		return fmt.Errorf("unsupported database driver: %s", db.driver)
	}

	_, err := db.Exec(sqlStmt)
	return err
}