# MigLite

`miglite` Go实现的极简的数据库迁移工具。

- 配置文件默认为 `./miglite.yaml`
- 基于 `database/sql` 进行开发，默认不添加任何驱动依赖包
- 支持 mysql, sqlite3, postgresql. 需要自己添加驱动依赖
- 迁移 SQL 都在事物中执行，确保数据一致性

## 安装

```bash
# install it by go
go install github.com/gookit/miglite/cmd/miglite@latest
```

## 快速开始

```bash
miglite - Go minimal database migration tool(Version: 0.0.1)

Usage: miglite COMMAND [--Options...] [...Arguments]
Options:
  -h, --help                Display application help
  --version, -v             Show version and exit

Commands:
  create          Create a new migration
  down            Rollback the most recent migration
  init            Initialize the migration schema on db
  status          Show the status of migrations
  up              Execute pending migrations
  help            Display application help

Use "miglite COMMAND --help" for about a command
```

## 配置

`miglite` 支持通过 `miglite.yaml` 文件或环境变量进行配置。

### miglite.yaml 示例

```yaml
database:
  driver: sqlite3  # or mysql, postgresql
  dsn: ./miglite.db  # or connection string for other databases
migrations:
  path: ./migrations
```

### 环境变量

- `DATABASE_URL`: 数据库连接 URL (例如: `sqlite://path/to/db.sqlite`, `mysql://user:pass@localhost/dbname`)
- `MIGRATIONS_PATH`: 迁移文件路径 (默认: `./migrations`)

## 创建迁移

```bash
miglite create add-users-table
```

这将在 `./migrations/` 目录下创建一个以当前日期命名的 SQL 文件，格式为 `YYYYMMDD-add-users-table.sql`。

文件内容包含模板：
```sql
-- Migrate:UP
-- 在这里添加迁移 SQL

-- Migrate:DOWN
-- 在这里添加回滚 SQL (可选)
```

## 运行迁移

```bash
# 应用所有待处理的迁移
miglite up

# 回滚最近的迁移
miglite down

# 回滚多个迁移
miglite down --count 3

# 查看迁移状态
miglite status
```

## 作为库使用

`miglite` 本身不依赖任何三方驱动库，你可以将其作为库使用。搭配你当前的数据库驱动库使用。

- Sqlite 驱动:
  - `modernc.org/sqlite` **CGO-free driver**
  - `github.com/ncruces/go-sqlite3` **CGO-free** Base on Wasm(wazero)
  - `github.com/mattn/go-sqlite3`  **NEED cgo**
  - `github.com/glebarez/go-sqlite`  Base on `modernc.org/sqlite`
- MySQL 驱动:
  - github.com/go-sql-driver/mysql
- PostgreSQL 驱动:
  - github.com/lib/pq

```go

```
