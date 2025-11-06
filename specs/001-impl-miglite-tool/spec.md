# Feature Specification: Implement MigLite Tool

**Feature Branch**: `001-impl-miglite-tool`  
**Created**: 2025-11-05  
**Status**: Draft  
**Input**: User description: "依据 dev.md 的描述实现 miglite 工具"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Basic Migration Execution (Priority: P1)

用户需要能够运行数据库迁移脚本。用户将准备SQL迁移文件，指定数据库连接信息，然后使用miglite命令执行迁移。

**Why this priority**: 这是miglite工具的核心功能，没有这个功能，工具就没有实用价值。

**Independent Test**: 可以通过准备一个简单的SQL迁移文件，连接到数据库，然后运行miglite up命令来完全测试这个功能。成功执行后，数据库模式应该反映迁移的更改。

**Acceptance Scenarios**:

1. **Given** 用户有一个SQL迁移文件在migrations目录中, **When** 用户运行`miglite up`命令, **Then** 工具应连接到数据库并执行迁移，更新数据库模式
2. **Given** 数据库中已有迁移记录, **When** 用户运行`miglite status`命令, **Then** 工具应显示已应用和未应用的迁移列表

---

### User Story 2 - Migration Rollback (Priority: P2)

用户需要能够回滚已执行的数据库迁移。当迁移出现问题或需要撤销更改时，用户可以使用回滚功能。

**Why this priority**: 虽然不是最核心的功能，但在生产环境中处理错误迁移时至关重要。

**Independent Test**: 可以通过运行一个迁移，然后使用miglite down命令来测试。数据库应该回滚到迁移前的状态。

**Acceptance Scenarios**:

1. **Given** 已经应用了一个迁移, **When** 用户运行`miglite down`命令, **Then** 工具应回滚最近的迁移并更新数据库模式
2. **Given** 用户想要回滚特定的迁移, **When** 用户运行`miglite down --target <migration_name>`命令, **Then** 工具应回滚到目标迁移之前的状态

---

### User Story 3 - Migration Creation (Priority: P3)

用户需要能够创建新的迁移文件。工具应提供一个命令来生成符合miglite规范的新迁移文件模板。

**Why this priority**: 提高开发者工作效率，确保迁移文件格式正确。

**Independent Test**: 用户运行miglite create命令，应该生成一个带有正确时间戳和结构的SQL文件，包含-- Migrate:UP --和-- Migrate:DOWN --部分。

**Acceptance Scenarios**:

1. **Given** 用户需要创建新的迁移, **When** 用户运行`miglite create migration-name`命令, **Then** 工具应创建一个格式为YYYYMMDD-migration-name.sql的新文件
2. **Given** 新迁移文件被创建, **When** 用户查看文件内容, **Then** 文件应包含-- Migrate:UP --和-- Migrate:DOWN --模板

---

### Edge Cases

- 当数据库连接失败时会发生什么？工具应提供清晰的错误消息并优雅地退出。
- 如果迁移脚本包含语法错误会如何处理？工具应捕获错误，回滚事务中的更改，并提供错误上下文。
- 当多个用户同时尝试运行迁移时会发生什么？工具应具备适当的锁定机制防止冲突。
- 当迁移脚本被修改后会如何处理？工具应检测到已应用迁移的哈希变化，并提示用户。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 工具 MUST 支持MySQL、SQLite3和PostgreSQL数据库
- **FR-002**: 工具 MUST 从miglite.yaml配置文件读取数据库配置信息
- **FR-003**: 工具 MUST 从.env文件加载环境变量，特别是DATABASE_URL
- **FR-004**: 用户 MUST 能够运行`miglite up`命令来应用未执行的迁移
- **FR-005**: 用户 MUST 能够运行`miglite down`命令来回滚最近的迁移
- **FR-006**: 用户 MUST 能够运行`miglite status`命令来查看迁移状态
- **FR-007**: 用户 MUST 能够运行`miglite create <name>`命令来创建新的迁移文件
- **FR-008**: 工具 MUST 按照文件名（YYYYMMDD格式的时间戳）对迁移文件进行排序
- **FR-009**: 工具 MUST 解析迁移文件中的-- Migrate:UP --和-- Migrate:DOWN --部分
- **FR-010**: 工具 MUST 记录已执行的迁移到数据库中的db_schema_migrations表
- **FR-011**: 工具 MUST 基于标准Go database/sql包构建，避免使用特定的ORM
- **FR-012**: 工具 MUST 仅包含基本依赖，不强制添加数据库驱动依赖包

### Key Entities

- **Migration File**: 一个包含UP和DOWN SQL语句的文件，格式为YYYYMMDD-migration-name.sql，描述数据库模式的变更
- **Migration Record**: 数据库中db_schema_migrations表中的记录，追踪已执行的迁移，包含版本号、应用时间和状态
- **Configuration**: 包含数据库连接信息的miglite.yaml文件，定义数据库驱动、DSN和迁移路径

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 用户可以在3分钟内设置miglite并执行第一个迁移
- **SC-002**: 工具可以成功处理至少100个迁移文件的存储库
- **SC-003**: 95%的用户能够不参考文档就成功执行基本的迁移操作
- **SC-004**: 迁移执行的成功率达到99%以上（在有效的SQL语句和适当的数据库权限下）
- **SC-005**: 用户可以在1分钟内创建并配置一个新的迁移项目