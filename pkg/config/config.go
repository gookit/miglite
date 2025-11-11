package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/ini/v2/dotenv"
	"github.com/gookit/miglite/pkg/migutil"
)

// Database configuration
type Database struct {
	// Driver name for database. 标准数据库类型名称。
	//
	// eg: mysql, postgres, sqlite, mssql, ...
	//
	// NOTE: 当使用的驱动库注册名称不一致时，需要配置 SqlDriver 用于 sql.Open()
	Driver string `yaml:"driver"`
	// SqlDriver sql driver name. 数据库驱动库注册到 database/sql 的名称
	//
	// NOTE: 跟使用的数据库驱动库有关
	SqlDriver string `yaml:"sql_driver" json:"sql_driver"`
	// DSN 连接配置
	DSN string `yaml:"dsn"`

	// 可以使用拆分配置项 - DSN 为空时，会跟据下面的信息构建DSN
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	DBName   string `yaml:"dbname" json:"dbname"`
	SSLMode  string `yaml:"ssl_mode" json:"ssl_mode"`

	// Connection pool settings
	MaxIdleConns    int `yaml:"max_idle_conns" json:"max_idle_conns"`
	MaxOpenConns    int `yaml:"max_open_conns" json:"max_open_conns"`
	ConnMaxIdleTime int `yaml:"conn_max_idle_time" json:"conn_max_idle_time"`
	ConnMaxLifetime int `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
}

// Migrations configuration
type Migrations struct {
	// Path to the migrations file directory.
	//  - allow use string-vars: {driver}
	Path string `yaml:"path"`
	// Use date(YYYY) as directory TODO
	DateDir bool `yaml:"date_dir"`
	// Table name for migration tracking TODO
	Table string `yaml:"table"`
}

// Config holds the application configuration
type Config struct {
	ConfigFile string     `yaml:"-"` // internal use
	Database   Database   `yaml:"database"`
	Migrations Migrations `yaml:"migrations"`
}

var std *Config

// Get returns the default configuration
func Get() *Config {
	if std == nil {
		panic("config not load and initialized")
	}
	return std
}

// Reset the default configuration
func Reset() { std = nil }

// Load loads configuration from YAML file and environment variables
//
// NOTE: will auto load .env file on working directory.
func Load(configPath string) (*Config, error) {
	// load .env file
	_ = dotenv.LoadExists("./")
	config := &Config{}

	// Load from YAML file if it exists
	if fsutil.FileExist(configPath) {
		config.ConfigFile = configPath
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, err
		}
		if err = yaml.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	// Override with environment variables
	if err := setDBConfigFromENV(&config.Database); err != nil {
		return nil, err
	}

	// Validate db configuration
	if err := checkDatabaseConfig(&config.Database); err != nil {
		return nil, err
	}

	// Set migrations config
	initMigrationsConfig(&config.Migrations, config.Database.Driver)

	std = config
	return config, nil
}

func initMigrationsConfig(migConfig *Migrations, fmtDriver string) {
	if path := os.Getenv("MIGRATIONS_PATH"); path != "" {
		migConfig.Path = path
	}

	// Set defaults if not defined
	dirPath := migConfig.Path
	if dirPath == "" {
		dirPath = "./migrations"
	} else if strings.Contains(dirPath, "{driver}") {
		dirPath = strings.Replace(dirPath, "{driver}", fmtDriver, 1)
	}

	migConfig.Path = dirPath
}

func checkDatabaseConfig(dbCfg *Database) error {
	// Validate configuration
	if dbCfg.Driver == "" {
		return fmt.Errorf("database driver is required")
	}

	driver := dbCfg.Driver
	// format driver name
	fmtDriver := migutil.FmtDriverName(driver)
	dbCfg.Driver = fmtDriver
	if dbCfg.SqlDriver == "" {
		dbCfg.SqlDriver = driver
	}

	// check DSN
	if dbCfg.DSN == "" {
		dbDSN := buildDSNFromConfig(dbCfg)
		if dbDSN == "" {
			return fmt.Errorf("database DSN is required")
		}
	}
	return nil
}

func setDBConfigFromENV(dbCfg *Database) error {
	if dsn := os.Getenv("DATABASE_DSN"); dsn != "" {
		dbCfg.DSN = dsn
	}
	if driver := os.Getenv("DATABASE_DRIVER"); driver != "" {
		dbCfg.Driver = driver
		dbCfg.SqlDriver = driver
	}

	// Infer driver from DATABASE_URL
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		driver, dsn, err := parseDatabaseURL(dbURL)
		if err != nil {
			return err
		}

		dbCfg.DSN = dsn
		dbCfg.Driver = driver
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
		return "", "", fmt.Errorf("invalid DATABASE_URL: %s(Should 'driver://DSN')", url)
	}

	dsnIndex := sepIdx + 3
	return url[:sepIdx], url[dsnIndex:], nil
}

func buildDSNFromConfig(dbCfg *Database) string {
	// sqlite
	if dbCfg.Driver == "sqlite" {
		return dbCfg.DSN
	}

	// username is required
	if dbCfg.User == "" {
		return ""
	}

	if dbCfg.Host == "" {
		dbCfg.Host = "localhost"
	}

	// mysql
	if dbCfg.Driver == "mysql" {
		if dbCfg.Port <= 0 {
			dbCfg.Port = 3306
		}
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.DBName,
		)
	}

	// postgres
	if dbCfg.Driver == "postgres" {
		if dbCfg.Port <= 0 {
			dbCfg.Port = 5432
		}
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			dbCfg.Host, dbCfg.Port, dbCfg.User, dbCfg.Password, dbCfg.DBName, dbCfg.SSLMode,
		)
	}

	// mssql
	if dbCfg.Driver == "mssql" {
		if dbCfg.Port <= 0 {
			dbCfg.Port = 1433
		}
		return fmt.Sprintf(
			"server=%s;user id=%s;password=%s;database=%s;port=%d;",
			dbCfg.Host, dbCfg.User, dbCfg.Password, dbCfg.DBName, dbCfg.Port,
		)
	}

	return ""
}
