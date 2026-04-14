package miglite

import (
	"github.com/gookit/miglite/internal/config"
	"github.com/gookit/miglite/internal/database"
)

// Config is the configuration struct for the Migrator
type Config = config.Config

// ConfigFn is a function type for updating the configuration
type ConfigFn func(c *Config)

// SetEnvPrefix set environment prefix
func SetEnvPrefix(prefix string) {
	config.EnvPrefix = prefix
}

// SqlProvider is the interface for database provider
type SqlProvider = database.SqlProvider

// SetSchemaTableName set schema table name
func SetSchemaTableName(tableName string) {
	database.SchemaTableName = tableName
}

// AddSqlProvider add database provider
func AddSqlProvider(driver string, provider SqlProvider) {
	database.AddProvider(driver, provider)
}
