package command

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gookit/goutil/arrutil"
	"github.com/gookit/goutil/cliutil"
	"github.com/gookit/goutil/strutil"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/internal/database"
	"github.com/gookit/miglite/internal/runtime"
	"github.com/gookit/miglite/pkg/migration"
)

// This file is the terminal output adapter. Both the CLI handlers (Handle*) and
// miglite.Migrator go through the Run* functions, so both print identical text.
//
// NOTE: the Run* signatures reference internal/runtime, so they are only usable
// inside this module; they are exported for miglite.Migrator. interactive=false
// means a non-CLI caller: confirmations are skipped, because the Yes option of
// command.*Option only affects the CLI.

// RunInit initializes the migration schema and prints the result.
func RunInit(r *runtime.Runtime, opt InitOption) error {
	if err := r.Init(runtime.InitOption{Drop: opt.Drop}); err != nil {
		return err
	}
	ccolor.Infoln("🎉  Migration schema initialized successfully.")
	return nil
}

// RunUp runs pending migrations.
func RunUp(r *runtime.Runtime, opt UpOption, interactive bool) error {
	start := time.Now()
	// matches the v0.6.0 CLI: a blank line separates the "Skipping" dots from the
	// progress lines, so non-verbose output starts with one
	skipNewline := !ShowVerbose
	return r.UpWithHooks(runtime.UpOption{SkipErr: opt.SkipErr, Number: opt.Number, StartTime: opt.StartTime}, runtime.MigrationHooks{
		Start: func(total int) {
			if total == 0 {
				ccolor.Infoln("🔎  No migrations found.")
				return
			}
			ccolor.Printf("🚀  Starting exec migrations(<green>founds=%d</>). Start at: %s\n\n", total, formatTime(start))
		},
		Complete: func(res runtime.MigrationResult) {
			if res.Failed > 0 {
				ccolor.Warnf("\n\n⚠️  %d migration(s) failed! 📘 apply:%d, skip:%d, failed:%d ⏱️ duration: %s\n",
					res.Failed, res.Applied, res.Skipped, res.Failed, time.Since(start))
				return
			}
			ccolor.Successf("\n\n🎉  All migrations applied successfully! 📘 apply:%d, skip:%d ⏱️ duration: %s\n",
				res.Applied, res.Skipped, time.Since(start))
		},
		Before: func(idx, _ int, mig *migration.Migration) error {
			if skipNewline {
				fmt.Println()
				skipNewline = false
			}
			ccolor.Printf("<green>%d.</> 🔄  Executing migration file: <green>%s</>\n", idx+1, mig.FileName)
			if interactive && !opt.Yes && !cliutil.Confirm("Are you sure you want to execute this migration?") {
				ccolor.Warnln("Exiting run migrations!")
				return runtime.ErrCancelled
			}
			return nil
		},
		Skip: func(idx, _ int, mig *migration.Migration, status string) {
			if ShowVerbose {
				ccolor.Printf("%d. ⏭️  <ylw>Skipping</> %s migration: %s\n", idx+1, migration.StatusText(status), mig.FileName)
				return
			}
			ccolor.Infop(".")
			skipNewline = true
		},
		After: func(_, _ int, mig *migration.Migration) {
			ccolor.Printf("✅  Successfully executed migration: %s\n", mig.FileName)
		},
		Error: func(_, _ int, mig *migration.Migration, err error) {
			ccolor.Errorf("❌  Failed migration %s: %v\n", mig.FileName, err)
			ccolor.Printf("UpSQL:\n%s\n", mig.UpSection)
		},
	})
}

// RunDown rolls back applied migrations.
func RunDown(r *runtime.Runtime, opt DownOption, interactive bool) error {
	rolledBack := 0
	err := r.DownWithHooks(runtime.DownOption{Number: opt.Number}, runtime.MigrationHooks{
		Start: func(total int) {
			if total == 0 {
				ccolor.Println("🔎  No applied migrations to rollback")
				return
			}
			ccolor.Printf("🚀  Will roll back recent %d migrations:\n\n", total)
		},
		Before: func(idx, _ int, mig *migration.Migration) error {
			ccolor.Printf("%d. Rolling back migration: <ylw>%s</> (appliedAt %s)\n", idx+1, mig.FileName, formatTime(mig.AppliedAt))
			if interactive && !opt.Yes && !cliutil.Confirm("Are you sure you want to roll back the migration?") {
				ccolor.Warnln("Skipping rollback the migration!")
				return runtime.ErrCancelled
			}
			return nil
		},
		After: func(_, _ int, mig *migration.Migration) {
			rolledBack++
			ccolor.Printf("✅  Success rolled back migration: %s\n", mig.FileName)
		},
		Skip: func(_, _ int, _ *migration.Migration, status string) {
			if status == runtime.StatusEmptyDown {
				ccolor.Warnln("Skipping empty DOWN migration!")
			}
		},
		Error: func(_, _ int, mig *migration.Migration, err error) {
			ccolor.Errorf("Failed to roll back migration %s: %v\n", mig.FileName, err)
			ccolor.Printf("DownSQL:\n%s\n", mig.DownSection)
		},
	})
	if err == nil && rolledBack > 0 {
		ccolor.Successf("\n🎉  Successfully rolled back %d migration(s)\n", rolledBack)
	}
	return err
}

// RunSkip marks migration files as skipped.
func RunSkip(r *runtime.Runtime, opt SkipOption) error {
	return r.SkipWithHooks(runtime.SkipOption{FileNames: opt.FileNames}, runtime.MigrationHooks{
		Start: func(total int) {
			ccolor.Magentaf("🚀  Start ignore %d migrations:\n\n", total)
		},
		Skip: func(_, _ int, mig *migration.Migration, _ string) {
			ccolor.Warnf("Migration %s already skipped", mig.Version)
		},
		After: func(_, _ int, mig *migration.Migration) {
			ccolor.Printf("- Migration <green>%s</> skipped", mig.Version)
		},
	})
}

// RunStatus loads the migration status and prints the status table.
func RunStatus(r *runtime.Runtime) error {
	records, err := r.Status(runtime.StatusOption{})
	if err != nil {
		return err
	}
	PrintStatusRecords(records)
	return nil
}

// RunShow shows database tables or one table schema.
func RunShow(r *runtime.Runtime, opt ShowOption) error {
	if err := validateShowOption(opt); err != nil {
		return err
	}

	if opt.Tables {
		ccolor.Println("🔍  Fetching database tables...")
	} else {
		ccolor.Printf("🔍  Fetching schema for table: <green>%s</>\n", opt.Schema)
	}

	result, err := r.Show(runtime.ShowOption{Tables: opt.Tables, Schema: opt.Schema})
	if err != nil {
		return err
	}
	if opt.Tables {
		return printTablesList(result.Tables)
	}
	return printTableSchema(result.Columns, opt.Schema)
}

// RunExec executes SQL statements or a SQL file.
func RunExec(r *runtime.Runtime, opt ExecOption, interactive bool) error {
	hooks := runtime.ExecHooks{
		Input: func(sqlText string, _ bool) {
			ccolor.Infop("📄  Input SQL: ")
			fmt.Println(sqlText)
		},
		BeforeStatement: func(idx, total int, _ string) error {
			ccolor.Printf("🚀  Executing SQL statement %d/%d...\n", idx+1, total)
			return nil
		},
		AfterStatement: func(_, _ int, _ string, result sql.Result) {
			rows, err := result.RowsAffected()
			if err != nil {
				ccolor.Printf("✅  SQL executed successfully (result info not available)\n")
				return
			}
			ccolor.Printf("✅  SQL executed successfully, rows affected: <green>%d</>\n", rows)
		},
		QueryResult: func(_, _ int, _ string, qr runtime.QueryResult) {
			printQueryResult(qr.Columns, qr.Rows)
		},
	}
	if interactive {
		hooks.Confirm = func(message string) bool {
			ccolor.Warnf("⚠️  %s\n", message)
			ok := cliutil.Confirm("Continue?")
			if !ok {
				ccolor.Magentaln("Exiting SQL execution!")
			}
			return ok
		}
	}

	err := r.ExecWithHooks(runtime.ExecOption{SQLOrFile: opt.SQLOrFile, Yes: opt.Yes}, hooks)
	if errors.Is(err, runtime.ErrCancelled) {
		return nil
	}
	return err
}

// PrintStatusRecords prints the migration status table.
func PrintStatusRecords(records []migration.Record) {
	ccolor.Cyanf("\n📊  Migrations Status:(total=%d)\n", len(records))
	fmt.Println(strings.Repeat("==", 44))
	ccolor.Printf("  <b>Status</>  | %13s<b>Version(migration file)</>%13s    |   <b>Operate Time</> \n", "", "")
	fmt.Println(strings.Repeat("--", 44))
	for _, st := range records {
		statusIcon := "<mga>pending</>"
		switch st.Status {
		case migration.StatusUp:
			statusIcon = "<green>applied</>"
		case migration.StatusDown:
			statusIcon = "<ylw>rolled</> "
		case migration.StatusSkip:
			statusIcon = "<gray>skipped</>"
		}
		ccolor.Printf("  %s | %-52s | %s\n", statusIcon, st.Version, formatTime(st.AppliedAt))
	}
}

func validateShowOption(opt ShowOption) error {
	if !opt.Tables && opt.Schema == "" {
		return fmt.Errorf("either --tables or --schema must be provided")
	}
	if opt.Tables && opt.Schema != "" {
		return fmt.Errorf("--tables and --schema cannot be used together")
	}
	return nil
}

// printTablesList prints all database tables, hiding the migration schema table.
func printTablesList(tables []string) error {
	tables = arrutil.Filter(tables, func(s string) bool {
		return s != database.SchemaTableName
	})
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

// printTableSchema prints the columns of one table.
func printTableSchema(columns []database.ColumnInfo, tableName string) error {
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

func printQueryResult(columns []string, rowsData [][]any) {
	ccolor.Successf("📘  Query Results(size=%d):\n", len(rowsData))
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
}
