# gap 命令文档

`gap` 是独立脚手架 `go-agent-platform-cli` 编译出的命令。当前版本为 `0.5.0`。

它参考 `/Users/mac/Projects/sail/earn/app/server`，将该目录视为一个独立工程根：

```text
project/
  main.go
  Makefile
  api/proto/       # Protobuf source contracts
  api/gen/go/      # generated protobuf/grpc Go code
  hack/
  manifest/
  api/
  internal/
    controller/   # 业务分层
    service/
    logic/
    model/
    dao/
    platform/     # 应用与基础设施：cmd/config/server/middleware/storage/auth/cache/queue/observability
    agent/        # AI 能力：eino/tools/knowledge/workflow/mcp/eval/runs
```

不生成 `app/server` 或 `app/admin` 外层目录。

## 编译

```bash
cd /Users/mac/Projects/sail/sail-golang-game-base/go-agent-platform-cli
make build
./bin/gap version
```

## 命令和参数

| 命令 | 作用 |
| --- | --- |
| `gap project` | 创建独立 Web + Agent 平台工程。 |
| `gap module` | 一次生成完整业务模块。 |
| `gap ctrl` | 生成 API 和 GF 式多文件 Controller。 |
| `gap service` | 从 Logic 公开方法生成 Service 接口及注册器。 |
| `gap dao` | 生成根级 DAO；数据库代码由 Ent 负责。 |
| `gap version` | 显示 CLI 版本。 |

| 参数 | 作用 |
| --- | --- |
| `-name` | 项目名或模块名。 |
| `-root` | 目标项目根目录，默认当前目录。 |
| `-module` | 项目 Go module；模块命令默认读取 `go.mod`。 |
| `-dir` | `project` 的输出目录。 |
| `-force` | 覆盖已存在的生成文件。 |

旧参数 `-app` 仅为兼容保留，已经没有路径作用。

## 生成项目

```bash
./bin/gap project gap-test
```

完整写法：

```bash
./bin/gap project gap-test \
  -module example.com/gap-test \
  -dir /tmp/gap-test \
  -force
```

项目名中的 `-` 会原样保留。

## 初始化与启动

新项目在首次使用时执行：

```bash
cd gap-test
make deps
make ent-gen
docker compose -f manifest/docker/docker-compose.yml up -d
make run
```

`make deps` 固定并下载 Gin、Viper、pgx、Redis、Ent、OpenTelemetry、Uber Zap 和 Protobuf 依赖，同时更新 `go.sum`。服务启动路径会：加载 Viper YAML 配置并应用环境变量覆盖，连接并 Ping PostgreSQL，连接并 Ping Redis，随后启动 Gin。任一基础设施不可用时进程会返回错误，不会错误地报告服务已就绪。

Gin 默认注册 `/healthz` 与 `/readyz`。`/readyz` 会检查 PostgreSQL 和 Redis；中间件依次提供 request ID、panic recovery、结构化访问日志、请求超时、body 限制、CORS 与 OpenTelemetry span。认证和限流中间件作为可注入扩展点，业务路由按权限策略自行挂载。

日志配置位于 `manifest/config/config.<env>.yaml` 的 `logging` 段。默认使用 JSON 格式，以 Zap 同时写入 stdout 和 `logs/app.log`；文件按 100MB 轮转，保留 10 份、30 天并压缩。环境变量 `LOG_LEVEL`、`LOG_FILE` 可以覆盖 level 和路径。

## 生成完整模块

```bash
./bin/gap module -name orders -root /tmp/gap-test -force
```

主要输出：

```text
api/orders/orders.go
api/orders/v1/create.go
api/orders/v1/get.go
internal/controller/orders/orders.go
internal/controller/orders/orders_new.go
internal/controller/orders/orders_v1_create.go
internal/controller/orders/orders_v1_get.go
internal/service/orders.go
internal/logic/orders/orders.go
internal/logic/orders/orders_create.go
internal/logic/orders/orders_get.go
internal/model/orders.go
internal/dao/orders.go
```

`module` 还会重新生成 `internal/logic/logic.go`，将新增 Logic 包加入空导入注册表。

## 重新生成 Controller

```bash
./bin/gap ctrl -name orders -root /tmp/gap-test -force
# 不带 -name 时扫描 api/ 下所有模块并强制覆盖生成文件：
./bin/gap ctrl -root /tmp/gap-test
```

`ctrl` 以 `api/<name>/v1/*.go` 的 `XxxReq` + `XxxRes` 配对为事实来源，为每个配对生成：

- `api/<name>/<name>.go` 聚合接口方法
- `internal/controller/<name>/<name>_new.go` 构造函数
- `internal/controller/<name>/<name>_v1_<xxx>.go` 单方法 Controller

在 v1 里新增一个请求/返回类（如 `TestReq`/`TestRes`），重新运行 `gap ctrl` 就会自动补上 `Test` 方法，无需改模板。省略 `-name` 会强制刷新所有 Controller 生成文件；不要把业务实现写在这些文件中，写入 `internal/logic`。

## 重新生成 Service

```bash
./bin/gap service -name orders -root /tmp/gap-test -force
# 不带 -name 时全量强制重生成
./bin/gap service -root /tmp/gap-test
```

`service` 生成两部分：

- `internal/model/<name>.go`：以 v1 DTO 配对为事实来源，每个方法一个输入结构 + 模块统一输出结构
- `internal/service/<name>.go`：**扫描 `internal/logic/<name>` 的公开方法**生成 `IXxx` 接口、`Xxx()` 获取函数、`RegisterXxx(service IXxx)`——在 logic 里手写或新增的方法会自动同步进接口

logic 目录尚不存在（例如只跑了 `ctrl`）时，回退按 v1 DTO 推导接口。

## 重新生成 DAO

```bash
./bin/gap dao -name orders -root /tmp/gap-test -force
# 不带 -name 时全量强制重生成
./bin/gap dao -root /tmp/gap-test
```

`dao` 以 v1 DTO 配对为事实来源，生成：

- `internal/dao/<name>.go` 每个方法一个 dao 桩
- `internal/logic/<name>/<name>_<xxx>.go` 每个方法一个 logic 桩（转发给 dao）
- 刷新 `internal/logic/logic.go` 注册表

项目模板默认包含一个用于验证生成链路的 `TestRecord` schema：

```bash
cd /tmp/gap-test
make ent-gen
```

执行后可以在 `internal/platform/storage/ent/testrecord*.go` 和 `internal/platform/storage/ent/migrate/schema.go` 看到生成结果。正式业务实体仍然使用：

```bash
make ent-new NAME=Commission
make ent-gen
```

Ent 的目录职责固定如下：

- `internal/platform/storage/ent/schema/`：手写实体 Schema，例如 `commission.go`
- `internal/platform/storage/ent/`：Ent 生成的 Client、Query、Create、Update、Migrate 代码
- `internal/dao/`：业务 DAO，对 Ent Client 做领域数据访问封装

`make ent-new` 会下载并写入 `entgo.io/ent` 依赖，然后在正确的 schema 目录创建实体。`make ent-gen` 重新生成 Ent 代码；`make dao` 发现 schema 时也会自动调用它。

生成项目的 `go.mod` 默认固定 Ent、Gin、Viper、pgx、Redis、OpenTelemetry 与 Protobuf 版本；Ent 的基础校验和已预置。首次运行 `make deps` 会补齐其余依赖的 `go.sum`，之后再执行 `go test ./...`。

## Protobuf

Protobuf 的职责边界如下：

- `api/proto/<domain>/v1/*.proto`：手写、版本化的跨服务契约；例如 `api/proto/platform/v1/health.proto`
- `api/gen/go/`：Buf 生成的 Go protobuf 与 gRPC 文件，不手改
- `buf.yaml`、`buf.gen.yaml`：Buf 模块、lint 和插件配置

```bash
make proto-lint
make proto-gen
```

`proto-gen` 使用 Buf remote plugins 输出 Go 与 gRPC 代码到 `api/gen/go`，并在生成后执行 `go mod tidy`。后续微服务调用只依赖 `api/gen/go` 生成包，HTTP DTO 仍留在现有 `api/<module>/v1`，两者不要混用。

## Make 命令

在生成项目根目录运行：

```bash
make ctrl NAME=orders
make service NAME=orders
make dao NAME=orders
make build
make test
make run
```

Makefile 与 `earn/app/server` 一样，将具体规则拆到 `hack/hack.mk` 和 `hack/hack-cli.mk`。

## 完整验证

```bash
./bin/gap project gap-compile -module example.com/gap-compile -dir /tmp/gap-compile -force
./bin/gap module -name orders -root /tmp/gap-compile -force
./bin/gap ctrl -name orders -root /tmp/gap-compile -force
./bin/gap service -name orders -root /tmp/gap-compile -force
./bin/gap dao -name orders -root /tmp/gap-compile -force

cd /tmp/gap-compile
go test ./...
```
