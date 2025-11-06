# DEV

Go 极简的数据库迁移工具。

- 默认不添加任何驱动依赖包
- 支持 mysql, sqlite3, postgresql 需要自己添加依赖
- 配置文件为 `miglite.yaml`

## Dev

- 只基于 `database/sql` 进行开发
- 使用 `github.com/gookit/goutil/cflag` 构建命令行工具
- 使用 `github.com/gookit/goutil/testutil/assert` 编写断言测试
- 每个迁移是一个 SQL 文件 参考 [20251023-user-add-field.sql](testdata/20251023-user-add-field.sql)
  - 根据文件名排序执行 推荐格式 `YYYYMMDD-migration-name.sql`
  - 内容通过固定的 `-- Migrate:UP` `-- Migrate:DOWN` 格式进行解析
  - 默认放在 `./migrations` 目录下
- 迁移 SQL 都在事物中执行，确保数据一致性

## Usage

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

### `.env` 文件

默认会读取 `.env` 文件加载环境变量

```ini
# 会根据 DATABASE_URL 环境变量自动识别数据库驱动
DATABASE_URL=postgres://pguser1:pg1234abcd@localhost:5432/app_upgrade_admin
MIGRATIONS_PATH=./migrations
```

### `miglite.yaml` 配置

通过 `miglite.yaml` 可以配置更多细节选项。

```yaml
database:
  driver: mysql
  dsn: root:root@tcp(127.0.0.1:3306)/miglite?charset=utf8mb4&parseTime=True&loc=Local
migrations:
  path: ./migrations
```