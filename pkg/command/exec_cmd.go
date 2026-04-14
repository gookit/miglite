package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/cliutil"
	"github.com/gookit/goutil/strutil"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/goutil/x/stdio"
	"github.com/gookit/miglite/internal/database"
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
	err := initConfigAndDB()
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

	ccolor.Infop("📄  Input SQL: ")
	fmt.Println(sql)

	// Confirmation prompt if --yes is not set
	if !opt.Yes {
		ccolor.Warnf("⚠️  %s\n", confirmTip)
		if !cliutil.Confirm("Continue?") {
			ccolor.Magentaln("Exiting SQL execution!")
			return nil
		}
	}

	// 检查是否为查询语句
	sqlLower := strings.ToLower(sql)
	isQuery := strings.HasPrefix(sqlLower, "select") ||
		strings.HasPrefix(sqlLower, "describe") || // mysql
		strings.HasPrefix(sqlLower, "pragma") || // sqlite
		strings.HasPrefix(sqlLower, "show")

	// 执行查询
	if isQuery {
		return execQuery(db, sql)
	}

	// Execute DDL statement
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

func execQuery(db *database.DB, sql string) error {
	rows, err := db.Query(sql)
	if err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}
	defer stdio.SafeClose(rows)

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %v", err)
	}

	// 获取行数据
	var rowsData [][]any
	for rows.Next() {
		// 创建一个any切片来存储每列的值
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err = rows.Scan(valuePtrs...); err != nil {
			ccolor.Errorf("Failed to scan row: %v", err)
			continue
		}

		// 转换为any类型
		row := make([]any, len(columns))
		for i := range columns {
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				row[i] = string(b)
			} else {
				row[i] = val
			}
		}
		rowsData = append(rowsData, row)
	}

	// 输出结果
	ccolor.Successf("📘  Query Results(size=%d):\n", len(rowsData))
	// 输出列名
	ccolor.Cyanf("  %s\n", strings.Join(columns, "  | "))
	sb := strutil.NewBuffer(256)
	sb.WriteString("----------------------------------------------\n")

	for _, row := range rowsData {
		sb.WriteString("  ")
		for i, col := range row {
			sb.Writef("%v", col)
			if i < len(columns)-1 {
				sb.WriteString("  | ")
			}
		}
		sb.WriteRune('\n')
	}
	fmt.Println(sb.String())
	return nil
}
