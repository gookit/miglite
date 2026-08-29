# Migrator Runtime 架构调整设计

## 目标

`miglite` 最初以 `pkg/command` 为核心，后续增加的 `miglite.Migrator` 只是对
command handler 的薄封装，导致 Migrator 的配置和数据库实际存放在 command
包级变量中。本次调整将 `Migrator` 变成真正持有运行时状态的实例，同时保留
现有 `pkg/command` API 作为兼容入口。

本次不实现 embed FS，但为后续 `fs.FS` 接入预留实例边界。

## 现状问题

- `command.cfg` 和 `command.db` 是全局变量，`Migrator.cfg` 并不是实际配置来源。
- 创建第二个 `Migrator` 会覆盖第一个实例的 command 配置。
- command handler 无条件关闭数据库，可能关闭调用方通过 `SetSqlDB` 注入的
  `*sql.DB`。
- 数据库关闭后全局 `db` 仍非 nil，后续操作可能复用已关闭连接。
- 配置、数据库、flag 和环境变量的生命周期没有统一边界。
- `Migrator` 和 command 只能按单例、顺序方式使用，无法表达实例隔离。

独立 CLI 是单进程单次执行模型，上述全局状态基本可接受；库使用场景必须修复
配置、连接所有权和实例间覆盖问题。

## 目标架构

```text
pkg/command
    命令行 flag、交互确认、输出、旧 API 兼容
              ↓
internal/runtime
    显式持有 Config、DB、生命周期和（预留的）fs.FS
              ↓
pkg/migration + internal/database
    迁移发现、解析、执行、记录和数据库访问
```

`miglite.Migrator` 和 `pkg/command` 都调用 `internal/runtime`，互相不调用。
`internal/runtime` 不依赖 command，避免依赖倒置和 import cycle。

## Runtime 设计

新增无包级运行状态的内部运行时类型：

```go
type Runtime struct {
    cfg    *config.Config
    db     *database.DB
    fsys   fs.FS // reserved for embed FS
    ownDB  bool
}
```

Runtime 负责执行 `Init`、`Up`、`Down`、`Skip`、`Status`、`Show`、`Exec`，并调用现有
`pkg/migration`、`internal/database` 逻辑。它不负责 flag 解析、命令注册、确认
提示或全局状态。

迁移执行核心从 command handler 下沉到 runtime；command 只准备输入并转发。

## Migrator 设计

`Migrator` 保存自己的运行时状态：

```go
type Migrator struct {
    cfg    *Config
    db     *database.DB
    fsys   fs.FS // reserved for embed FS
    ownDB  bool
}
```

保留现有构造函数：

```go
func NewAuto(fns ...ConfigFn) (*Migrator, error)
func New(configFile string, fns ...ConfigFn) (*Migrator, error)
func NewWithConfig(cfg *Config) *Migrator
```

增加最小 setter：

```go
func (m *Migrator) SetSqlDB(db *sql.DB) *Migrator
func (m *Migrator) SetFS(fsys fs.FS) *Migrator // reserved, implementation later
func (m *Migrator) Close() error
```

不增加各种 DB/FS 组合构造函数；组合场景通过 setter 完成。

`Migrator.Init/Up/Down/Skip/Status/Show/Exec` 直接调用 runtime，不再调用
`command.Handle*`。Runtime 自己定义运行选项；`command.*Option` 只保留在
command 兼容层，handler 在转发时显式转换。Migrator 方法第一阶段可继续接受
`command.*Option` 以保持源码兼容，但不得让 runtime 依赖 command。

## 数据库所有权与生命周期

必须区分 Runtime/Migrator 自己创建的连接和调用方注入的连接。

### 自己创建的连接

- `ownDB = true`；
- `Close()` 负责关闭；
- 关闭后清空内部引用；
- CLI 兼容 handler 在命令结束时执行 cleanup。

### 调用方注入的连接

- `SetSqlDB` 设置 `ownDB = false`；
- `Migrator` 和 Runtime 不关闭调用方的 `*sql.DB`；
- 调用方负责连接生命周期。

操作方法 `Init`、`Up`、`Down`、`Skip`、`Status`、`Show`、`Exec` 等不隐式关闭 Migrator 持有的连接；关闭由
`Close()` 或明确的 CLI 生命周期负责。

## command 兼容层

继续保留以下公开 API：

```go
SetCfg
Cfg
SetDB
DB
HandleInit
HandleUp
HandleDown
HandleSkip
HandleStatus
HandleShow
```

这些 API 标记为兼容入口，内部状态可重命名为 `legacyCfg`、`legacyDB`，避免
新代码误用。

每个 `Handle*`（包括 `HandleExec`）通过统一的 `legacyRuntime()` 构造 Runtime：

1. 读取旧 API 设置的配置和数据库；
2. 没有配置时按原有逻辑加载配置；
3. 没有数据库时创建连接并标记 `ownDB = true`；
4. 执行 Runtime 对应方法；
5. 只清理 Runtime 自己拥有的数据库连接；清理后将 legacy 全局 db 置 nil，避免复用已关闭连接。

这样旧代码继续有效：

```go
command.SetCfg(cfg)
command.SetDB(db)
return command.HandleUp(opt)
```

官方 CLI 仍然使用 command 的 flag 和 handler，但实际业务逻辑已经走 Runtime。

## 配置边界

- `Migrator.cfg` 是 Migrator 实例唯一的配置来源；
- `Runtime.cfg` 是当前执行唯一的配置来源；
- `command` 全局配置只存在于兼容模式；
- legacy handler 创建 Runtime 时一次性形成最终配置，不能混用旧的全局 flag、
  环境变量和已缓存对象；
- `SetCfg(nil)`、`SetDB(nil)`、重复 `SetSqlDB`、`Close` 后再次执行和重复 `Close`
  的行为必须在 API 契约中定义并测试；推荐 nil 返回明确错误，Close 幂等。
- 本次不把 `fs.FS` 放进 YAML 或环境变量。

## 分阶段实施

### 阶段一：修复 DB 生命周期

- 区分内部创建连接与外部注入连接；
- 外部 `SetSqlDB` 连接不再被 handler 关闭；
- 自动创建连接关闭后清空引用；
- 修改全部 handler（包括 Exec）的 cleanup 路径；
- `database.Connect` 在 Ping 失败时关闭已打开的 `*sql.DB`，避免泄漏；
- 保持现有 command API 和执行路径。

### 阶段二：抽离 Runtime

- 新增 `internal/runtime`；
- 迁移 `HandleInit/Up/Down/Skip/Status/Show/Exec` 的业务逻辑；
- `Migrator` 直接调用 Runtime；
- command handler 改为构造 Runtime 后转发。

### 阶段三：收口兼容层

- 统一 `legacyRuntime()`；
- 为旧 setter 和 handler 增加 deprecated 文档；
- 保证官方 CLI 行为不变；
- 明确旧 command API 仍是单例、顺序使用模型。

### 阶段四：接入 embed FS

- Runtime/Migrator 使用实例级 `fs.FS`；
- `pkg/migration` 增加显式 FS API；
- command 增加可选的兼容绑定；
- 增加 `embed.FS` 和 `fstest.MapFS` 测试。

## 测试要求

- 两个 Migrator 的配置互不覆盖；
- 外部注入的 `*sql.DB` 在操作完成后仍可使用；
- Runtime 自己创建的连接会正确关闭且不会复用已关闭引用；
- 旧 `command.SetCfg/SetDB/Handle*` API 继续工作；
- command 和 Migrator 对同一 SQLite 数据库产生一致结果；
- 重复调用同一 Migrator 的生命周期行为明确且有测试；
- 本阶段暂不承诺旧 command API 的并发安全。
- `database.SchemaTableName`、SQL provider registry、`config.EnvPrefix` 等现有
  进程级全局扩展点本阶段不实例化；文档明确它们仍要求进程级配置，不能据此
  宣称多个 Migrator 完全并发隔离。

## 非目标

- 不立即删除或重命名现有 command 公开 API；
- 不在本阶段实现 embed FS；
- 不重写 SQL 解析、数据库 provider 或迁移记录模型；
- 不一次性把 CLI 重构成全新的命令框架；
- 不在本阶段重构 schema table name、provider registry 或环境变量为实例状态；
- 不为尚无需求的并发执行增加复杂锁或全局调度器。

## 兼容和迁移策略

现有 CLI 无需修改使用方式。库使用方可以继续调用 `miglite.New` 和
`Migrator` 方法；行为变化仅包括：外部注入的数据库不再被自动关闭，以及不同
`Migrator` 实例不再互相覆盖配置。直接依赖 `command.Set*` 的旧代码继续运行，
但建议迁移到 `miglite.Migrator`。
