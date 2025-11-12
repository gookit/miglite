package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/cliutil"
	"github.com/gookit/goutil/x/ccolor"
)

// ExecOption represents options for the exec command
type ExecOption struct {
	// SQL statement to execute
	SQL string
	// Path to SQL file to execute
	File string
	// Skip confirmation prompt
	Yes bool
}

// NewExecCommand executes SQL statement or SQL file directly
func NewExecCommand() *capp.Cmd {
	var execOpt = ExecOption{}

	c := capp.NewCmd("exec", "Execute SQL statement or SQL file directly", func(c *capp.Cmd) error {
		return HandleExec(execOpt)
	})

	c.Aliases = []string{"execute", "run-sql"}
	bindCommonFlags(c)

	c.StringVar(&execOpt.SQL, "sql", "", "SQL statement to execute;;s")
	c.StringVar(&execOpt.File, "file", "", "Path to SQL file to execute;;f")
	c.BoolVar(&execOpt.Yes, "yes", false, "Skip confirmation prompt;;y")

	return c
}

// HandleExec handles the exec command logic
func HandleExec(opt ExecOption) error {
	// Validate options
	if opt.SQL == "" && opt.File == "" {
		return fmt.Errorf("either --sql or --file must be provided")
	}

	if opt.SQL != "" && opt.File != "" {
		return fmt.Errorf("--sql and --file cannot be used together")
	}

	// Load configuration and connect to database
	_, db, err := initConfigAndDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Prepare SQL to execute
	var sql string
	if opt.File != "" {
		// Read SQL from file
		var err error
		sql, err = readSQLFromFile(opt.File)
		if err != nil {
			return fmt.Errorf("failed to read SQL file: %v", err)
		}
	} else {
		// Use SQL from command line
		sql = opt.SQL
	}

	// Confirmation prompt if --yes is not set
	confirmTip := "Are you sure you want to execute the following SQL statement?"
	if opt.File != "" {
		confirmTip = fmt.Sprintf("Are you sure you want to execute SQL from file: %s", opt.File)
	}

	if !opt.Yes {
		ccolor.Printf("⚠️  %s\n", confirmTip)
		ccolor.Infof("SQL:\n%s\n", sql)

		if !cliutil.Confirm("Continue?") {
			ccolor.Warnln("Exiting SQL execution!")
			return nil
		}
	}

	// Execute SQL
	ccolor.Printf("🚀  Executing SQL...\n")
	result, err := db.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to execute SQL: %v", err)
	}

	// Print execution result
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		ccolor.Printf("✅  SQL executed successfully (result info not available)\n")
	} else {
		ccolor.Printf("✅  SQL executed successfully, rows affected: <green>%d</>\n", rowsAffected)
	}

	return nil
}

// readSQLFromFile reads SQL content from file
func readSQLFromFile(filePath string) (string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("invalid file path: %v", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", absPath)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(absPath))
	if ext != ".sql" {
		return "", fmt.Errorf("file must be a .sql file: %s", absPath)
	}

	// Read file content
	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	return strings.TrimSpace(string(data)), nil
}