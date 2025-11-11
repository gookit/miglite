# miglite - Lite database schema migration tool by Go

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/miglite?style=flat-square)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/gookit/miglite)](https://github.com/gookit/miglite)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/miglite)](https://goreportcard.com/report/github.com/gookit/miglite)
[![Unit-Tests](https://github.com/gookit/miglite/workflows/Unit-Tests/badge.svg)](https://github.com/gookit/miglite/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/gookit/miglite.svg)](https://pkg.go.dev/github.com/gookit/miglite)

> **[中文说明](README.zh-CN.md)**

`miglite` is a minimalist database schema migration tool implemented in Golang.

- Easy to use with minimal dependencies
- Developed based on `database/sql` without adding any driver dependencies by default
- Uses raw SQL files as migration files
  - SQL filename format: `YYYYMMDD-HHMMSS-{migration-name}.sql`
- Migration SQL is executed within transactions to ensure data consistency
- Can run migrations with zero configuration via environment variables (e.g., `DATABASE_URL`, `MIGRATIONS_PATH`)
  - Automatically attempts to load `.env` file in the directory
  - Automatically loads default configuration file `./miglite.yaml`
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

### Configuration

`miglite` supports configuration via `miglite.yaml` file or environment variables.

- Can work without a configuration file, using the environment variable `DATABASE_URL` directly
- Configuration file defaults to `./miglite.yaml`, but can be specified via the `--config` parameter

#### miglite.yaml Example

```yaml
database:
  driver: sqlite  # or mysql, postgresql
  dsn: ./miglite.db  # or connection string for other databases
migrations:
  path: ./migrations
```

#### Environment Variables

- `DATABASE_URL`: Database connection URL (e.g., `sqlite://path/to/db.sqlite`, `mysql://user:pass@localhost/dbname`)
- `MIGRATIONS_PATH`: Migration files path (default: `./migrations`)

### Creating Migrations

```bash
miglite create add-users-table
```

This will create an SQL file named with the current date in the `./migrations/` directory, with the format `YYYYMMDD-HHMMSS-add-users-table.sql`.

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

`miglite` does not depend on any third-party DB driver libraries by itself, so you can use it as a library with your current database driver library.

- Sqlite drivers:
  - `modernc.org/sqlite` **CGO-free driver**
  - `github.com/ncruces/go-sqlite3` **CGO-free** Based on Wasm(wazero)
  - `github.com/mattn/go-sqlite3`  **NEED cgo**
  - `github.com/glebarez/go-sqlite`  Based on `modernc.org/sqlite`
- MySQL driver:
  - `github.com/go-sql-driver/mysql`
- Postgres driver:
  - `github.com/lib/pq`
- MSSQL driver:
  - `github.com/microsoft/go-mssqldb`

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
	// miglite.InitInfo(Version, GoVersion, BuildTime, GitCommit)

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