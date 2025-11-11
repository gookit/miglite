package database

import (
	"fmt"

	"github.com/gookit/miglite/pkg/migutil"
)

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
	name, err := migutil.ResolveDriver(driver)
	if err != nil {
		name = driver
	}

	provider, ok := sqlProviders[name]
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
	return `
CREATE TABLE IF NOT EXISTS db_schema_migrations (
    version VARCHAR(160) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(24) -- up,skip,down
);`
}

// DropSchema 删除数据库结构
func (b *ReSqlProvider) DropSchema() string {
	return "DROP TABLE IF EXISTS db_schema_migrations"
}

// QueryAll 查询所有
func (b *ReSqlProvider) QueryAll() string {
	return "SELECT version, status, applied_at FROM db_schema_migrations"
}

// QueryOne 获取指定版本
func (b *ReSqlProvider) QueryOne() string {
	return "SELECT version, status, applied_at FROM db_schema_migrations WHERE version = ?"
}

// QueryStatus 查询指定版本状态
func (b *ReSqlProvider) QueryStatus() string {
	return "SELECT status FROM db_schema_migrations WHERE version = ?"
}

// QueryExists 查询指定版本是否存在
func (b *ReSqlProvider) QueryExists() string {
	return "SELECT EXISTS(SELECT 1 FROM db_schema_migrations WHERE version = ?)"
}

// DeleteByVersion 删除指定版本
func (b *ReSqlProvider) DeleteByVersion() string {
	return "DELETE FROM db_schema_migrations WHERE version = ?"
}

// InsertMigration 插入迁移记录
func (b *ReSqlProvider) InsertMigration() string {
	return "INSERT INTO db_schema_migrations (version, status) VALUES (?, ?)"
}

// UpdateMigration 更新迁移记录
func (b *ReSqlProvider) UpdateMigration() string {
	return "UPDATE db_schema_migrations SET applied_at = CURRENT_TIMESTAMP, status = ? WHERE version = ?"
}

// GetAppliedSortedByDate 获取所有已迁移的版本，按迁移时间排序
func (b *ReSqlProvider) GetAppliedSortedByDate() string {
	return "SELECT version, applied_at FROM db_schema_migrations WHERE status=? ORDER BY applied_at DESC LIMIT ?"
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
	return `
CREATE TABLE IF NOT EXISTS db_schema_migrations (
    version VARCHAR(160) PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(24) -- up,skip,down
);`
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
	return `
CREATE TABLE db_schema_migrations (
    version NVARCHAR(160) NOT NULL PRIMARY KEY,
    applied_at DATETIME2 DEFAULT CURRENT_TIMESTAMP,
    status NVARCHAR(24) -- up,skip,down
);`
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

// QueryOne 获取指定版本
func (b *PgSqlProvider) QueryOne() string {
	return "SELECT version, status, applied_at FROM db_schema_migrations WHERE version = $1"
}

// QueryStatus 查询指定版本状态
func (b *PgSqlProvider) QueryStatus() string {
	return "SELECT status FROM db_schema_migrations WHERE version = $1"
}

// QueryExists 获取指定版本是否存在
func (b *PgSqlProvider) QueryExists() string {
	return "SELECT EXISTS(SELECT 1 FROM db_schema_migrations WHERE version = $1)"
}

// DeleteByVersion 删除指定版本
func (b *PgSqlProvider) DeleteByVersion() string {
	return "DELETE FROM db_schema_migrations WHERE version = $1"
}

// InsertMigration 插入迁移记录
func (b *PgSqlProvider) InsertMigration() string {
	return "INSERT INTO db_schema_migrations (version, status) VALUES ($1, $2)"
}

// UpdateMigration 插入迁移记录
func (b *PgSqlProvider) UpdateMigration() string {
	return "UPDATE db_schema_migrations SET applied_at = CURRENT_TIMESTAMP, status = $1 WHERE version = $2"
}

// GetAppliedSortedByDate 获取所有已迁移的版本，按迁移时间排序
func (b *PgSqlProvider) GetAppliedSortedByDate() string {
	return "SELECT version, applied_at FROM db_schema_migrations WHERE status=$1 ORDER BY applied_at DESC LIMIT $2"
}
