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
	// 默认每执行一个都需要确认
	Yes bool
	// 跳过错误迁移并继续执行
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

	c.BoolVar(&ShowVerbose, "verbose", false, "Enable verbose output;;v")
	c.StringVar(&ConfigFile, "config", "./miglite.yaml", "Path to the configuration file;;c")
	c.BoolVar(&upOpt.Yes, "yes", false, "Skip confirmation prompt;;y")
	c.IntVar(&upOpt.Number, "number", 0, "Execute only the specified number of migrations;;n")
	c.BoolVar(&upOpt.SkipErr, "skip-err", false, "Skip the error migration and continue with the execution;;s")

	// c.LongHelp = `  <mga>Note</>: if set --number, will auto set --yes=true`
	return c
}

// HandleUp executes pending migrations
func HandleUp(opt UpOption) error {
	// Load configuration and connect to database
	cfg, db, err1 := initConfigAndDB()
	if err1 != nil {
		return fmt.Errorf("failed to connect to database: %v", err1)
	}

	// Initialize schema if needed
	if err := db.InitSchema(); err != nil {
		return fmt.Errorf("failed to initialize schema: %v", err)
	}

	// Discover migrations
	migrations, err2 := migration.FindMigrations(cfg.Migrations.Path)
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

	var appliedNum int
	confirmTip := "Are you sure you want to execute this migration?"
	ccolor.Printf("🔀  Starting exec migrations(<green>pending=%d</>). Start at: %s\n\n", len(migrations), formatTime(startTime))

	// Execute pending migrations
	for idx, mig := range migrations {
		// Check if migration is already applied
		applied, status, err := migration.IsApplied(db, mig.FileName)
		if err != nil {
			return err
		}
		if applied || status == migration.StatusSkip {
			ccolor.Printf("<ylw>Skip</>ping applied migration: %s\n", mig.FileName)
			continue
		}

		// not applied OR status=down
		ccolor.Printf("<green>%d.</> Executing migration file: <green>%s</>\n", idx+1, mig.FileName)
		if !opt.Yes && !cliutil.Confirm(confirmTip) {
			ccolor.Warnln("Exiting run migrations!")
			break
		}

		if err := mig.Parse(); err != nil {
			return err
		}
		if err := executor.ExecuteUp(mig); err != nil {
			return fmt.Errorf("failed to execute migration %s: %v\nUpSQL:\n%s", mig.FileName, err, mig.UpSection)
		}

		// free memory
		mig.ResetContents()
		ccolor.Printf("Successfully executed migration: %s\n", mig.FileName)

		appliedNum++
		if opt.Number > 0 && appliedNum >= opt.Number {
			break
		}
	}

	ccolor.Successln("\n🎉  All migrations applied successfully! ⏱️ costTime:", time.Since(startTime))
	return nil
}
