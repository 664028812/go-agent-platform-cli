# go-agent-platform-cli

`go-agent-platform-cli` 是独立脚手架工程，用来生成课程里的高级 Web 后端 + Eino Agent 平台项目。

它不放在业务工程 `go-agent-platform` 里面，后续可以单独编译、安装和升级。

详细命令和实跑示例见：[docs/COMMANDS.md](/Users/mac/Projects/sail/sail-golang-game-base/go-agent-platform-cli/docs/COMMANDS.md)。

生成结构直接参考 `/Users/mac/Projects/sail/earn/app/server`：把它视为一个独立应用根，项目根目录直接包含 `main.go`、`api`、`internal`、`manifest` 和 `hack`，不再生成外层 `app/server`。

## Commands

```bash
make build
./bin/gap project gap-test
./bin/gap project -name go-agent-platform -module example.com/go-agent-platform -dir ../go-agent-platform
./bin/gap module -name orders -root ../go-agent-platform
./bin/gap ctrl [-name orders] -root ../go-agent-platform    # 不带 -name 全量强制重生成
./bin/gap service [-name orders] -root ../go-agent-platform # 不带 -name 全量强制重生成
./bin/gap dao [-name orders] -root ../go-agent-platform     # 不带 -name 全量强制重生成
```

| 命令            | 作用                                                                                                       |
| --------------- | ---------------------------------------------------------------------------------------------------------- |
| `gap project` | 生成独立应用根，包含 `main.go`、`api`、`internal`、`manifest`、`hack` 和 `Makefile`。 |
| `gap module`  | 在项目根一次性生成 `api / controller / service / logic / model / dao`。 |
| `gap ctrl`    | 以 `api/<name>/v1` 的 `XxxReq`/`XxxRes` 配对为事实来源，生成聚合接口、构造函数和按方法拆分的 Controller；`-name` 缺省时遍历所有模块。 |
| `gap service` | 以 v1 DTO 配对为事实来源，生成 model 输入输出与带注册/获取函数的 Service 接口；`-name` 缺省时全量重生成。 |
| `gap dao`     | 以 v1 DTO 配对为事实来源，生成 dao 与 logic 桩代码并刷新 logic 注册表；`-name` 缺省时全量重生成。 |
| `gap version` | 输出当前 CLI 版本。                                                                                        |

> 在 `api/<name>/v1/` 新增一对 `XxxReq`/`XxxRes` 后，重新运行 `gap ctrl` / `gap service` / `gap dao`（配合 `-force`），对应的接口方法与各层桩代码会自动扩展。

## Scaffold Shape

生成出来的业务项目参考 GoFrame 分层习惯，但技术栈固定为：

```text
Web: Gin
Config: Viper + YAML + env override
Database: PostgreSQL
Data Layer: Ent
Cache: Redis + go-redis
Async Jobs: PostgreSQL durable jobs first
Agent: Eino
Observability: slog + OpenTelemetry + Prometheus
Layering: api / controller / service / logic / model / dao
```

## Make Commands

CLI 自己的 `Makefile` 负责脚手架工程：

```bash
make fmt
make test
make build
make install
make example-project
make example-module NAME=orders
make example-ctrl NAME=orders
make example-service NAME=orders
make example-dao NAME=orders
make clean
```

| Make 命令                | 作用                                                        |
| ------------------------ | ----------------------------------------------------------- |
| `make fmt`             | 格式化 CLI 工程代码。                                       |
| `make test`            | 运行 CLI 单元测试。                                         |
| `make build`           | 编译`bin/gap`。                                           |
| `make install`         | 安装`gap` 到 Go bin。                                     |
| `make example-project` | 在`/tmp/go-agent-platform-example` 生成一个完整项目示例。 |
| `make example-module`  | 在示例项目中生成完整业务模块。                              |
| `make example-ctrl`    | 在示例项目中只生成 api 和 controller。                      |
| `make example-service` | 按示例项目 logic 重新生成 service 接口。                    |
| `make example-dao`     | 在示例项目中生成 dao 文件，并可验证默认 TestRecord schema。 |
| `make clean`           | 删除 CLI 编译产物。                                         |

生成出来的平台工程也会自带 `Makefile` 和 `hack/*.mk`，包含 `ctrl`、`service`、`dao`、`run`、`deps`、`ent-new`、`ent-gen`、`compose-up` 等项目命令。模板内置 `internal/platform/storage/ent/schema/test_record.go` 作为 Ent 生成测试表；运行 `make ent-gen` 可验证生成链路，正式实体使用 `make ent-new NAME=Commission` 创建。

新项目先执行 `make deps`，然后执行 `make ent-gen`。模板会初始化 Viper 配置加载、PostgreSQL/Redis Ping、Gin HTTP 服务和基础中间件。Protobuf 源文件在 `api/proto/<domain>/v1`，Buf 生成代码在 `api/gen/go`，使用 `make proto-lint` 和 `make proto-gen`。
