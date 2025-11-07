package command

import (
	"fmt"
	"time"

	"github.com/gookit/goutil/dump"
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

func initLoadConfig() (*config.Config, error) {
	// Load configuration
	cfg, err := config.Load(ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
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
	// Load configuration
	cfg, err := initLoadConfig()
	if err != nil {
		return nil, nil, err
	}

	// Connect to database
	db, err := database.Connect(cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	db.SetDebug(ShowVerbose)
	return cfg, db, nil
}

func formatTime(t time.Time) string {
	return t.Format(TimeLayout)
}
