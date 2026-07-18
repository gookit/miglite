# Multi-SQL Exec and Database Override Design

## Scope

Implement the two items in `.github/TODO.md`:

1. Allow `miglite exec` to execute multiple SQL statements.
2. Add the global `--db` option to override the target database name.

No new dependencies or full SQL dialect parser will be introduced.

## Multi-Statement Exec

Split input SQL on semicolons with a small stateful scanner. Semicolons inside
single-quoted strings, double-quoted identifiers, backtick identifiers, line
comments, and block comments are preserved. Empty statements are ignored.
Stored-procedure bodies and PostgreSQL dollar-quoted bodies are outside this
change.

After the existing confirmation prompt, `exec` starts one transaction and runs
the statements in order. Query statements (`SELECT`, `DESCRIBE`, `PRAGMA`, and
`SHOW`) print their result tables individually. Other statements print their
affected-row result individually. The first failure returns an error containing
the statement number and rolls back the transaction. All statements succeeding
commits the transaction.

The existing single-statement and SQL-file inputs remain unchanged at the CLI
surface.

## Database Override

Add `--db <name>` to the common command flags so it works before or after the
subcommand. Its value is applied after YAML and environment configuration is
loaded and before the database connection is created, giving it final priority.

For split connection settings, replace `Database.DBName` and rebuild the DSN.
For complete DSNs loaded from YAML, `DATABASE_DSN`, or `DATABASE_URL`, rewrite
the database component according to the configured driver. Support MySQL,
PostgreSQL, MSSQL, and SQLite. For SQLite, the value is the database file path.
An empty `--db` value makes no change.

## Tests

- Move the existing `testdrv` package to `cmd/miglite/testdrv` before feature
  work. Keep its package name, move its YAML fixture, update the migration path
  to `../../../testdata/migrations/{driver}`, and delete the standalone
  `testdrv/go.mod` and `testdrv/go.sum`. The migration integration test owns
  its SQLite connection instead of reusing the connection closed by the CLI
  `init` handler. The tests then share the CLI module's existing MySQL,
  PostgreSQL, and SQLite driver dependencies.
- Unit-test SQL splitting around quoted text, identifiers, comments, and empty
  statements.
- Use SQLite to verify multiple writes, individual query execution, commit, and
  rollback when a later statement fails.
- Table-test database override behavior for split configuration and supported
  DSN formats.
- Verify common CLI parsing accepts `--db` at the root and command levels.

Tests use `github.com/gookit/goutil/x/assert` as required by the project.

## Delivery

Deliver two focused commits:

1. `refactor(testdrv): share CLI module dependencies`
2. `feat(exec): support transactional multi-statement SQL`
3. `feat(cli): add database name override option`

The third commit also updates the English and Chinese README examples and
checks both completed items in `.github/TODO.md`.
