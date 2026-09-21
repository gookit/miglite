package miglite

import (
	"io/fs"

	"github.com/gookit/miglite/internal/config"
	"github.com/gookit/miglite/internal/database"
	"github.com/gookit/miglite/pkg/command"
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

// BindFS binds a process level migrations filesystem, eg. an embed.FS, for an
// application built with pkg/command (see command.NewApp). BindFS(nil) restores
// local filesystem reading. Library callers should prefer Migrator.SetFS.
func BindFS(fsys fs.FS) {
	command.SetMigrationFS(fsys)
}
