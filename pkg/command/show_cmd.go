package command

import (
	"fmt"
	"strings"

	"github.com/gookit/goutil/arrutil"
	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/strutil"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/pkg/database"
)

// ShowOption represents options for the show command
type ShowOption struct {
	// Show database tables
	Tables bool
	// Show one table schema
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
	if err := initConfigAndDB(); err != nil {
		return err
	}
	defer db.SilentClose()

	// Show database tables
	if opt.Tables {
		return showTables(db)
	}

	// Show table schema
	if opt.Schema != "" {
		return showTableSchema(db, opt.Schema)
	}
	return nil
}

// showTables displays all tables in the database
func showTables(db *database.DB) error {
	ccolor.Println("🔍  Fetching database tables...")

	tables, err := db.ShowTables()
	if err != nil {
		return err
	}
	if len(tables) == 0 {
		ccolor.Infoln("No tables found in the database.")
		return nil
	}

	tables = arrutil.Filter(tables, func(s string) bool {
		return s != database.SchemaTableName
	})

	ccolor.Printf("📋  Found <green>%d</> table(s):\n", len(tables))
	for i, table := range tables {
		ccolor.Printf("  %d. %s\n", i+1, table)
	}
	return nil
}

// showTableSchema displays the schema of a specific table
func showTableSchema(db *database.DB, tableName string) error {
	ccolor.Printf("🔍  Fetching schema for table: <green>%s</>\n", tableName)
	columns, err := db.QueryTableSchema(tableName)
	if err != nil {
		return err
	}

	if len(columns) == 0 {
		ccolor.Warnf("No columns found for table: %s\n", tableName)
		return nil
	}

	hLine := strings.Repeat("-", 110)
	ccolor.Printf("📋  Table <green>%s</> has <green>%d</> column(s):\n", tableName, len(columns))
	fmt.Println(hLine)
	ccolor.Printf(" %-20s | %-30s | %-4s | %-20s | %-10s | %-15s\n", "Name", "Type", "Null", "Default", "Key", "Extra")
	fmt.Println(hLine)
	for _, col := range columns {
		defVal := strutil.OrCond(col.Default.Valid, col.Default.String, "NULL")
		fmt.Printf(" %-20s | %-30s | %-4s | %-20s | %-10s | %-15s\n",
			col.Name, col.Type, col.NotNull, defVal, col.Key, col.Extra,
		)
	}
	fmt.Println(hLine)

	return nil
}
