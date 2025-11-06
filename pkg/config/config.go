package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/miglite/pkg/util"
)

type Database struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type Migrations struct {
	// Path to the migrations directory.
	//  - allow use string-vars: {driver}
	Path string `yaml:"path"`
}

// Config holds the application configuration
type Config struct {
	Database   Database   `yaml:"database"`
	Migrations Migrations `yaml:"migrations"`
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
	if err := loadFromENV(config); err != nil {
		return nil, err
	}

	// Validate db configuration
	if err := checkDatabaseConfig(config); err != nil {
		return nil, err
	}

	// Set defaults if not defined
	if config.Migrations.Path == "" {
		config.Migrations.Path = "./migrations"
	}
	if strings.Contains(config.Migrations.Path, "{driver}") {
		config.Migrations.Path = strings.Replace(config.Migrations.Path, "{driver}", config.Database.Driver, 1)
	}

	return config, nil
}

func checkDatabaseConfig(config *Config) error {
	// Validate configuration
	if config.Database.Driver == "" {
		return fmt.Errorf("database driver is required")
	}
	if config.Database.DSN == "" {
		return fmt.Errorf("database DSN is required")
	}

	// format driver name
	driver, err := util.ResolveDriver(config.Database.Driver)
	if err != nil {
		return err
	}

	config.Database.Driver = driver
	return nil
}

func loadFromENV(config *Config) error {
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
			return err
		}
		config.Database.Driver = driver
		config.Database.DSN = dsn
	}

	if path := os.Getenv("MIGRATIONS_PATH"); path != "" {
		config.Migrations.Path = path
	}
	return nil
}

// parseDatabaseURL infers the database driver and DSN from a DATABASE_URL
func parseDatabaseURL(url string) (string, string, error) {
	if url == "" {
		return "", "", fmt.Errorf("DATABASE_URL is empty")
	}

	// url eg: mysql://user:password@localhost:3306/dbname
	sepIdx := strings.Index(url, "://")
	if sepIdx < 1 {
		return "", "", fmt.Errorf("invalid DATABASE_URL: %s", url)
	}

	driver, err := util.ResolveDriver(url[:sepIdx])
	if err != nil {
		return "", "", err
	}

	dsnIndex := sepIdx + 3
	return driver, url[dsnIndex:], nil
}
