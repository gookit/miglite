package command

import (
	"fmt"

	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/pkg/config"
	"github.com/gookit/miglite/pkg/database"
)

var (
	showVerbose bool
	// ConfigFile path to the configuration file
	configFile string
)

func initConfigAndDB() (*config.Config, *database.DB, error) {
	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %v", err)
	}

	ccolor.Printf("Loaded config file: %s\n", configFile)
	if showVerbose {
		fmt.Println("Config:")
		dump.NoLoc(cfg)
	}

	// Connect to database
	db, err := database.Connect(cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	db.SetDebug(showVerbose)
	return cfg, db, nil
}