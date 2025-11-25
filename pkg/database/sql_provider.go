package database

import (
	"fmt"

	"github.com/gookit/miglite/pkg/migutil"
)

// SchemaTableName 默认数据库迁移记录表名
var SchemaTableName = "schema_migrations"

// 内置SQL语句提供者适配
var sqlProviders = map[string]SqlProvider{
	"mssql":    &MSSqlProvider{},
	"mysql":    &MySqlProvider{},
	"postgres": &PgSqlProvider{},
	"sqlite":   &SqliteProvider{},
}

// AddProvider 添加数据库 SQL 语句提供者
func AddProvider(driver string, provider SqlProvider) {
	sqlProviders[driver] = provider
}

// GetSqlProvider 获取数据库 SQL 语句提供者
func GetSqlProvider(driver string) (SqlProvider, error) {
	fmtName := migutil.FmtDriverName(driver)
	provider, ok := sqlProviders[fmtName]
	if !ok {
		return nil, fmt.Errorf("unsupported OR un-registered database driver: %s", driver)
	}
	return provider, nil
}

// SqlProvider 通用的数据库 SQL 语句提供者
//   - ReSQL: mysql, postgres, sqlite3, oracle, mssql, ...
//   - NoSQL: MongoDB, Redis, ElasticSearch, ...
type SqlProvider interface {
	// CreateSchema 创建数据库结构SQL
	CreateSchema() string
	DropSchema() string
	ShowTables() string
	// QueryTableSchema 获取数据库表结构SQL
	QueryTableSchema(tableName string) string

	QueryAll() string
	// QueryOne by version. params: version
	QueryOne() string
	// QueryStatus 获取指定版本状态 params: version
	QueryStatus() string
	// QueryExists 获取指定版本是否存在 params: version
	QueryExists() string
	// InsertMigration 插入迁移记录 params: version, status
	InsertMigration() string
	// UpdateMigration 更新迁移记录 params: status, version
	UpdateMigration() string
	// GetAppliedSortedByDate 获取所有已迁移的版本，按迁移时间排序 params: status, limit
	GetAppliedSortedByDate() string
	// DeleteByVersion() string
}

//
// region Re-Sql Provider
//

// ReSqlProvider 通用的关系型数据库 SQL 语句提供者
type ReSqlProvider struct{}

// CreateSchema 创建数据库结构
func (b *ReSqlProvider) CreateSchema() string {
	return "CREATE TABLE IF NOT EXISTS " + SchemaTableName + ` (
    version VARCHAR(160) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(24) -- up,skip,down
);`
}

// DropSchema 删除数据库结构
func (b *ReSqlProvider) DropSchema() string {
	return "DROP TABLE IF EXISTS " + SchemaTableName
}

// ShowTables 显示所有表
func (b *ReSqlProvider) ShowTables() string { return "SHOW TABLES" }

// QueryTableSchema 获取数据库表结构
func (b *ReSqlProvider) QueryTableSchema(tableName string) string {
	return fmt.Sprintf("DESCRIBE `%s`", tableName)
}

// QueryAll 查询所有
func (b *ReSqlProvider) QueryAll() string {
	return "SELECT version, status, applied_at FROM " + SchemaTableName
}

// QueryOne 获取指定版本
func (b *ReSqlProvider) QueryOne() string {
	return "SELECT version, status, applied_at FROM " + SchemaTableName + " WHERE version = ?"
}

// QueryStatus 查询指定版本状态
func (b *ReSqlProvider) QueryStatus() string {
	return "SELECT status FROM " + SchemaTableName + " WHERE version = ?"
}

// QueryExists 查询指定版本是否存在
func (b *ReSqlProvider) QueryExists() string {
	return "SELECT EXISTS(SELECT 1 FROM " + SchemaTableName + " WHERE version = ?)"
}

// DeleteByVersion 删除指定版本
func (b *ReSqlProvider) DeleteByVersion() string {
	return "DELETE FROM " + SchemaTableName + " WHERE version = ?"
}

// InsertMigration 插入迁移记录
func (b *ReSqlProvider) InsertMigration() string {
	return "INSERT INTO " + SchemaTableName + " (version, status) VALUES (?, ?)"
}

// UpdateMigration 更新迁移记录
func (b *ReSqlProvider) UpdateMigration() string {
	return "UPDATE " + SchemaTableName + " SET applied_at = CURRENT_TIMESTAMP, status = ? WHERE version = ?"
}

// GetAppliedSortedByDate 获取所有已迁移的版本，按迁移时间排序
func (b *ReSqlProvider) GetAppliedSortedByDate() string {
	return "SELECT version, applied_at FROM " + SchemaTableName + " WHERE status=? ORDER BY applied_at DESC LIMIT ?"
}

//
// region MySql Provider
//

// MySqlProvider for mysql
type MySqlProvider struct {
	ReSqlProvider
}

//
// region Sqlite Provider
//

// SqliteProvider for sqlite
type SqliteProvider struct {
	ReSqlProvider
}

// CreateSchema 创建数据库结构. sqlite 时间字段是 DATETIME
func (b *SqliteProvider) CreateSchema() string {
	return "CREATE TABLE IF NOT EXISTS " + SchemaTableName + `(
    version VARCHAR(160) PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(24) -- up,skip,down
);`
}

// ShowTables 显示所有表
func (b *SqliteProvider) ShowTables() string {
	return "SELECT name FROM sqlite_master WHERE type='table'"
}

// QueryTableSchema 获取数据库表结构
func (b *SqliteProvider) QueryTableSchema(tableName string) string {
	return fmt.Sprintf("PRAGMA table_info(`%s`)", tableName)
}

//
// region MsSql Provider
//

// MSSqlProvider for mssql
type MSSqlProvider struct {
	ReSqlProvider
}

// CreateSchema 创建数据库结构. mssql 使用 DATETIME2 和 IDENTITY
func (b *MSSqlProvider) CreateSchema() string {
	return "CREATE TABLE " + SchemaTableName + `(
    version NVARCHAR(160) NOT NULL PRIMARY KEY,
    applied_at DATETIME2 DEFAULT CURRENT_TIMESTAMP,
    status NVARCHAR(24) -- up,skip,down
);`
}

// ShowTables 显示所有表
func (b *MSSqlProvider) ShowTables() string {
	return `SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE = 'BASE TABLE'`
}

// QueryTableSchema 获取数据库表结构
func (b *MSSqlProvider) QueryTableSchema(tableName string) string {
	return fmt.Sprintf(`
SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT,
	COLUMNPROPERTY(OBJECT_ID(TABLE_SCHEMA+'.'+TABLE_NAME), COLUMN_NAME, 'IsIdentity') AS IS_IDENTITY,
	'' AS EXTRA
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_NAME = '%s'
ORDER BY ORDINAL_POSITION`, tableName)
}

//
// region PgSql Provider
//

// PgSqlProvider postgres sql 语句提供者
//
// NOTE: pgsql 绑定参数语法不一样，使用 $N
type PgSqlProvider struct {
	ReSqlProvider
}

// ShowTables 显示所有表
func (b *PgSqlProvider) ShowTables() string {
	return `SELECT tablename FROM pg_tables WHERE schemaname = 'public'`
}

// QueryTableSchema 获取数据库表结构
func (b *PgSqlProvider) QueryTableSchema(tableName string) string {
	return fmt.Sprintf(`
SELECT column_name, data_type, is_nullable, column_default 
FROM information_schema.columns 
WHERE table_name = '%s' 
ORDER BY ordinal_position`, tableName)
}

// QueryOne 获取指定版本
func (b *PgSqlProvider) QueryOne() string {
	return "SELECT version, status, applied_at FROM " + SchemaTableName + " WHERE version = $1"
}

// QueryStatus 查询指定版本状态
func (b *PgSqlProvider) QueryStatus() string {
	return "SELECT status FROM " + SchemaTableName + " WHERE version = $1"
}

// QueryExists 获取指定版本是否存在
func (b *PgSqlProvider) QueryExists() string {
	return "SELECT EXISTS(SELECT 1 FROM " + SchemaTableName + " WHERE version = $1)"
}

// DeleteByVersion 删除指定版本
func (b *PgSqlProvider) DeleteByVersion() string {
	return "DELETE FROM " + SchemaTableName + " WHERE version = $1"
}

// InsertMigration 插入迁移记录
func (b *PgSqlProvider) InsertMigration() string {
	return "INSERT INTO " + SchemaTableName + " (version, status) VALUES ($1, $2)"
}

// UpdateMigration 插入迁移记录
func (b *PgSqlProvider) UpdateMigration() string {
	return "UPDATE " + SchemaTableName + " SET applied_at = CURRENT_TIMESTAMP, status = $1 WHERE version = $2"
}

// GetAppliedSortedByDate 获取所有已迁移的版本，按迁移时间排序
func (b *PgSqlProvider) GetAppliedSortedByDate() string {
	return "SELECT version, applied_at FROM " + SchemaTableName + " WHERE status=$1 ORDER BY applied_at DESC LIMIT $2"
}
