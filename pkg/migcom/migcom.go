package migcom

// supported database driver names
const (
	DriverMySQL    = "mysql"
	DriverMSSQL    = "mssql"
	DriverPostgres = "postgres"
	DriverSQLite   = "sqlite"
)

// DefaultConfigFile default config file.
const DefaultConfigFile = "./miglite.yaml"

// DefaultMigrationsDir default migrations dirpath. can override by env: MIGRATIONS_PATH
const DefaultMigrationsDir = "./migrations"
