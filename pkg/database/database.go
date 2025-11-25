package database

import (
	"database/sql"
	"fmt"

	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/goutil/x/stdio"
)

// DB represents a database connection
type DB struct {
	*sql.DB
	dsn string
	// more information about the database
	debug  bool
	driver string // formatted driver name. eg migcom.DriverMySQL
	// provider
	provider SqlProvider
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
//	DSN:
//	 - mysql: username:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
//	 - postgres: host=localhost port=5432 user=username password=password dbname=dbname sslmode=disable
//	 - sqlite: filepath
func Connect(driver, sqlDriver, dsn string) (*DB, error) {
	if std != nil && std.dsn == dsn {
		return std, nil
	}

	// get the register driver name
	db, err := sql.Open(sqlDriver, dsn)
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

// Driver returns the formatted driver name
func (db *DB) Driver() string { return db.driver }

// Close closes the database connection
func (db *DB) Close() error { return db.DB.Close() }

// SilentClose closes the database connection
func (db *DB) SilentClose() {
	if err := db.DB.Close(); err != nil {
		ccolor.Errorln("[ERROR] database.Close:", err)
	}
}

// SetDebug sets the debug mode
func (db *DB) SetDebug(debug bool) { db.debug = debug }

// SqlProvider for database driver
func (db *DB) SqlProvider() (SqlProvider, error) {
	var err error
	if db.provider == nil {
		db.provider, err = GetSqlProvider(db.driver)
	}
	return db.provider, err
}

// InitSchema creates the migrations table if it doesn't exist
func (db *DB) InitSchema() error {
	provide, err := db.SqlProvider()
	if err != nil {
		return err
	}

	var sqlStmt = provide.CreateSchema()
	if db.debug {
		fmt.Println("[DEBUG] database.InitSchema:", sqlStmt)
	}
	_, err = db.Exec(sqlStmt)
	return err
}

// DropSchema drops the migrations table
func (db *DB) DropSchema() error {
	provide, err := db.SqlProvider()
	if err != nil {
		return err
	}

	var sqlStmt = provide.DropSchema()
	if db.debug {
		fmt.Println("[DEBUG] database.DropSchema:", sqlStmt)
	}
	_, err = db.Exec(sqlStmt)
	return err
}

// ShowTables displays all tables in the database
func (db *DB) ShowTables() ([]string, error) {
	provide, err := db.SqlProvider()
	if err != nil {
		return nil, fmt.Errorf("unsupported database driver: %s", db.Driver())
	}

	rows, err := db.Query(provide.ShowTables())
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %v", err)
	}
	defer stdio.SafeClose(rows)

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %v", err)
		}
		tables = append(tables, tableName)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over table rows: %v", err)
	}
	return tables, nil
}

// ColumnInfo represents information about a database column
type ColumnInfo struct {
	Name    string
	Type    string
	NotNull string
	Default sql.NullString
	Key     string
	Extra   string
}

// QueryTableSchema queries the schema of a specific table
func (db *DB) QueryTableSchema(tableName string) ([]ColumnInfo, error) {
	provide, err := db.SqlProvider()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(provide.QueryTableSchema(tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to query table schema: %v", err)
	}
	defer stdio.SafeClose(rows)

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		if db.Driver() == "postgres" {
			// For PostgreSQL, use different column order
			err = rows.Scan(&col.Name, &col.Type, &col.NotNull, &col.Default)
			if err != nil {
				return nil, fmt.Errorf("failed to scan column info: %v", err)
			}
		} else {
			// For MySQL, SQLite, etc.
			err = rows.Scan(&col.Name, &col.Type, &col.NotNull, &col.Default, &col.Key, &col.Extra)
			if err != nil {
				return nil, fmt.Errorf("failed to scan column info: %v", err)
			}
		}
		columns = append(columns, col)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over schema rows: %v", err)
	}
	return columns, nil
}