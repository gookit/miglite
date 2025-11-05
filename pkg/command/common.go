package command

import (
	"fmt"

	"github.com/gookit/miglite/pkg/config"
	"github.com/gookit/miglite/pkg/database"
)

var (
	showVerbose bool
	// ConfigFile path to the configuration file
	configFile string
)

func loadConfigAndDB() (*config.Config, *database.DB, error) {
	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.Connect(cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return cfg, db, nil
}