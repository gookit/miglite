package database

import (
	"database/sql"
	"fmt"
)

// supported database drivers
const (
	DriverMySQL    = "mysql"
	DriverMSSQL    = "mssql"
	DriverPostgres = "postgres"
	DriverSQLite   = "sqlite"
)

// DB represents a database connection
type DB struct {
	*sql.DB
	dsn string
	// more information about the database
	debug bool
	driver string
}

var std *DB

// GetDB returns the default database connection
func GetDB() *DB {
	if std == nil {
		panic("database not initialized")
	}
	return std
}

// Close closes the default database connection
func Close() error { return GetDB().Close() }

// Connect establishes a database connection
//
//	driver: mysql, postgres, sqlite
//	dsn:
//	 - mysql: username:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
//	 - postgres: host=localhost port=5432 user=username password=password dbname=dbname sslmode=disable
//	 - sqlite: filepath
func Connect(driver, dsn string) (*DB, error) {
	if std != nil && std.dsn == dsn {
		return std, nil
	}

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

	// db.SetMaxOpenConns(1) TODO support options
	std = &DB{DB: db, driver: driver, dsn: dsn}
	return std, nil
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

// Driver returns the database driver name
func (db *DB) Driver() string { return db.driver }

// Close closes the database connection
func (db *DB) Close() error { return db.DB.Close() }

// SetDebug sets the debug mode
func (db *DB) SetDebug(debug bool) { db.debug = debug }

// SqlBuilder for database driver
func (db *DB) SqlBuilder() (SqlBuilder, error) {
	return GetSqlBuilder(db.driver)
}

// InitSchema creates the migrations table if it doesn't exist
func (db *DB) InitSchema() error {
	b, err := GetSqlBuilder(db.driver)
	if err != nil {
		return err
	}

	var sqlStmt = b.CreateSchema()
	if db.debug {
		fmt.Println("[DEBUG] database.InitSchema:", sqlStmt)
	}
	_, err = db.Exec(sqlStmt)
	return err
}

// DropSchema drops the migrations table
func (db *DB) DropSchema() error {
	b, err := GetSqlBuilder(db.driver)
	if err != nil {
		return err
	}

	var sqlStmt = b.DropSchema()
	if db.debug {
		fmt.Println("[DEBUG] database.DropSchema:", sqlStmt)
	}
	_, err = db.Exec(sqlStmt)
	return err
}
