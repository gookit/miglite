package config

import (
	"fmt"
	"os"

	"github.com/gookit/goutil/fsutil"
	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	Database struct {
		Driver string `yaml:"driver"`
		DSN    string `yaml:"dsn"`
	} `yaml:"database"`
	Migrations struct {
		Path string `yaml:"path"`
	} `yaml:"migrations"`
}

// Load loads configuration from YAML file and environment variables
func Load(configPath string) (*Config, error) {
	config := &Config{}

	// Load from YAML file if it exists
	if fsutil.FileExist(configPath) {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, err
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	// Override with environment variables
	if driver := os.Getenv("DATABASE_DRIVER"); driver != "" {
		config.Database.Driver = driver
	}
	if dsn := os.Getenv("DATABASE_DSN"); dsn != "" {
		config.Database.DSN = dsn
	}
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		// Infer driver from DATABASE_URL
		driver, dsn, err := parseDatabaseURL(dbURL)
		if err != nil {
			return nil, err
		}
		config.Database.Driver = driver
		config.Database.DSN = dsn
	}
	if path := os.Getenv("MIGRATIONS_PATH"); path != "" {
		config.Migrations.Path = path
	}

	// Set defaults if not defined
	if config.Migrations.Path == "" {
		config.Migrations.Path = "./migrations"
	}

	// Validate configuration
	if config.Database.Driver == "" {
		return nil, fmt.Errorf("database driver is required")
	}
	if config.Database.DSN == "" {
		return nil, fmt.Errorf("database DSN is required")
	}

	return config, nil
}

// parseDatabaseURL infers the database driver and DSN from a DATABASE_URL
func parseDatabaseURL(url string) (string, string, error) {
	if url == "" {
		return "", "", fmt.Errorf("DATABASE_URL is empty")
	}

	switch {
	case len(url) > 5 && url[:5] == "mysql":
		return "mysql", url[8:], nil
	case len(url) > 8 && url[:8] == "postgres":
		return "postgres", url[11:], nil
	case len(url) > 6 && url[:6] == "sqlite":
		return "sqlite3", url[9:], nil
	default:
		return "", "", fmt.Errorf("unsupported database URL: %s", url)
	}
}