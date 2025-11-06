# Data Model: MigLite Tool

## Migration File
- **Description**: 包含UP和DOWN SQL语句的文件，格式为YYYYMMDD-migration-name.sql，描述数据库模式的变更
- **Fields**:
  - fileName: string (格式: YYYYMMDD-migration-name.sql)
  - content: string (文件内容)
  - upMigration: string (由-- Migrate:UP --标记分隔的部分)
  - downMigration: string (由-- Migrate:DOWN --标记分隔的部分)
  - timestamp: string (从文件名解析的日期时间戳)
- **Validation Rules**:
  - 文件名必须遵循YYYYMMDD-migration-name.sql格式
  - 必须包含-- Migrate:UP --标记
  - 可选包含-- Migrate:DOWN --标记（用于回滚）
- **State Transitions**: 无

## Migration Record
- **Description**: 数据库中db_schema_migrations表中的记录，追踪已执行的迁移，包含版本号、应用时间和状态
- **Fields**:
  - version: string (迁移的版本号，对应文件名)
  - appliedAt: timestamp (迁移应用的时间)
  - status: string (迁移状态: up, skip, down)
  - hash: string (迁移内容的哈希，用于检测内容变更)
- **Validation Rules**:
  - version必须唯一
  - status必须是预定义值之一 (up, skip, down)
- **State Transitions**:
  - pending → up (当迁移成功应用)
  - up → down (当迁移回滚)
  - pending → skip (当迁移被跳过)

## Configuration
- **Description**: 包含数据库连接信息的miglite.yaml文件，定义数据库驱动、DSN和迁移路径
- **Fields**:
  - database.driver: string (数据库驱动: mysql, postgresql, sqlite3)
  - database.dsn: string (数据库连接字符串)
  - migrations.path: string (迁移文件的路径，默认: ./migrations)
  - database.url: string (从环境变量DATABASE_URL加载)
- **Validation Rules**:
  - 必须提供有效的数据库驱动
  - DSN必须符合所选数据库的格式
  - migrations.path必须指向有效的目录
- **State Transitions**: 无

## Relationships
- Migration File 与 Migration Record 之间是一对一的关系（通过version字段关联）
- Configuration 定义了 Migration Record 存储的数据库连接