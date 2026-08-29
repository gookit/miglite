# Embed FS 支持设计

## 目标

允许库使用方将 SQL 迁移文件通过 Go `embed.FS` 打包进自己的程序，并继续使用
`miglite.Migrator` 或复用 `pkg/command` 的命令处理逻辑。现有 CLI 的磁盘目录
使用方式保持兼容。

## 现状与约束

- `pkg/migration` 当前通过 `os.ReadFile`、`fsutil.FindInDir` 和
  `filepath.WalkDir` 读取本地文件。
- `pkg/command` 使用包级 `cfg` 和 `db`，命令处理函数没有接收 `Migrator` 实例。
- `miglite.NewWithConfig` 会把配置和数据库同步到 `command` 的全局状态，因此
  `Migrator` 当前已经是顺序使用模型，不承诺多个实例并发执行。
- `embed.FS` 遵循 `io/fs` 逻辑路径，路径分隔符固定为 `/`，不能使用 Windows
  本地路径规则。

## 设计方案

### 1. 迁移包提供显式 FS API

新增以下函数，底层不保存全局 `fs.FS`：

```go
func ParseFS(fsys fs.FS, filePath string) (*Migration, error)

func FindMigrationsFS(fsys fs.FS, migrationsDir string, recursive bool) ([]*Migration, error)

func MigrationsFromFS(fsys fs.FS, migPath string, files []string) ([]*Migration, error)
```

FS API 使用 `fs.ReadFile`、`fs.ReadDir`/`fs.WalkDir` 和 `path.Base`。解析后的
`Migration.FilePath` 保存 FS 逻辑路径，例如
`migrations/20260829-120000-create-users.sql`。

现有 `ParseFile`、`FindMigrations` 和 `MigrationsFrom` 保持原行为，继续读取磁盘，
避免破坏已有调用方。

### 2. Migration 解析复用

把“根据文件名创建 `Migration`”与“从内容解析 UP/DOWN”保留为共享逻辑：

1. FS API 校验并解析逻辑文件名；
2. 读取文件内容；
3. 设置 `Migration.Contents`；
4. 调用现有 `ParseContents()`。

不向 `Migration` 增加来源字段，也不让 `Parse()` 隐式判断磁盘或 FS，避免扩大
结构和状态复杂度。

### 3. command 层提供可选绑定

`pkg/command` 增加进程级可选绑定：

```go
func SetMigrationFS(fsys fs.FS)
```

`nil` 表示恢复现有磁盘模式。`command.findMigrations()` 根据绑定状态选择
`FindMigrationsFS` 或 `FindMigrations`；`skip` 命令同样选择
`MigrationsFromFS` 或 `MigrationsFrom`，不能只改 `up/status`。

该变量只属于 command 的兼容适配层，`pkg/migration` 不持有全局 FS。

### 4. Migrator 实例适配

`Migrator` 增加实例字段：

```go
type Migrator struct {
    cfg  *Config
    fsys fs.FS
}
```

新增：

```go
func NewWithConfigAndFS(cfg *Config, fsys fs.FS) *Migrator
func (m *Migrator) SetFS(fsys fs.FS) *Migrator
```

`Init`、`Up`、`Down`、`Skip`、`Status`、`Show` 每次调用命令处理函数前，先将
`m.fsys` 绑定到 command。这样同一个 `Migrator` 可在磁盘模式和 FS 模式之间
明确切换；`nil` 保持兼容。

根包可提供便捷入口：

```go
//go:embed migrations/*.sql
var migrationFS embed.FS

m, err := miglite.New("", func(cfg *miglite.Config) {
    cfg.Migrations.Path = "migrations"
})
if err != nil { return err }
m.SetFS(migrationFS).SetSqlDB(db)
return m.Up(command.UpOption{})
```

不把 `fs.FS` 写入 YAML 或环境配置；FS 是运行时依赖，不是可序列化配置。

### 5. 复用 command 的场景

需要直接构建命令应用时，根包提供可选便捷转发：

```go
func BindFS(fsys fs.FS) { command.SetMigrationFS(fsys) }
```

调用方在 `app.Run()` 前调用 `miglite.BindFS(migrationFS)`，CLI 仍可不调用该
函数而使用磁盘目录。该入口主要服务于不使用 `Migrator`、直接复用 command 的
应用；优先推荐库使用方采用实例级 `Migrator.SetFS`。

## 行为与错误处理

- FS 路径必须是合法 `io/fs` 相对路径；绝对路径和反斜杠路径由 `fs` API 返回错误。
- 仅处理 `.sql` 文件；忽略 `_` 开头的文件和目录；递归开关与磁盘模式一致。
- 文件不存在、文件名格式非法、UP 段缺失时，错误语义与现有 API 保持一致，并在
  错误中保留 FS 逻辑路径。
- `SetFS(nil)` / `BindFS(nil)` 恢复磁盘读取。
- 由于 command 当前使用全局配置、数据库和 FS 绑定，不支持多个 `Migrator` 并发
  执行；文档明确要求顺序使用。

## 测试

- 使用 `testing/fstest.MapFS` 测试 `ParseFS`、非递归和递归发现。
- 验证 `_ignored` 目录、`_ignored.sql` 文件和非 `.sql` 文件过滤。
- 验证迁移按版本排序、FS 逻辑路径写入 `FilePath`。
- 验证 `MigrationsFromFS` 支持自动补 `.sql`、多目录和不存在文件错误。
- 用一个最小 `Migrator` 测试确认 `SetFS` 后 `Up`/`Status`/`Skip` 都从 FS 读取。
- 保留现有磁盘测试，确认 `nil` 模式无行为回归。

## 非目标

- 不把嵌入文件释放到临时目录。
- 不改变迁移文件格式、版本排序、SQL 执行和数据库记录逻辑。
- 不在本次工作中重构 command 的全局状态为实例化架构。
- 不增加第三方依赖。

