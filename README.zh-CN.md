# miglite - lite database migration tool

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/miglite?style=flat-square)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/gookit/miglite)](https://github.com/gookit/miglite)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/miglite)](https://goreportcard.com/report/github.com/gookit/miglite)
[![Unit-Tests](https://github.com/gookit/miglite/workflows/Unit-Tests/badge.svg)](https://github.com/gookit/miglite/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/gookit/miglite.svg)](https://pkg.go.dev/github.com/gookit/miglite)

> **👉 [EN README](README.md)**

`miglite` Golang 实现的极简的数据库 Schema 迁移工具。

- 使用简单，极简依赖
- 基于 `database/sql` 进行开发，默认不添加任何驱动依赖包
- 迁移 SQL 都在事物中执行，确保数据一致性
- 基于原始 SQL 方式作为迁移文件
    - 固定文件名格式为 `YYYYMMDD-HHMMSS-{migration-name}.sql`
- 默认会递归搜索迁移目录下的所有sql文件(含子目录)
    - 查找sql文件时会忽略以 `_` 开始的目录(eg. `_backup/xx.sql`)
    - 迁移目录支持使用环境变量(eg `./migrations/${MODULE_NAME}`)
    - 迁移目录支持使用逗号 `,` 分割添加多个路径
- 可以通过环境变量零配置直接运行迁移(eg: `DATABASE_URL`, `MIGRATIONS_PATH`)
    - 会自动尝试加载目录下的 `.env` 文件(可选)
    - 会自动加载默认配置文件 `./miglite.yaml`(可选)
- 内置支持通过 `miglite exec` 执行 SQL 语句，方便调试和测试
- 支持 `mysql`, `sqlite`, `postgres` 数据库
    - 作为库使用时，需要自己添加DB驱动依赖
    - 直接使用 `miglite` 命令行工具时，已经添加了驱动依赖

## 安装

使用 `miglite` 命令行工具：

```bash
# install it by go
go install github.com/gookit/miglite/cmd/miglite@latest
```

作为Go依赖库使用：

```bash
go get github.com/gookit/miglite

# import "github.com/gookit/miglite"
```

## CLI直接使用

直接使用 `miglite` 命令行工具。

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

### 配置

`miglite` 支持通过 `miglite.yaml` 文件或环境变量进行配置。

- 可以允许没有配置文件，直接使用环境变量 `DATABASE_URL`
- 配置文件默认为 `./miglite.yaml`，也可以通过 `--config` 参数指定

#### miglite.yaml 示例

```yaml
database:
  driver: sqlite  # or mysql, postgresql
  dsn: ./miglite.db  # or connection string for other databases
migrations:
  path: ./migrations
```

#### 环境变量

- `MIGRATIONS_PATH`: 迁移文件所在目录路径 (默认: `./migrations`)
  - 支持使用环境变量(eg `./migrations/${MODULE_NAME}`)
  - 支持使用逗号 `,` 分割添加多个路径
- `DATABASE_URL`: 数据库连接 URL (例如: `sqlite://path/to/your.db`, `mysql://user:pass@localhost/dbname`)
- `MIGLITE_ENV_PREFIX`: 环境变量前缀, 默认为空字符串 (主要是用于快速的支持多DB迁移)
  - 设置后环境变量名会添加前缀读取，例如  `MIGLITE_ENV_PREFIX=MY_`, 则读取 `MY_DATABASE_URL` 而不是 `DATABASE_URL`
  - 上面的几个环境变量都受到ENV前缀设置影响

`ENV` 示例:

```ini
MIGRATIONS_PATH = "./migrations"
# sqlite
DATABASE_URL="sqlite://path/to/your.db"
# mysql
DATABASE_URL="mysql://user:passwd@tcp(127.0.0.1:3306)/local_test?charset=utf8mb4&parseTime=True&loc=Local"
# postgresql
DATABASE_URL="postgres://host=localhost port=5432 user=username password=password dbname=dbname sslmode=disable"
```

> **NOTE**: mysql 的 DSN 必须带上 `tcp(...)` 协议标记，否则会报错。

### 创建迁移

```bash
miglite create add-users-table
```

这将在 `./migrations/` 目录下创建一个以当前日期命名的 SQL 文件，格式为 `YYYYMMDD-HHMMSS-add-users-table.sql`。

```text
./migrations/20251105-102325-create-users-table.sql
```

SQL文件内容包含模板：

```sql
-- Migrate:UP
-- 在这里添加迁移 SQL

-- Migrate:DOWN
-- 在这里添加回滚 SQL (可选)
```

示例迁移文件：

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


### 运行迁移

```bash
# 初始化迁移表到DB
miglite init

# 应用所有待处理的迁移
miglite up
# 无需确认，立即执行
miglite up --yes

# 回滚最近的迁移
miglite down
# 回滚多个迁移
miglite down --number 3

# 查看迁移状态
miglite status
```

查看迁移状态:

![status](./testdata/status.png)

## 作为库使用

`miglite` 包本身**不依赖**任何三方DB驱动库，你可以将其作为库使用。搭配你当前的数据库驱动库使用。

- Sqlite 驱动:
    - `modernc.org/sqlite` **CGO-free driver**
    - `github.com/glebarez/go-sqlite`  基于 `modernc.org/sqlite` 封装
    - `github.com/ncruces/go-sqlite3` **CGO-free** Base on Wasm(wazero)
    - `github.com/mattn/go-sqlite3`  **NEED cgo**
- MySQL 驱动:
    - `github.com/go-sql-driver/mysql`
- Postgres 驱动:
    - `github.com/lib/pq`
    - `github.com/jackc/pgx/v5`
- MSSQL 驱动:
    - `github.com/microsoft/go-mssqldb`

> 更多驱动查看: https://go.dev/wiki/SQLDrivers

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

### 构建自己的命令工具

可以直接使用 `miglite` 库来快速构建自己的迁移命令工具，可以只注册自己需要的数据库驱动。

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
	// 可选：需要在构建时通过 ldflags 指定信息
	// command.SetBuildInfo(Version, GoVersion, BuildTime, GitCommit)

	// Create the CLI application
	app := command.NewApp("miglite", Version, "Lite database schema migration tool by Go")

	// Run the application
	app.Run()
}
```

> **NOTE**: 如果还要进一步自定义CLI应用，可以自由选择其他cli库，解析选项后调用 `command` 下面的 `handleXXX()` 方法执行逻辑。

## 相关的项目

- [golang-migrate](https://github.com/golang-migrate/migrate)
- [pressly/goose](https://github.com/pressly/goose)
- [amacneil/dbmate](https://github.com/amacneil/dbmate)


