package database

import (
	"fmt"

	"github.com/gookit/miglite/pkg/migutil"
)

var sqlBuilders = map[string]SqlBuilder{
	"mysql":    &MySqlBuilder{},
	"postgres": &PgSqlBuilder{},
	"sqlite":   &SqliteBuilder{},
}

// AddSqlBuilder 添加数据库 SQL 语句构建处理
func AddSqlBuilder(driver string, builder SqlBuilder) {
	sqlBuilders[driver] = builder
}

// GetSqlBuilder 获取数据库 SQL 语句构建处理器
func GetSqlBuilder(driver string) (SqlBuilder, error) {
	name, err := migutil.ResolveDriver(driver)
	if err != nil {
		name = driver
	}

	builder, ok := sqlBuilders[name]
	if !ok {
		return nil, fmt.Errorf("unsupported OR un-registered database driver: %s", driver)
	}
	return builder, nil
}

// SqlBuilder 通用的数据库 SQL 语句构建处理
//   - ReSQL: mysql, postgres, sqlite3, oracle, mssql, ...
//   - NoSQL: MongoDB, Redis, ElasticSearch, ...
type SqlBuilder interface {
	CreateSchema() string
	DropSchema() string
	QueryAll() string
	// QueryOne by version. params: version
	QueryOne() string
	// QueryStatus 获取指定版本状态 params: version
	QueryStatus() string
	QueryExists() string
	DeleteByVersion() string
	// InsertMigration 插入迁移记录 params: version, status
	InsertMigration() string
	// UpdateMigration 更新迁移记录 params: status, version
	UpdateMigration() string
	// GetAppliedSortedByDate 获取所有已迁移的版本，按迁移时间排序 params: status, limit
	GetAppliedSortedByDate() string
}

//
// region default sql build
//

// SqlBuilderImpl 通用的关系型数据库 SQL 语句构建处理
type SqlBuilderImpl struct{}

// CreateSchema 创建数据库结构
func (b *SqlBuilderImpl) CreateSchema() string {
	return `
CREATE TABLE IF NOT EXISTS db_schema_migrations (
    version VARCHAR(160) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(24) -- up,skip,down
);`
}

// DropSchema 删除数据库结构
func (b *SqlBuilderImpl) DropSchema() string {
	return "DROP TABLE IF EXISTS db_schema_migrations"
}

// QueryAll 查询所有
func (b *SqlBuilderImpl) QueryAll() string {
	return "SELECT version, status, applied_at FROM db_schema_migrations"
}

// QueryOne 获取指定版本
func (b *SqlBuilderImpl) QueryOne() string {
	return "SELECT version, status, applied_at FROM db_schema_migrations WHERE version = ?"
}

// QueryStatus 查询指定版本状态
func (b *SqlBuilderImpl) QueryStatus() string {
	return "SELECT status FROM db_schema_migrations WHERE version = ?"
}

// QueryExists 查询指定版本是否存在
func (b *SqlBuilderImpl) QueryExists() string {
	return "SELECT EXISTS(SELECT 1 FROM db_schema_migrations WHERE version = ?)"
}

// DeleteByVersion 删除指定版本
func (b *SqlBuilderImpl) DeleteByVersion() string {
	return "DELETE FROM db_schema_migrations WHERE version = ?"
}

// InsertMigration 插入迁移记录
func (b *SqlBuilderImpl) InsertMigration() string {
	return "INSERT INTO db_schema_migrations (version, status) VALUES (?, ?)"
}

// UpdateMigration 更新迁移记录
func (b *SqlBuilderImpl) UpdateMigration() string {
	return "UPDATE db_schema_migrations SET applied_at = CURRENT_TIMESTAMP, status = ? WHERE version = ?"
}

// GetAppliedSortedByDate 获取所有已迁移的版本，按迁移时间排序
func (b *SqlBuilderImpl) GetAppliedSortedByDate() string {
	return "SELECT version, applied_at FROM db_schema_migrations WHERE status=? ORDER BY applied_at DESC LIMIT ?"
}

//
// region sql build for mysql
//

type MySqlBuilder struct {
	SqlBuilderImpl
}

//
// region sql build for sqlite
//

type SqliteBuilder struct {
	SqlBuilderImpl
}

// CreateSchema 创建数据库结构. sqlite 时间字段是 DATETIME
func (b *SqliteBuilder) CreateSchema() string {
	return `
CREATE TABLE IF NOT EXISTS db_schema_migrations (
    version VARCHAR(160) PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(24) -- up,skip,down
);`
}

//
// region sql build for pgsql
//

// PgSqlBuilder for postgresql
//
// NOTE: pgsql 绑定参数语法不一样，使用 $N
type PgSqlBuilder struct {
	SqlBuilderImpl
}

// QueryOne 获取指定版本
func (b *PgSqlBuilder) QueryOne() string {
	return "SELECT version, status, applied_at FROM db_schema_migrations WHERE version = $1"
}

// QueryStatus 查询指定版本状态
func (b *PgSqlBuilder) QueryStatus() string {
	return "SELECT status FROM db_schema_migrations WHERE version = $1"
}

// QueryExists 获取指定版本是否存在
func (b *PgSqlBuilder) QueryExists() string {
	return "SELECT EXISTS(SELECT 1 FROM db_schema_migrations WHERE version = $1)"
}

// DeleteByVersion 删除指定版本
func (b *PgSqlBuilder) DeleteByVersion() string {
	return "DELETE FROM db_schema_migrations WHERE version = $1"
}

// InsertMigration 插入迁移记录
func (b *PgSqlBuilder) InsertMigration() string {
	return "INSERT INTO db_schema_migrations (version, status) VALUES ($1, $2)"
}

// UpdateMigration 插入迁移记录
func (b *PgSqlBuilder) UpdateMigration() string {
	return "UPDATE db_schema_migrations SET applied_at = CURRENT_TIMESTAMP, status = $1 WHERE version = $2"
}

// GetAppliedSortedByDate 获取所有已迁移的版本，按迁移时间排序
func (b *PgSqlBuilder) GetAppliedSortedByDate() string {
	return "SELECT version, applied_at FROM db_schema_migrations WHERE status=$1 ORDER BY applied_at DESC LIMIT $2"
}
