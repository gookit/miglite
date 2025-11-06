# Research: Implement MigLite Tool

## Overview
This research document addresses unknowns and decisions required for implementing the MigLite tool as specified in the feature specification.

## Decision: Go Version and Project Structure
- **Rationale**: 项目根目录的go.mod文件显示使用Go 1.19，所以我们应该遵循此版本要求。项目结构将采用Go项目的标准模式。
- **Alternatives considered**: 
  - 使用更高版本的Go (20+): 这可能会与现有go.mod不兼容
  - 不同的目录结构: 标准的Go项目结构是最佳实践

## Decision: CLI Framework
- **Rationale**: 根据dev.md，我们将使用`github.com/gookit/goutil/cflag`来构建命令行工具，这是项目已经指定的依赖。
- **Alternatives considered**:
  - Cobra: 更流行的Go CLI库，但dev.md明确指定了cflag
  - Pflag: Go标准的flags包，功能可能不够

## Decision: Database Driver Dependencies
- **Rationale**: 根据dev.md，我们不强制添加数据库驱动依赖包，而是让用户自己添加MySQL、SQLite3、PostgreSQL的依赖。
- **Implementation**: 工具将使用标准的database/sql接口，但不会引入具体驱动，用户需要在项目中引入所需的驱动
- **Alternatives considered**: 
  - 强制包含所有常见驱动: 这会违反宪法的极简主义原则
  - 使用特定ORM: 这违反了dev.md和宪法中关于使用标准database/sql接口的要求

## Decision: Configuration Management
- **Rationale**: 根据dev.md，工具需要支持miglite.yaml配置文件和.env环境变量加载，特别是DATABASE_URL。
- **Implementation**: 将实现两个配置模块，一个用于YAML解析，一个用于环境变量加载
- **Alternatives considered**: 其他配置格式（如JSON、TOML）但dev.md指定了YAML

## Decision: Migration File Format and Parsing
- **Rationale**: 根据dev.md，迁移文件必须遵循YYYYMMDD-migration-name.sql格式，内容通过-- Migrate:UP --和-- Migrate:DOWN --注释进行解析。
- **Implementation**: 将实现文件名解析器和SQL内容解析器
- **Alternatives considered**: 其他格式或解析方式，但dev.md明确指定了这种格式

## Decision: Migration Tracking System
- **Rationale**: 根据dev.md，需要在数据库中创建db_schema_migrations表来记录已执行的迁移。我们已经看到在main.go中有创建这个表的SQL语句。
- **Implementation**: 将实现迁移状态跟踪逻辑，使用db_schema_migrations表
- **Alternatives considered**: 文件系统存储、其他格式的跟踪，但数据库表是最常见的方法

## Decision: Migration Execution Commands
- **Rationale**: 根据spec.md，需要支持up、down、status、create命令。
- **Implementation**: 将实现这些命令的逻辑
- **Alternatives considered**: 其他命令名称或组织方式，但这些是标准的迁移工具命令

## Key Technical Challenges Identified
1. 并发访问控制 - 当多个用户尝试同时运行迁移时需要处理
2. 事务处理 - 每个迁移应该在事务中执行
3. 错误处理 - 当迁移失败时如何回滚和报告错误
4. 数据库锁定 - 防止多个实例同时运行迁移
5. 向下兼容性 - 确保对已应用迁移的更改被正确处理