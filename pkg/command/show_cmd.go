package command

import (
	"fmt"
	"github.com/gookit/goutil/arrutil"
	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/strutil"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/internal/database"
	"github.com/gookit/miglite/internal/runtime"
	"strings"
)

type ShowOption struct {
	Tables bool
	Schema string
}

func NewShowCommand() *capp.Cmd {
	var opt ShowOption
	c := capp.NewCmd("show", "Show database information like tables or table schema", func(*capp.Cmd) error { return HandleShow(opt) })
	c.Aliases = []string{"info", "describe"}
	bindCommonFlags(c)
	c.BoolVar(&opt.Tables, "tables", false, "Show database tables;;t")
	c.StringVar(&opt.Schema, "schema", "", "Show table schema;;s")
	return c
}

func HandleShow(opt ShowOption) error {
	if !opt.Tables && opt.Schema == "" {
		return fmt.Errorf("either --tables or --schema must be provided")
	}
	if opt.Tables && opt.Schema != "" {
		return fmt.Errorf("--tables and --schema cannot be used together")
	}
	r, cleanup, err := legacyRuntime()
	if err != nil {
		return err
	}
	defer cleanup()
	result, err := r.Show(runtime.ShowOption{Tables: opt.Tables, Schema: opt.Schema})
	if err != nil {
		return err
	}
	if opt.Tables {
		return showTablesList(result.Tables)
	}
	return showTableSchemaColumns(result.Columns, opt.Schema)
}

func showTablesList(tables []string) error {
	ccolor.Println("🔍  Fetching database tables...")
	tables = arrutil.Filter(tables, func(s string) bool { return s != database.SchemaTableName })
	if len(tables) == 0 {
		ccolor.Infoln("No tables found in the database.")
		return nil
	}
	ccolor.Printf("📋  Found <green>%d</> table(s):\n", len(tables))
	for i, table := range tables {
		ccolor.Printf("  %d. %s\n", i+1, table)
	}
	return nil
}

func showTableSchemaColumns(columns []database.ColumnInfo, table string) error {
	ccolor.Printf("🔍  Fetching schema for table: <green>%s</>\n", table)
	if len(columns) == 0 {
		ccolor.Warnf("No columns found for table: %s\n", table)
		return nil
	}
	hLine := strings.Repeat("-", 110)
	ccolor.Printf("📋  Table <green>%s</> has <green>%d</> column(s):\n", table, len(columns))
	fmt.Println(hLine)
	ccolor.Printf(" %-20s | %-30s | %-4s | %-20s | %-10s | %-15s\n", "Name", "Type", "Null", "Default", "Key", "Extra")
	fmt.Println(hLine)
	for _, col := range columns {
		defVal := strutil.OrCond(col.Default.Valid, col.Default.String, "NULL")
		fmt.Printf(" %-20s | %-30s | %-4s | %-20s | %-10s | %-15s\n", col.Name, col.Type, col.NotNull, defVal, col.Key, col.Extra)
	}
	fmt.Println(hLine)
	return nil
}
