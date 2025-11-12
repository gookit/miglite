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
	// SQL or sql-file to execute
	SQLOrFile string
	// Skip confirmation prompt
	Yes bool
}

// NewExecCommand executes SQL statement or SQL file directly
func NewExecCommand() *capp.Cmd {
	var execOpt = ExecOption{}

	c := capp.NewCmd("exec", "Execute SQL statement or SQL file directly", func(c *capp.Cmd) error {
		execOpt.SQLOrFile = c.Arg("sql-or-file").String()
		return HandleExec(execOpt)
	})

	c.Aliases = []string{"execute", "run-sql"}
	bindCommonFlags(c)

	// c.StringVar(&execOpt.SQL, "sql", "", "SQL statement to execute;;s")
	// c.StringVar(&execOpt.File, "file", "", "Path to SQL file to execute;;f")
	c.BoolVar(&execOpt.Yes, "yes", false, "Skip confirmation prompt;;y")

	c.AddArg("sql-or-file", "SQL statement/file to execute", true, nil)
	return c
}

// HandleExec handles the exec command logic
func HandleExec(opt ExecOption) error {
	// Validate options
	sqlOrFile := strings.TrimSpace(opt.SQLOrFile)
	if sqlOrFile == "" {
		return fmt.Errorf("either SQL or sql-file must be provided")
	}

	// Load configuration and connect to database
	_, db, err := initConfigAndDB()
	if err != nil {
		return err
	}
	defer db.SilentClose()

	// Prepare SQL to execute
	var sql = sqlOrFile
	var sqlFile string
	confirmTip := "Are you sure you want to execute the following SQL statement?"

	// if sqlOrFile is a sql file path, read SQL from file
	if len(sqlOrFile) < 128 && strings.HasSuffix(sqlOrFile, ".sql") {
		sqlFile = sqlOrFile
		confirmTip = fmt.Sprintf("Are you sure you want to execute SQL from file: %s", sqlFile)

		// Read SQL from file
		sql, err = readSQLFromFile(sqlFile)
		if err != nil {
			return fmt.Errorf("failed to read SQL file: %v", err)
		}
		if sql == "" {
			return fmt.Errorf("no SQL contents in file: %s", sqlFile)
		}
	}

	ccolor.Infoln("Input SQL:")
	fmt.Println(sql)

	// Confirmation prompt if --yes is not set
	if !opt.Yes {
		ccolor.Warnf("⚠️  %s\n", confirmTip)
		if !cliutil.Confirm("Continue?") {
			ccolor.Magentaln("Exiting SQL execution!")
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
	if _, err = os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", absPath)
	}

	// Read file content
	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	return strings.TrimSpace(string(data)), nil
}
