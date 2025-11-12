package command

import (
	"fmt"
	"strings"

	"github.com/gookit/goutil/arrutil"
	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/pkg/database"
)

// DB alias for database.DB
type DB = database.DB

// ShowOption represents options for the show command
type ShowOption struct {
	// Show database tables
	Tables bool
	// Show table schema
	Schema string
}

// NewShowCommand shows database information like tables or table schema
func NewShowCommand() *capp.Cmd {
	var showOpt = ShowOption{}

	c := capp.NewCmd("show", "Show database information like tables or table schema", func(c *capp.Cmd) error {
		return HandleShow(showOpt)
	})

	c.Aliases = []string{"info", "describe"}
	bindCommonFlags(c)

	c.BoolVar(&showOpt.Tables, "tables", false, "Show database tables;;t")
	c.StringVar(&showOpt.Schema, "schema", "", "Show table schema;;s")

	return c
}

// HandleShow handles the show command logic
func HandleShow(opt ShowOption) error {
	// Validate options
	if !opt.Tables && opt.Schema == "" {
		return fmt.Errorf("either --tables or --schema must be provided")
	}

	if opt.Tables && opt.Schema != "" {
		return fmt.Errorf("--tables and --schema cannot be used together")
	}

	// Load configuration and connect to database
	_, db, err := initConfigAndDB()
	if err != nil {
		return err
	}
	defer db.Close()

	if opt.Tables {
		// Show database tables
		return showTables(db)
	}

	// Show table schema
	if opt.Schema != "" {
		return showTableSchema(db, opt.Schema)
	}
	return nil
}

// showTables displays all tables in the database
func showTables(db *DB) error {
	ccolor.Println("🔍  Fetching database tables...")

	query, err := getTablesQuery(db.Driver())
	if err != nil {
		return fmt.Errorf("unsupported database driver: %s", db.Driver())
	}

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %v", err)
		}
		tables = append(tables, tableName)
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating over table rows: %v", err)
	}

	if len(tables) == 0 {
		ccolor.Infoln("No tables found in the database.")
		return nil
	}

	tables = arrutil.Filter(tables, func(s string) bool {
		return s != database.MigrateSchemaName
	})

	ccolor.Printf("📋  Found <green>%d</> table(s):\n", len(tables))
	for i, table := range tables {
		ccolor.Printf("  %d. %s\n", i+1, table)
	}
	return nil
}

// showTableSchema displays the schema of a specific table
func showTableSchema(db *DB, tableName string) error {
	ccolor.Printf("🔍  Fetching schema for table: <green>%s</>\n", tableName)

	query, err := getSchemaQuery(db.Driver(), tableName)
	if err != nil {
		return fmt.Errorf("unsupported database driver: %s", db.Driver())
	}

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query table schema: %v", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		if db.Driver() == "postgres" {
			// For PostgreSQL, use different column order
			err := rows.Scan(&col.Name, &col.Type, &col.NotNull, &col.Default)
			if err != nil {
				return fmt.Errorf("failed to scan column info: %v", err)
			}
		} else {
			// For MySQL, SQLite, etc.
			err := rows.Scan(&col.Name, &col.Type, &col.NotNull, &col.Default, &col.Key, &col.Extra)
			if err != nil {
				return fmt.Errorf("failed to scan column info: %v", err)
			}
		}
		columns = append(columns, col)
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating over schema rows: %v", err)
	}

	if len(columns) == 0 {
		ccolor.Printf("No columns found for table: %s\n", tableName)
		return nil
	}

	ccolor.Printf("📋  Table <green>%s</> has <green>%d</> column(s):\n", tableName, len(columns))
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-20s %-20s %-10s %-15s %-10s %-15s\n", "Name", "Type", "Null", "Default", "Key", "Extra")
	fmt.Println(strings.Repeat("-", 80))
	for _, col := range columns {
		fmt.Printf("%-20s %-20s %-10s %-15s %-10s %-15s\n",
			col.Name, col.Type, col.NotNull, col.Default, col.Key, col.Extra)
	}
	fmt.Println(strings.Repeat("-", 80))

	return nil
}

// getTablesQuery returns the appropriate query for fetching table names based on database driver
func getTablesQuery(driver string) (string, error) {
	switch driver {
	case "mysql":
		return "SHOW TABLES", nil
	case "postgres":
		return `SELECT tablename FROM pg_tables WHERE schemaname = 'public'`, nil
	case "sqlite":
		return `SELECT name FROM sqlite_master WHERE type='table'`, nil
	case "mssql":
		return `SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE = 'BASE TABLE'`, nil
	default:
		return "", fmt.Errorf("unsupported driver: %s", driver)
	}
}

// getSchemaQuery returns the appropriate query for fetching table schema based on database driver
func getSchemaQuery(driver, tableName string) (string, error) {
	switch driver {
	case "mysql":
		return fmt.Sprintf("DESCRIBE `%s`", tableName), nil
	case "postgres":
		return fmt.Sprintf(`
			SELECT 
				column_name, 
				data_type, 
				is_nullable, 
				column_default 
			FROM information_schema.columns 
			WHERE table_name = '%s' 
			ORDER BY ordinal_position`, tableName), nil
	case "sqlite":
		return fmt.Sprintf("PRAGMA table_info(`%s`)", tableName), nil
	case "mssql":
		return fmt.Sprintf(`
			SELECT 
				COLUMN_NAME,
				DATA_TYPE,
				IS_NULLABLE,
				COLUMN_DEFAULT,
				COLUMNPROPERTY(OBJECT_ID(TABLE_SCHEMA+'.'+TABLE_NAME), COLUMN_NAME, 'IsIdentity') AS IS_IDENTITY,
				'' AS EXTRA
			FROM INFORMATION_SCHEMA.COLUMNS
			WHERE TABLE_NAME = '%s'
			ORDER BY ORDINAL_POSITION`, tableName), nil
	default:
		return "", fmt.Errorf("unsupported driver: %s", driver)
	}
}

// ColumnInfo represents information about a database column
type ColumnInfo struct {
	Name    string
	Type    string
	NotNull string
	Default string
	Key     string
	Extra   string
}
