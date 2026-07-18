# miglite - Lite database schema migration tool by Go

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/miglite?style=flat-square)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/gookit/miglite)](https://github.com/gookit/miglite)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/miglite)](https://goreportcard.com/report/github.com/gookit/miglite)
[![Unit-Tests](https://github.com/gookit/miglite/workflows/Unit-Tests/badge.svg)](https://github.com/gookit/miglite/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/gookit/miglite.svg)](https://pkg.go.dev/github.com/gookit/miglite)

> **👉 [中文说明](README.zh-CN.md)**

`miglite` is a minimalist database schema migration tool implemented in Golang.

- Easy to use with minimal dependencies
- Developed based on `database/sql` without adding any driver dependencies by default
- Migration SQL is executed within transactions to ensure data consistency
- Uses raw SQL files as migration files
  - SQL filename format: `YYYYMMDD-NNNNNN-{migration-name}.sql`
  - The `YYYYMMDD` date must be valid; the next 6 digits only need to be comparable numbers
- By default, all SQL files (including subdirectories) in the migration directory are recursively searched
  - Directories starting with `_` (eg. `_backup/xx.sql`) are ignored when looking for SQL files
  - Migration directories support the use of environment variables (eg `.migrations/${MODULE_NAME}`)
  - Migration directories support adding multiple paths using comma `,` splitting
- Can run migrations with zero configuration via environment variables (e.g., `DATABASE_URL`, `MIGRATIONS_PATH`)
  - Automatically attempts to load `.env` file in the directory(Optional)
  - Automatically tries default configuration files `./miglite.yaml`, `./miglite.local.yaml`(Optional)
- Built-in support for executing SQL statements via `miglite exec`, convenient for debugging and testing
- Supports `mysql`, `sqlite`, `postgres` databases
  - When used as a library, you need to add your own DB driver dependencies
  - When using the `miglite` command-line tool directly, driver dependencies are already included

## Installation

Using the `miglite` command-line tool:

```bash
# install it by go
go install github.com/gookit/miglite/cmd/miglite@latest
```

Using as a Go dependency library:

```bash
go get github.com/gookit/miglite

# import "github.com/gookit/miglite"
```

## Direct CLI Usage

Using the `miglite` command-line tool directly.

![help](./testdata/help.png)

**Commands**:

```bash
  create, new                 Create new migration SQL files
  down, rollback              Rollback the most recent migration
  exec, execute, run-sql      Execute SQL statement or SQL file directly
  init                        Initialize the migration schema on database
  show, info, describe        Show database information like tables or table schema
  skip, ignore                Manual skip one or multi migration file(s)
  status, st                  Show the status of migrations
  up, migrate, run            Execute pending migrations
  help                        Display application help
```

### Configuration

`miglite` supports configuration via `miglite.yaml` file or environment variables.

- Can work without a configuration file, using the environment variable `DATABASE_URL` directly
- When `--config` is not specified, `miglite` tries `./miglite.yaml`, then `./miglite.local.yaml`
- A specific configuration file can be set with the `--config` parameter
- Dotenv files are auto-loaded from `.env.local`, `.env.dev`, `.env` by default; specify another file with `--env-file` or `--efile`
- `--db` overrides the database name from YAML, `DATABASE_DSN`, or `DATABASE_URL`; for SQLite, it overrides the database file path

#### miglite.yaml Example

```yaml
database:
  driver: sqlite  # or mysql, postgresql
  # Prefer dsn for database connection.
  dsn: ./sqlite_test.db
  # These split connection settings are used when dsn is empty.
  host: localhost
  port: 5432
  user: ${PG_DB_USER}
  password: ${PG_DB_PWD | pg1234abcd}
  dbname: pg_test_db
  ssl_mode: disable
migrations:
  path: ./migrations
```

> As shown above, config file values also support ENV placeholders.

#### Environment Variables

- `MIGRATIONS_PATH`: Migration files directory path (default: `./migrations`)
  - Supports using environment variables (eg `./migrations/${MODULE_NAME}`)
  - Supports adding multiple paths using comma `,` splitting
- `DATABASE_URL`: Database connection URL (e.g., `sqlite://path/to/your.db`, `mysql://user:pass@tcp(host:port)/dbname`)
- `MIGLITE_ENV_PREFIX`: Environment variable prefix, default is empty string
  - After setting, environment variables will be prefixed with the prefix, eg. `MIGLITE_ENV_PREFIX=MY_`, then read `MY_DATABASE_URL` instead of `DATABASE_URL`
  - All environment variables will be affected by the prefix setting

**Examples**:

```ini
MIGRATIONS_PATH = "./migrations"
# sqlite
DATABASE_URL="sqlite://path/to/your.db"
# mysql
DATABASE_URL="mysql://user:passwd@tcp(127.0.0.1:3306)/local_test?charset=utf8mb4&parseTime=True&loc=Local"
# postgresql
DATABASE_URL="postgres://host=localhost port=5432 user=username password=password dbname=dbname sslmode=disable"
```

Use a custom env file:

```bash
miglite --env-file ./configs/dev.env status
miglite up --efile ./configs/dev.env
```

Override the target database:

```bash
miglite --db new_db status
miglite exec --db new_db --yes "SELECT current_database();"
```

> **NOTE**: mysql DSNs must be tagged with the 'tcp(...)' protocol. Otherwise, it will throw an error.

### Creating Migrations

```bash
miglite create add-users-table
```

This will create an SQL file named with the current timestamp in the `./migrations/` directory, with the format `YYYYMMDD-HHMMSS-add-users-table.sql`.

```text
./migrations/20251105-102325-create-users-table.sql
```

SQL file content includes a template:

```sql
-- Migrate:UP
-- Add migration SQL here

-- Migrate:DOWN
-- Add rollback SQL here (optional)
```

Example migration file:

```sql
-- Migrate:UP
CREATE TABLE post (
  id int NOT NULL,
  title text,
  body text,
  PRIMARY KEY(id)
);

-- Migrate:DOWN
DROP TABLE post;
```

### Running Migrations

```bash
# Initialize the migrations schema table
miglite init

# Apply all pending migrations
miglite up
# Execute immediately without confirmation
miglite up --yes

# Rollback the most recent migration
miglite down
# Rollback multiple migrations
miglite down --number 3

# View migration status
miglite status
```

View migration status:

![status](./testdata/status.png)

## Using as a Library

`miglite` **does not depend on** any third-party DB driver libraries by itself, so you can use it as a library with your current database driver library.

- Sqlite drivers:
  - `modernc.org/sqlite` **CGO-free driver**
  - `github.com/glebarez/go-sqlite`  Based on `modernc.org/sqlite`
  - `github.com/ncruces/go-sqlite3` **CGO-free** Based on Wasm(wazero)
  - `github.com/mattn/go-sqlite3`  **NEED cgo**
- MySQL driver:
  - `github.com/go-sql-driver/mysql`
- Postgres driver:
  - `github.com/lib/pq`
  - `github.com/jackc/pgx/v5`
- MSSQL driver:
  - `github.com/microsoft/go-mssqldb`

> More drivers see: https://go.dev/wiki/SQLDrivers

```go
package main

import (
  "github.com/gookit/miglite"

  // add your database driver
  _ "github.com/go-sql-driver/mysql"
  // _ "github.com/lib/pq"
  // _ "modernc.org/sqlite"
)

func main() {
  mig, err := miglite.NewAuto(func(cfg *config.Config) {
    // update config options
  })
  goutil.PanicIfErr(err) // handle error

  // run up migrations
  err = mig.Up(command.UpOption{
    Yes: true, // dont confirm
    // ... options
  })
  goutil.PanicIfErr(err) // handle error

  // run down migrations ...
}
```

### Building Your Own Command Tool

You can directly use the `miglite` library to quickly build your own migration command tool, allowing you to register only the database drivers you need.

```go
package main

import (
	"github.com/gookit/miglite"
	"github.com/gookit/miglite/pkg/command"

	// add your database driver
	_ "github.com/go-sql-driver/mysql"
	// _ "github.com/lib/pq"
	// _ "modernc.org/sqlite"
)

var Version = "0.1.0"

func main() {
	// Optional: Information needs to be specified at build time via ldflags
	// command.SetBuildInfo(Version, GoVersion, BuildTime, GitCommit)

	// Create the CLI application
	app := command.NewApp("miglite", Version, "Lite database schema migration tool by Go")

	// Run the application
	app.Run()
}
```

> **NOTE**: If you want to further customize the CLI application, you can freely choose other CLI libraries, parse options, and then call the `handleXXX()` methods under `command` to execute the logic.

## Related Projects

- [golang-migrate](https://github.com/golang-migrate/migrate)
- [pressly/goose](https://github.com/pressly/goose)
- [amacneil/dbmate](https://github.com/amacneil/dbmate)
