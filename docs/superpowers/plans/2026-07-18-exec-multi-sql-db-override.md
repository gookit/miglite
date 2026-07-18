# Multi-SQL Exec and Database Override Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Execute multiple SQL statements atomically with individual results, and let `--db` override the configured database for every CLI command.

**Architecture:** A dependency-free scanner splits common SQL safely enough for MySQL, PostgreSQL, and SQLite command usage. `HandleExec` owns one transaction and sends each statement through either `Query` or `Exec`. A config helper rewrites the final driver-specific DSN after file and environment configuration load.

**Tech Stack:** Go 1.21, `database/sql`, existing gookit packages, `github.com/gookit/goutil/x/assert`, and the existing SQLite test-driver module.

## Global Constraints

- No new dependency or complete SQL dialect parser.
- Preserve semicolons in quotes, identifiers, line comments, and block comments.
- Stored-procedure and PostgreSQL dollar-quoted bodies are out of scope.
- Run all statements in one transaction; any failure rolls everything back.
- Print every query and non-query result separately.
- `--db` overrides YAML, `DATABASE_DSN`, and `DATABASE_URL`.
- Support MySQL, PostgreSQL, MSSQL, and SQLite; for SQLite, `--db` is the file path.
- Code comments are English; tests use `github.com/gookit/goutil/x/assert`.

---

### Task 1: Transactional Multi-Statement Exec

**Files:**
- Create: `pkg/command/sql_split.go`
- Create: `pkg/command/sql_split_test.go`
- Create: `testdrv/exec_sqlite_test.go`
- Modify: `pkg/command/exec_cmd.go`
- Modify: this plan file for progress tracking

**Interfaces:**
- Produces: `splitSQLStatements(sqlText string) []string`
- Produces: `isQuerySQL(sqlText string) bool`
- Changes: `execQuery(queryer interface { Query(string, ...any) (*sql.Rows, error) }, sqlText string) error`

- [ ] **Step 1: Write failing scanner tests**

Create table-driven `TestSplitSQLStatements` cases with `assert.Eq` for:

```go
{"two statements", "CREATE TABLE users(id int); INSERT INTO users VALUES (1);",
 []string{"CREATE TABLE users(id int)", "INSERT INTO users VALUES (1)"}}
{"quoted semicolons", "INSERT INTO logs VALUES ('a;b', \"c;d\", `e;f`); SELECT 1;",
 []string{"INSERT INTO logs VALUES ('a;b', \"c;d\", `e;f`)", "SELECT 1"}}
{"comment semicolons", "-- keep ; here\nSELECT 1; /* keep ; here */ SELECT 2;",
 []string{"-- keep ; here\nSELECT 1", "/* keep ; here */ SELECT 2"}}
{"empty statements", "; SELECT 1;;", []string{"SELECT 1"}}
```

Add table-driven `TestIsQuerySQL` for leading whitespace/comments and `SELECT`, `DESCRIBE`, `PRAGMA`, `SHOW`, plus a non-query.

- [ ] **Step 2: Verify the tests fail**

Run from the repository root:

```powershell
go test ./pkg/command -run 'TestSplitSQLStatements|TestIsQuerySQL' -count=1
```

Expected: build failure because both functions are undefined.

- [ ] **Step 3: Implement the scanner**

Create `pkg/command/sql_split.go` with normal, single-quote, double-quote, backtick, line-comment, and block-comment states. Treat doubled quote characters and backslash-escaped bytes as content. Recognize `--` and `#` line comments and `/* ... */` block comments. Split only on a semicolon in normal state, trim with `strings.TrimSpace`, and omit empty statements.

Implement `isQuerySQL` by removing leading whitespace/comments, lowercasing the remaining prefix, and checking the four approved keywords with a word boundary.

- [ ] **Step 4: Verify scanner tests pass**

Run the Step 2 command again. Expected: `ok github.com/gookit/miglite/pkg/command`.

- [ ] **Step 5: Write failing SQLite transaction tests**

Create `testdrv/exec_sqlite_test.go` using `t.TempDir()`, the registered SQLite driver, `database.NewWithSqlDB`, `command.SetDB`, and `command.SetCfg`.

Success input:

```go
err := command.HandleExec(command.ExecOption{
    SQLOrFile: `CREATE TABLE items(id INTEGER PRIMARY KEY, name TEXT);
        INSERT INTO items(name) VALUES ('first');
        SELECT id, name FROM items;
        UPDATE items SET name = 'updated' WHERE id = 1;`,
    Yes: true,
})
```

Reopen the file and assert `name == "updated"`. A rollback subtest executes `CREATE TABLE`, a valid insert, then an insert into a missing table; reopen the file and assert querying `items` returns a missing-table error.

- [ ] **Step 6: Verify SQLite tests fail**

Run from `testdrv/`:

```powershell
go test ./... -run TestExecMultiSQL -count=1
```

Expected: current batch execution fails the atomic/per-statement assertions.

- [ ] **Step 7: Make `HandleExec` transactional**

In `pkg/command/exec_cmd.go`, define the small `queryer` interface and change `execQuery` to accept it. After the existing confirmation, split input and reject an empty result. Begin one transaction, defer rollback until a `committed` flag is true, and iterate with one-based statement numbers. Query statements call `execQuery(tx, statement)`; other statements call `tx.Exec(statement)` and print affected rows. Wrap execution errors as `failed to execute SQL statement %d: %w`. Commit after the loop.

- [ ] **Step 8: Verify Task 1**

Run from the root:

```powershell
gofmt -w pkg/command/exec_cmd.go pkg/command/sql_split.go pkg/command/sql_split_test.go
go test ./pkg/command -count=1
go test ./... -count=1
```

Run from `testdrv/`:

```powershell
gofmt -w exec_sqlite_test.go
go test ./... -run TestExecMultiSQL -count=1
```

Expected: all commands exit 0.

- [ ] **Step 9: Update progress and commit Task 1**

Change Task 1 checkboxes to `[x]`, then:

```powershell
git add pkg/command/exec_cmd.go pkg/command/sql_split.go pkg/command/sql_split_test.go testdrv/exec_sqlite_test.go docs/superpowers/plans/2026-07-18-exec-multi-sql-db-override.md
git commit -m "feat(exec): support transactional multi-statement SQL"
```

---

### Task 2: Global Database Override

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`
- Modify: `pkg/command/cliapp.go`
- Modify: `pkg/command/cliapp_test.go`
- Modify: `pkg/command/common.go`
- Modify: `README.md`, `README.zh-CN.md`, and `.github/TODO.md`
- Modify: this plan file for progress tracking

**Interfaces:**
- Produces: `config.OverrideDBName(dbCfg *Database, dbName string) error`
- Produces: package option variable `command.DBName string`
- Consumes: override in `initLoadConfig` before `OnConfigLoaded` and connection creation

- [ ] **Step 1: Write failing config tests**

Add `TestOverrideDBName` using `t.Run` and `assert` for:

```go
{"sqlite", Database{Driver: "sqlite", DSN: "old.db"}, "new.db", "new.db"}
{"mysql", Database{Driver: "mysql", DSN: "user:pass@tcp(localhost:3306)/old?parseTime=true"}, "new", "user:pass@tcp(localhost:3306)/new?parseTime=true"}
{"postgres keywords", Database{Driver: "postgres", DSN: "host=localhost user=user dbname=old sslmode=disable"}, "new", "host=localhost user=user dbname=new sslmode=disable"}
{"postgres URL", Database{Driver: "postgres", DSN: "postgres://user:pass@localhost/old?sslmode=disable"}, "new", "postgres://user:pass@localhost/new?sslmode=disable"}
{"postgres stripped URL", Database{Driver: "postgres", DSN: "user:pass@localhost/old?sslmode=disable"}, "new", "user:pass@localhost/new?sslmode=disable"}
{"mssql", Database{Driver: "mssql", DSN: "server=localhost;database=old;user id=sa;"}, "new", "server=localhost;database=new;user id=sa;"}
```

Assert both `DSN` and `DBName`. Add a split-config YAML test proving `Load` stores its generated DSN. Add malformed MySQL and unsupported-driver cases expecting errors.

- [ ] **Step 2: Verify config tests fail**

```powershell
go test ./internal/config -run TestOverrideDBName -count=1
```

Expected: build failure because `OverrideDBName` is undefined.

- [ ] **Step 3: Implement DSN rewriting**

In `checkDatabaseConfig`, assign `dbCfg.DSN = buildDSNFromConfig(dbCfg)` for split settings.

Add `OverrideDBName`: return for an empty name; SQLite assigns the file path; MySQL replaces the segment after the final slash and before query parameters; PostgreSQL uses `net/url` for URLs, path replacement for stripped URLs, and a case-insensitive `dbname` field replacement for keyword DSNs; MSSQL replaces the case-insensitive `database` field. Append PostgreSQL/MSSQL fields when absent. Set `DBName` only after a successful rewrite. Return descriptive errors for malformed DSNs or unsupported drivers. Use only `net/url`, `regexp`, and `strings`.

- [ ] **Step 4: Verify config tests pass**

```powershell
gofmt -w internal/config/config.go internal/config/config_test.go
go test ./internal/config -count=1
```

Expected: package passes.

- [ ] **Step 5: Write failing CLI tests**

Extend `pkg/command/cliapp_test.go` to reset global state in cleanup and verify:

```go
[]string{"--db", "root_db", "noop"}
[]string{"noop", "--db", "command_db"}
```

Both forms must populate `DBName`. Add `TestInitLoadConfigDBOverride`: set `DATABASE_URL=sqlite://old.db`, set `DBName="new.db"`, call `initLoadConfig`, and assert `Cfg().Database.DSN == "new.db"` before connection creation.

- [ ] **Step 6: Verify CLI tests fail**

```powershell
go test ./pkg/command -run 'Test.*DB.*Flag|TestInitLoadConfigDBOverride' -count=1
```

Expected: failures because `--db` and `DBName` do not exist.

- [ ] **Step 7: Bind and apply `--db`**

Add package variable `DBName string` in `pkg/command/cliapp.go` and bind it in `bindCommonFlags`:

```go
c.StringVar(&DBName, "db", "", "Override the configured database name")
```

Immediately after `config.Load` succeeds in `initLoadConfig`, before `OnConfigLoaded`, add:

```go
if err = config.OverrideDBName(&cfg.Database, DBName); err != nil {
    return fmt.Errorf("failed to override database name: %v", err)
}
```

- [ ] **Step 8: Verify Task 2 code**

```powershell
gofmt -w internal/config/config.go internal/config/config_test.go pkg/command/cliapp.go pkg/command/cliapp_test.go pkg/command/common.go
go test ./pkg/command ./internal/config -count=1
go test ./... -count=1
```

Expected: all commands exit 0.

- [ ] **Step 9: Update docs and TODO**

In both READMEs, document that `--db` overrides YAML/environment DSNs and SQLite treats it as a file path. Add:

```bash
miglite --db new_db status
miglite exec --db new_db --yes "SELECT current_database();"
```

Change both `.github/TODO.md` checkboxes from `[ ]` to `[x]`.

- [ ] **Step 10: Verify and commit Task 2**

Run root tests, then `go test ./... -count=1` from `testdrv/`, followed by `git diff --check`. Change Task 2 checkboxes to `[x]`, then:

```powershell
git add internal/config/config.go internal/config/config_test.go pkg/command/cliapp.go pkg/command/cliapp_test.go pkg/command/common.go README.md README.zh-CN.md .github/TODO.md docs/superpowers/plans/2026-07-18-exec-multi-sql-db-override.md
git commit -m "feat(cli): add database name override option"
```

---

### Task 3: Final Verification

**Files:** Verify only; update this plan's final checkboxes.

**Interfaces:** Consumes both feature commits and produces delivery evidence.

- [ ] **Step 1: Build the real CLI**

Run `go build ./...` from `cmd/miglite/`. Expected: exit 0.

- [ ] **Step 2: Run final checks**

From root run `go test ./... -count=1`, `git diff --check`, and `git status --short`. From `testdrv/` run `go test ./... -count=1`. Expected: tests pass and no unexpected worktree changes exist.

- [ ] **Step 3: Record final progress**

Change Task 3 checkboxes to `[x]`, stage this plan, and amend the Task 2 commit with `git commit --amend --no-edit`. Re-run `git status --short`; expected: clean worktree.
