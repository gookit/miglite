# MigLite

Go 极简的数据库迁移工具。

## 安装

```bash
# Clone the repository
git clone https://github.com/gookit/miglite.git
cd miglite

# Build the tool
go build -o miglite ./cmd/miglite

# Or install it
go install ./cmd/miglite
```

## 快速开始

```bash
miglite --help

Usage: miglite [command]
Commands:
  up      Run migrations
  down    Rollback migrations
  status  Show migrations status
  create  Create a new migration
```

## 配置

MigLite 支持通过 `miglite.yaml` 文件或环境变量进行配置。

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
