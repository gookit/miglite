package command

import (
	"fmt"
	"time"

	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/envutil"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/pkg/config"
	"github.com/gookit/miglite/pkg/database"
	"github.com/gookit/miglite/pkg/migration"
)

const TimeLayout = "2006-01-02 15:04:05"

// OnConfigLoaded hook. you can modify or validate the configuration here.
var OnConfigLoaded = func(cfg *config.Config) error {
	return nil
}

// cache for testing
var cfg *config.Config
var db *database.DB

// Cfg get config instance
func Cfg() *config.Config { return cfg }

// SetCfg set config instance. use on manual run logic.
func SetCfg(c *config.Config) {
	cfg = c
	ConfigFile = c.ConfigFile
	ShowVerbose = c.Verbose
}

// DB get database instance
func DB() *database.DB { return db }

// SetDB set database instance. use on manual run logic.
func SetDB(d *database.DB) { db = d }

func initLoadConfig() error {
	if cfg != nil {
		return nil
	}

	// Load configuration
	var err error
	cfg, err = config.Load(ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// fire OnConfigLoaded hook
	if OnConfigLoaded != nil {
		if err = OnConfigLoaded(cfg); err != nil {
			return err
		}
	}

	if envFiles := envutil.LoadedEnvFiles(); len(envFiles) > 0 {
		ccolor.Printf("📄  Loaded environment variables from <green>%s</>\n", envFiles[0])
	}
	if cfg.ConfigFile != "" {
		ccolor.Printf("📄  Loaded config file from <green>%s</>\n", cfg.ConfigFile)
	}
	if ShowVerbose {
		dump.NoLoc(cfg)
	}

	return nil
}

func initConfigAndDB() (err error) {
	// Load configuration
	if err = initLoadConfig(); err != nil {
		return err
	}

	// Connect to database
	if db == nil {
		dbCfg := cfg.Database
		db, err = database.NewDB(dbCfg.Driver, dbCfg.SqlDriver, dbCfg.DSN)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %v", err)
		}
		ccolor.Printf("✅  Database connect successful! driver: <green>%s</>\n", db.Driver())
	}

	db.SetDebug(ShowVerbose)
	return nil
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	return t.Format(TimeLayout)
}

func findMigrations() ([]*migration.Migration, error) {
	return migration.FindMigrations(cfg.Migrations.Path, cfg.Migrations.Recursive)
}
