# Migrator Runtime Architecture Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 `miglite.Migrator` 改造成持有独立配置和数据库生命周期的实例，同时保留 `pkg/command` 旧 API 并让其转发到无全局状态的 runtime。

**Architecture:** 新增 `internal/runtime`，显式持有 config、database 和预留的 `fs.FS`；`Migrator` 直接调用 runtime；`pkg/command` 保留 flag、交互、输出和旧 handler，通过 `legacyRuntime()` 兼容转发。Runtime 不依赖 command。

**Tech Stack:** Go、`database/sql`、现有 migration/database 包、`github.com/gookit/goutil/x/assert`。

## Global Constraints

- 不增加第三方依赖。
- 代码注释使用英文。
- 现有 `pkg/command` 公开 API 保留。
- 外部注入的 `*sql.DB` 默认由调用方关闭；Runtime 自己创建的连接由 Runtime 关闭。
- `internal/runtime` 不得导入 `pkg/command`。
- 本阶段不实现 embed FS，只保留 `fs.FS` 字段和接入边界。
- 旧 command API 仍是单例、顺序使用模型；不承诺其并发安全。

---

### Task 1: 固化数据库所有权和关闭契约

**Files:** `internal/database/database.go`; `pkg/command/common.go`; all handlers in `pkg/command/{init,up,down,skip,status,show,exec}_cmd.go`; tests in `pkg/command/cliapp_test.go`.

- [ ] 写失败测试：外部 DB 在 handler 后仍可用；内部创建 DB cleanup 后引用为 nil；`SetDB(nil)` 清状态；Ping 失败关闭连接。
- [ ] 运行 `go test ./pkg/command ./internal/database`，确认新测试失败。
- [ ] 增加私有 ownership 标记；`SetDB` 标记外部所有权；内部创建标记 owned；统一 cleanup 只关闭 owned 且将全局 db 置 nil。
- [ ] 替换所有 handler 的无条件 `defer db.SilentClose()`，包括 `HandleExec`。
- [ ] 在 `database.Connect` 的 Ping 失败路径关闭已打开连接。
- [ ] 运行 `go test ./pkg/command ./internal/database; go test ./...`。
- [ ] 提交 `fix(command): define database ownership lifecycle`。

### Task 2: 新增无全局状态的 Runtime

**Files:** create `internal/runtime/{runtime,options,init,up,down,skip,status,show,exec}.go`; test `internal/runtime/runtime_test.go`.

- [ ] 在 runtime 定义独立 option 类型，不能引用 `pkg/command`；覆盖 Init、Up、Down、Skip、Status、Show、Exec 当前使用的字段。
- [ ] 写 SQLite 失败测试，覆盖迁移执行、回滚、状态、跳过、Exec，以及外部注入 DB 不被关闭。
- [ ] 实现 `Runtime{cfg, db, fsys fs.FS, ownDB}`、构造、Close 和所有操作方法；业务逻辑从 command handler 下沉，保留现有错误和迁移行为。
- [ ] 运行 `go test ./internal/runtime -count=1`。
- [ ] 提交 `refactor(runtime): extract migration execution core`。

### Task 3: 让 Migrator 使用实例 Runtime

**Files:** `migrator.go`; `miglite.go`; `miglite_test.go`.

- [ ] 写失败测试：两个 Migrator 配置互不覆盖；外部 DB 不被关闭；`SetSqlDB(nil)`、重复 setter、重复 `Close`、Close 后操作行为明确。
- [ ] 给 Migrator 增加 cfg、runtime DB、reserved fsys 和 ownership；实现 `SetSqlDB(*sql.DB) *Migrator`、`SetFS(fs.FS) *Migrator`、幂等 `Close() error`。
- [ ] 令 `Init/Up/Down/Skip/Status/Show/Exec` 直接调用 runtime；在 Migrator 边界显式把 `command.*Option` 转为 runtime option，runtime 不导入 command。
- [ ] 运行 `go test . -count=1; go test ./internal/runtime ./pkg/migration -count=1`。
- [ ] 提交 `refactor(migrator): use instance runtime state`。

### Task 4: 将 command 改为兼容转发层

**Files:** `pkg/command/common.go`; all command handlers; `pkg/command/cliapp_test.go`.

- [ ] 写兼容测试：`SetCfg + SetDB + Handle*` 继续工作；重复 handler 不复用已关闭 owned DB；配置和 flag 在一次 legacy 调用中形成快照。
- [ ] 实现 `legacyRuntime()`：读取 legacy 状态，构造 runtime，区分外部/内部 DB ownership，并在所有错误路径 cleanup。
- [ ] 将 `HandleInit/Up/Down/Skip/Status/Show/Exec` 改为 option 转换后调用 runtime；保留签名并增加 deprecated 文档。
- [ ] 保留 CLI 的 flag、确认、输出、`OnConfigLoaded` 和环境同步行为在 command 边界处理。
- [ ] 运行 `go test ./pkg/command -count=1; go test ./... -count=1`。
- [ ] 提交 `refactor(command): forward legacy handlers to runtime`。

### Task 5: 文档和最终验证

**Files:** `README.md`; `README.zh-CN.md`; `docs/superpowers/specs/2026-08-29-migrator-runtime-architecture-design.md`; `miglite_test.go`.

- [ ] 记录 Migrator 顺序使用、SetSqlDB ownership、Close、legacy command 兼容和 Close 后行为。
- [ ] 明确 `SchemaTableName`、provider registry、环境前缀等仍为进程级全局，本阶段不宣称完全实例隔离。
- [ ] 更新示例和架构设计文档中的完成状态，不提前描述 embed FS 已实现。
- [ ] 运行 `go test ./... -count=1` 和 `go vet ./...`，确认通过。
- [ ] 提交 `docs(miglite): document runtime lifecycle and compatibility`。

## Self-review checklist

- Runtime 不导入 command，option 转换只发生在 Migrator/command 边界。
- `Exec` 纳入 runtime、Migrator 和 command 兼容路径。
- 所有 handler cleanup 遵守 DB ownership，外部 DB 永不被关闭。
- owned DB 关闭后引用清空，避免复用已关闭连接。
- nil setter、重复 Close、Close 后操作和重复 Migrator 已定义并测试。
- 剩余进程级全局状态明确为非目标。

