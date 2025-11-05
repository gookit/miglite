package main

// PgSqlInit PostgreSQL数据库初始化脚本
const PgSqlInit = `
CREATE TABLE IF NOT EXISTS db_schema_migrations (
    version VARCHAR(160) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(24) -- up,skip,down
);
`

// MySqlInit MySQL数据库初始化脚本
const MySqlInit = `
CREATE TABLE IF NOT EXISTS db_schema_migrations (
    version VARCHAR(160) PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(24) -- up,skip,down
);
`

// SQLiteInit SQLite数据库初始化脚本
const SQLiteInit = `
CREATE TABLE
`;


func main() {

}
