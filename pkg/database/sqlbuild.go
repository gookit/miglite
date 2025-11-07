package database

import (
	"fmt"

	"github.com/gookit/miglite/pkg/migutil"
)

const (
	DriverMySQL    = "mysql"
	DriverPostgres = "postgres"
	DriverSQLite   = "sqlite"
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
	QueryExists() string
	DeleteByVersion() string
}

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

// QueryExists 查询指定版本是否存在
func (b *SqlBuilderImpl) QueryExists() string {
	return "SELECT EXISTS(SELECT 1 FROM db_schema_migrations WHERE version = ?)"
}

// DeleteByVersion 删除指定版本
func (b *SqlBuilderImpl) DeleteByVersion() string {
	return "DELETE FROM db_schema_migrations WHERE version = ?"
}

type MySqlBuilder struct {
	SqlBuilderImpl
}

type PgSqlBuilder struct {
	SqlBuilderImpl
}

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
