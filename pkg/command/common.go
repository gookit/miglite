package command

import (
	"fmt"
	"time"

	"github.com/gookit/goutil/cflag/capp"
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

var (
	// ShowVerbose flag
	ShowVerbose bool
	// ConfigFile path to the configuration file
	ConfigFile string
)

func bindCommonFlags(c *capp.Cmd) {
	c.BoolVar(&ShowVerbose, "verbose", false, "Enable verbose output;;v")
	c.StringVar(&ConfigFile, "config", "./miglite.yaml", "Path to the configuration file;;c")
}

// cache for testing
var cfg *config.Config

// Cfg get config instance
func Cfg() *config.Config {
	return cfg
}

func initLoadConfig() (*config.Config, error) {
	var err error
	if cfg != nil {
		return cfg, nil
	}

	// Load configuration
	cfg, err = config.Load(ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	// fire OnConfigLoaded hook
	if OnConfigLoaded != nil {
		if err = OnConfigLoaded(cfg); err != nil {
			return nil, err
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

	return cfg, nil
}

func initConfigAndDB() (*database.DB, error) {
	var err error
	// Load configuration
	cfg, err = initLoadConfig()
	if err != nil {
		return nil, err
	}
	dbCfg := cfg.Database

	// Connect to database
	db, err := database.Connect(dbCfg.Driver, dbCfg.SqlDriver, dbCfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	db.SetDebug(ShowVerbose)
	ccolor.Printf("✅  Database connect successful! driver: <green>%s</>\n", db.Driver())
	return db, nil
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
