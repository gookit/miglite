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
)

const TimeLayout = "2006-01-02 15:04:05"

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

func initConfigAndDB() (*config.Config, *database.DB, error) {
	var err error
	// Load configuration
	cfg, err = initLoadConfig()
	if err != nil {
		return nil, nil, err
	}
	dbCfg := cfg.Database

	// Connect to database
	db, err := database.Connect(dbCfg.Driver, dbCfg.SqlDriver, dbCfg.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	db.SetDebug(ShowVerbose)
	ccolor.Printf("✅  Database connect successful! driver: <green>%s</>\n", db.Driver())
	return cfg, db, nil
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	return t.Format(TimeLayout)
}
