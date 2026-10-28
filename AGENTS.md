# AGENTS.md - miglite 开发指南

> 本文件为 AI 编程助手提供项目上下文与必须遵守的限制。

## 项目概述

**miglite** 是极简的数据库 schema 迁移工具，既是命令行工具，也可作为库嵌入应用。

- 根模块：`github.com/gookit/miglite`（库），`go.mod` 的 go 版本 **1.20**
- 命令行：`cmd/miglite` 是**独立模块**（`replace github.com/gookit/miglite => ../../`），只有它依赖数据库驱动
- 结构：`internal/`（config、database、migutil、runtime）、`pkg/`（command、migcom、migration）

## 依赖限制（必须遵守）

- **根模块不得依赖任何数据库驱动**（`modernc.org/sqlite`、`github.com/go-sql-driver/mysql`、`github.com/lib/pq`、pgx 等），包括测试文件：根包的 `*_test.go` 也会把驱动写进根 `go.mod`。
- **需要真实驱动的测试一律放在 `cmd/miglite/testdrv/`**（该目录已有 sqlite/mysql/pgsql 驱动测试与 `embedfixtures/` 嵌入迁移）。以使用者视角测试库 API 时用 `package testdrv_test`，只调用公开 API。
- **不要抬高根模块 `go.mod` 的 go 版本，不要顺手升级依赖**：它是库，最低 Go 版本与依赖版本会传导给所有使用者。
- 在多模块工作区（`go.work`）里开发时，**不要运行 `go work sync`**：它会把版本反写进工作区内每个模块的 `go.mod`。改完依赖后在本仓库根目录与 `cmd/miglite` 分别 `go mod tidy`，并检查 `git diff go.mod` 只有预期的变化。

## 作为库使用

- 外部可构造的配置：`miglite.Config` / `miglite.Database` / `miglite.Migrations`；`miglite.NewWithConfigAndFS(cfg, fsys).SetSqlDB(db)`。
- 库调用的输出经 `Config.Output`（`io.Writer`）控制，传 `io.Discard` 完全静默；为 nil 时保持命令行默认的彩色输出。
- 新增库能力时保持命令行默认行为不变。

## 构建与测试

```bash
go test ./...                              # 根模块（不含驱动测试）
cd cmd/miglite && go test ./...            # 命令行与驱动测试
make build                                 # 构建命令行（见 Makefile）
```

- 单元测试用 `github.com/gookit/goutil/x/assert`；同一方法多个用例用 `t.Run()` 分组。
- 迁移文件命名 `yyyymmdd-HHIISS-{name}.sql`，用 `-- Migrate:UP` 与可选的 `-- Migrate:DOWN` 分段。
