package command

import (
	"fmt"
	"time"

	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/cliutil"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/pkg/migration"
)

// UpOption represents options for the up command
type UpOption struct {
	// 默认每执行一个都需要确认 default: false
	Yes bool
	// 跳过错误迁移并继续执行 default: false
	SkipErr bool
	// 只执行指定数量的迁移
	Number int
	// 查找迁移开始时间，默认只查找最近6个月的迁移文件
	StartTime string
}

// NewUpCommand executes pending migrations
func NewUpCommand() *capp.Cmd {
	var upOpt = UpOption{}

	c := capp.NewCmd("up", "Execute pending migrations", func(c *capp.Cmd) error {
		return HandleUp(upOpt)
	})

	c.Aliases = []string{"migrate", "run"}
	bindCommonFlags(c)

	c.BoolVar(&upOpt.Yes, "yes", false, "Skip confirmation prompt;;y")
	c.IntVar(&upOpt.Number, "number", 0, "Execute only the specified number of migrations;;n")
	c.BoolVar(&upOpt.SkipErr, "skip-err", false, "Skip the error migration and continue with the execution;;s")

	// c.LongHelp = `  <mga>Note</>: if set --number, will auto set --yes=true`
	return c
}

// HandleUp executes pending migrations
func HandleUp(opt UpOption) error {
	// Load configuration and connect to database
	if err1 := initConfigAndDB(); err1 != nil {
		return fmt.Errorf("failed to connect to database: %v", err1)
	}
	defer db.SilentClose()

	// Initialize schema if needed
	if err := db.InitSchema(); err != nil {
		return fmt.Errorf("failed to initialize schema: %v", err)
	}

	// Discover migrations
	migrations, err2 := findMigrations()
	if err2 != nil {
		return fmt.Errorf("failed to discover migrations: %v", err2)
	}

	if len(migrations) == 0 {
		ccolor.Infoln("🔎  No migrations found.")
		return nil
	}

	// Get executor
	executor := migration.NewExecutor(db, ShowVerbose)
	startTime := time.Now()

	var appliedNum, skippedNum int
	var splitSkipped = !ShowVerbose
	confirmTip := "Are you sure you want to execute this migration?"
	ccolor.Printf("🚀  Starting exec migrations(<green>founds=%d</>). Start at: %s\n\n", len(migrations), formatTime(startTime))

	// Execute pending migrations
	for idx, mig := range migrations {
		// Check if migration is already applied
		applied, status, err := migration.IsApplied(db, mig.FileName)
		if err != nil {
			return err
		}
		if applied || status == migration.StatusSkip {
			skippedNum++
			if ShowVerbose {
				ccolor.Printf("%d. ⏭️  <ylw>Skipping</> %s migration: %s\n", idx+1, migration.StatusText(status), mig.FileName)
			} else {
				ccolor.Infop(".")
				splitSkipped = true
			}
			continue
		}

		if splitSkipped {
			fmt.Println()
			splitSkipped = false
		}

		// not applied OR status=down
		ccolor.Printf("<green>%d.</> 🔄  Executing migration file: <green>%s</>\n", idx+1, mig.FileName)
		if !opt.Yes && !cliutil.Confirm(confirmTip) {
			ccolor.Warnln("Exiting run migrations!")
			break
		}

		if err = mig.Parse(); err != nil {
			return err
		}
		if err = executor.ExecuteUp(mig); err != nil {
			return fmt.Errorf("failed to execute migration %s: %v\nUpSQL:\n%s", mig.FileName, err, mig.UpSection)
		}

		// free memory
		mig.ResetContents()
		ccolor.Printf("✅  Successfully executed migration: %s\n", mig.FileName)

		appliedNum++
		if opt.Number > 0 && appliedNum >= opt.Number {
			break
		}
	}

	ccolor.Successf("\n\n🎉  All migrations applied successfully! 📘 apply:%d, skip:%d ⏱️ duration: %s\n", appliedNum, skippedNum, time.Since(startTime))
	return nil
}
