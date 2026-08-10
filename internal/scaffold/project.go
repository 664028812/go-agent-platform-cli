package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// projectDirs 项目骨架需要预建的目录。
// 业务分层（controller/service/logic/model/dao）留在顶层；
// 平台基础设施收敛到 platform/，AI 能力收敛到 agent/。
func projectDirs() []string {
	return []string{
		"cmd/worker",
		"cmd/mcp-server",
		"hack",
		"manifest/config",
		"manifest/docker",
		"manifest/deploy",
		"api/proto/platform/v1",
		"api/gen/go",
		"docs",
		"migrations",
		"internal/controller",
		"internal/service",
		"internal/logic",
		"internal/model",
		"internal/dao",
		"internal/platform/cmd",
		"internal/platform/consts",
		"internal/platform/config",
		"internal/platform/server",
		"internal/platform/middleware",
		"internal/platform/storage/ent/schema",
		"internal/platform/storage",
		"internal/platform/auth",
		"internal/platform/cache",
		"internal/platform/queue",
		"internal/platform/observability",
		"internal/platform/tools",
		"internal/agent/eino",
		"internal/agent/tools",
		"internal/agent/knowledge",
		"internal/agent/workflow",
		"internal/agent/mcp",
		"internal/agent/eval",
		"internal/agent/runs",
		"pkg/errors",
		"pkg/response",
	}
}

// projectFiles 项目级固定文件（不随模块变化）
func projectFiles(name string, modulePath string) map[string]string {
	files := map[string]string{
		"go.mod":                                     projectGoMod(modulePath),
		"go.sum":                                     projectGoSum(),
		"buf.yaml":                                   bufConfig(),
		"buf.gen.yaml":                               bufGenerateConfig(),
		"api/proto/platform/v1/health.proto":         protoHealthFile(modulePath),
		"README.md":                                  projectReadme(name),
		"Makefile":                                   projectMakefile(),
		"hack/hack.mk":                               projectHackMakefile(),
		"hack/hack-cli.mk":                           projectHackCLIMakefile(),
		".env.example":                               projectEnvExample(),
		"manifest/config/config.local.yaml":          projectConfig("local"),
		"manifest/config/config.dev.yaml":            projectConfig("dev"),
		"manifest/config/config.prod.yaml":           projectConfig("prod"),
		"manifest/docker/docker-compose.yml":         projectCompose(),
		"docs/README.md":                             docsReadme(name),
		"migrations/.gitkeep":                        "",
		"main.go":                                    standaloneMainFile(modulePath),
		"cmd/worker/main.go":                         mainFile(modulePath, "Worker", "RunWorker"),
		"cmd/mcp-server/main.go":                     mainFile(modulePath, "MCPServer", "RunMCPServer"),
		"internal/platform/cmd/server.go":            serverCommandFile(modulePath),
		"internal/platform/consts/app.go":            "package consts\n\nconst ServiceName = \"" + name + "\"\n",
		"internal/platform/cmd/worker.go":            workerCommandFile(modulePath),
		"internal/platform/cmd/mcp.go":               mcpCommandFile(modulePath),
		"internal/platform/config/config.go":         configTypes(),
		"internal/platform/config/loader.go":         configLoader(),
		"internal/platform/server/router.go":         routerFile(),
		"internal/platform/server/http_server.go":    httpServerFile(modulePath),
		"internal/platform/server/response.go":       responseServerFile(),
		"internal/platform/middleware/request_id.go": requestIDMiddlewareFile(),
		"internal/platform/middleware/recover.go":    recoveryMiddlewareFile(),
		"internal/platform/middleware/timeout.go":    timeoutMiddlewareFile(),
		"internal/platform/middleware/access_log.go": accessLogMiddlewareFile(),
		"internal/platform/middleware/body_limit.go": bodyLimitMiddlewareFile(),
		"internal/platform/middleware/cors.go":       corsMiddlewareFile(),
		"internal/platform/middleware/auth.go":       authMiddlewareFile(modulePath),
		"internal/platform/middleware/rate_limit.go": rateLimitMiddlewareFile(),
		"internal/platform/middleware/otel.go":       otelMiddlewareFile(),
		"internal/platform/auth/jwt.go":              `package auth` + "\n\n" + `type TokenManager struct { Secret string }` + "\n",
		"internal/platform/auth/rbac.go":             rbacFile(),
		"internal/platform/auth/principal.go":        principalFile(),
		"internal/platform/cache/redis.go":           redisFile(modulePath),
		"internal/platform/cache/lock.go":            `package cache` + "\n\n" + `type Lock struct { Key string }` + "\n",
		"internal/platform/cache/rate_limiter.go":    `package cache` + "\n\n" + `type RateLimiter struct { RequestsPerMinute int }` + "\n",
		"internal/platform/cache/cache_aside.go":     `package cache` + "\n\n" + `type CacheAside struct{}` + "\n",
		"internal/platform/queue/job.go":             jobFile(),
		"internal/platform/queue/job_store.go":       `package queue` + "\n\n" + `type JobStore interface { Enqueue(job Job) error }` + "\n",
		"internal/platform/queue/dispatcher.go":      `package queue` + "\n\n" + `type Dispatcher struct{}` + "\n",
		"internal/platform/queue/worker_pool.go":     `package queue` + "\n\n" + `type WorkerPool struct { Concurrency int }` + "\n",
		"internal/platform/queue/retry.go":           `package queue` + "\n\n" + `type RetryPolicy struct { MaxRetries int }` + "\n",
		"internal/platform/queue/dead_letter.go":     `package queue` + "\n\n" + `type DeadLetter struct { JobID string; Error string }` + "\n",
		"internal/platform/observability/logger.go":  loggerFile(modulePath),
		"internal/platform/observability/metrics.go": `package observability` + "\n\n" + `type Metrics struct { Port int }` + "\n",
		"internal/platform/observability/tracing.go": `package observability` + "\n\n" + `type Tracing struct { Endpoint string }` + "\n",
		"internal/platform/storage/ent/generate.go": `package ent

//go:generate env GOFLAGS=-mod=mod go run -mod=mod entgo.io/ent/cmd/ent generate ./schema
`,
		"internal/platform/storage/ent/schema/test_record.go": entTestRecordSchema(),
		"internal/platform/storage/ent/tools.go": `//go:build tools

package ent

import _ "entgo.io/ent/cmd/ent"
`,
		"internal/platform/tools/tools.go":        dependencyToolsFile(),
		"internal/platform/storage/ent/client.go": entClientFile(),
		"internal/platform/storage/ent/tx.go":     `package ent` + "\n\n" + `type Tx struct{}` + "\n",
		"internal/platform/storage/postgres.go":   postgresFile(modulePath),
		"internal/agent/runs/state_machine.go":    runStateMachineFile(),
		"internal/agent/runs/statistics.go":       `package runs` + "\n\n" + `type Statistics struct { TotalRuns int64; SuccessRuns int64; FailedRuns int64 }` + "\n",
		"internal/agent/runs/events.go":           `package runs` + "\n\n" + `type Event struct { RunID string; Type string }` + "\n",
		"internal/agent/eino/chat_model.go":       `package eino` + "\n\n" + `type ChatModel struct { Provider string; Model string }` + "\n",
		"internal/agent/eino/agent.go":            `package eino` + "\n\n" + `type Agent struct { Name string }` + "\n",
		"internal/agent/eino/callbacks.go":        `package eino` + "\n\n" + `type CallbackEvent struct { Type string }` + "\n",
		"internal/agent/eino/stream.go":           `package eino` + "\n\n" + `type StreamEvent struct { Text string }` + "\n",
		"internal/agent/tools/registry.go":        toolRegistryFile(),
		"internal/agent/tools/permission.go":      `package tools` + "\n\n" + `type PermissionChecker struct{}` + "\n",
		"internal/agent/tools/audit.go":           `package tools` + "\n\n" + `type AuditRecord struct { ToolName string; Error string }` + "\n",
		"internal/agent/knowledge/loader.go":      `package knowledge` + "\n\n" + `type Loader struct{}` + "\n",
		"internal/agent/knowledge/splitter.go":    `package knowledge` + "\n\n" + `type Splitter struct{}` + "\n",
		"internal/agent/knowledge/embedding.go":   `package knowledge` + "\n\n" + `type Embedder struct{}` + "\n",
		"internal/agent/knowledge/retriever.go":   `package knowledge` + "\n\n" + `type Retriever struct{}` + "\n",
		"internal/agent/knowledge/citation.go":    `package knowledge` + "\n\n" + `type Citation struct { Source string }` + "\n",
		"internal/agent/workflow/research_report.go": `package workflow

type ResearchReport struct{}
`,
		"internal/agent/workflow/review.go": `package workflow` + "\n\n" + `type Review struct{}` + "\n",
		"internal/agent/mcp/server.go":      `package mcp` + "\n\n" + `type Server struct{}` + "\n",
		"internal/agent/mcp/tools.go":       `package mcp` + "\n\n" + `type ToolBridge struct{}` + "\n",
		"internal/agent/eval/golden.go":     `package eval` + "\n\n" + `type GoldenCase struct { Input string; Expected string }` + "\n",
		"internal/agent/eval/runner.go":     `package eval` + "\n\n" + `type Runner struct{}` + "\n",
		"internal/agent/eval/metrics.go":    `package eval` + "\n\n" + `type Metrics struct { SuccessRate float64 }` + "\n",
		"pkg/errors/errors.go":              errorsFile(),
		"pkg/response/response.go":          responseFile(),
	}
	files["internal/logic/logic.go"] = logicRegistryContent([]string{"agents"}, modulePath)

	return files
}

func projectGoMod(modulePath string) string {
	return fmt.Sprintf(`module %s

go 1.25.6

require (
	entgo.io/ent v0.14.6
	github.com/gin-gonic/gin v1.12.0
	github.com/jackc/pgx/v5 v5.10.0
	github.com/redis/go-redis/v9 v9.22.0
	github.com/spf13/viper v1.21.0
	go.opentelemetry.io/otel v1.45.0
	go.uber.org/zap v1.27.1
	google.golang.org/grpc v1.83.0
	google.golang.org/protobuf v1.36.11
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)
`, modulePath)
}

func projectGoSum() string {
	return `entgo.io/ent v0.14.6 h1:/f2696BpwuWAEEG6PVGWflg6+Inrpq4pRWuNlWz/Skk=
entgo.io/ent v0.14.6/go.mod h1:z46QBUdGC+BATwsedbDuREfSS0oSCV+csdEYlL4p73s=
`
}

func dependencyToolsFile() string {
	return `//go:build tools

package tools

import (
	_ "google.golang.org/grpc"
	_ "google.golang.org/protobuf/proto"
)
`
}

func bufConfig() string {
	return `version: v2
modules:
  - path: api/proto
lint:
  use:
    - STANDARD
breaking:
  use:
    - FILE
`
}

func bufGenerateConfig() string {
	return `version: v2
plugins:
  - remote: buf.build/protocolbuffers/go
    out: api/gen/go
    opt:
      - paths=source_relative
  - remote: buf.build/grpc/go
    out: api/gen/go
    opt:
      - paths=source_relative
`
}

func protoHealthFile(modulePath string) string {
	return fmt.Sprintf(`syntax = "proto3";

package platform.v1;

option go_package = "%s/api/gen/go/platform/v1;platformv1";

service HealthService {
  rpc Check(HealthCheckRequest) returns (HealthCheckResponse);
}

message HealthCheckRequest {}

message HealthCheckResponse {
  string status = 1;
}
`, modulePath)
}

func projectReadme(name string) string {
	return fmt.Sprintf(`# %s

Go 高级 Web 后端 + 分布式平台 + Eino AI Agent 平台项目骨架。

## Stack

~~~text
Web: Gin
Config: Viper + YAML + env override
Database: PostgreSQL
Data Layer: Ent
Cache: Redis + go-redis
Async Jobs: PostgreSQL durable jobs first
Agent: Eino
Observability: Uber Zap + OpenTelemetry + Prometheus
Layering: api / controller / service / logic / model / dao
~~~

## Layer Rule

~~~text
api -> controller -> service -> logic -> dao -> Ent/PostgreSQL
~~~

## Project Shape

~~~text
main.go                  # standalone application entry
api/                     # public request/response contracts
internal/controller/     # GF-style controller files
internal/service/        # IXxx + Xxx + RegisterXxx
internal/logic/          # sXxx implementations and logic.go registry
internal/model/          # internal inputs and outputs
internal/dao/            # Ent data access wrappers
internal/platform/       # app infra: cmd/config/server/middleware/storage/auth/cache/queue/observability
internal/agent/          # AI capabilities: eino/tools/knowledge/workflow/mcp/eval/runs
api/proto/               # protobuf source contracts, grouped by domain and version
api/gen/go/              # Buf-generated Go protobuf/grpc code
manifest/                # config and Docker Compose
hack/                    # Make targets, following earn/app/server
~~~

`+"首次执行 `make deps` 安装 Gin、Viper、pgx、Redis、Ent 与可观测性依赖；再执行 `make ent-gen` 生成默认 `TestRecord` 的 Ent 代码。服务启动时会分别 Ping PostgreSQL 和 Redis，连接失败会阻止应用进入可服务状态。\n", name)
}

func projectMakefile() string {
	return `ROOT_DIR = $(shell pwd)
GAP ?= gap

include ./hack/hack-cli.mk
include ./hack/hack.mk
`
}

func projectHackCLIMakefile() string {
	return `.PHONY: cli.install

cli.install:
	@command -v $(GAP) >/dev/null 2>&1 || (echo "gap CLI is required; add go-agent-platform-cli/bin to PATH" && exit 1)
`
}

func projectHackMakefile() string {
	return `.DEFAULT_GOAL := build

.PHONY: build ctrl service dao fmt test run deps ent-new ent-gen proto-gen proto-lint compose-up compose-down clean

build: cli.install
	mkdir -p bin
	GOCACHE=/tmp/go-agent-platform-cache go build -o bin/server .
	GOCACHE=/tmp/go-agent-platform-cache go build -o bin/worker ./cmd/worker
	GOCACHE=/tmp/go-agent-platform-cache go build -o bin/mcp-server ./cmd/mcp-server

ctrl: cli.install
	@if [ -n "$${NAME}" ]; then $(GAP) ctrl -name $${NAME} -root . -force; else $(GAP) ctrl -root . -force; fi

service: cli.install
	@if [ -n "$${NAME}" ]; then $(GAP) service -name $${NAME} -root . -force; else $(GAP) service -root . -force; fi

dao: cli.install
	@if [ -n "$${NAME}" ]; then $(GAP) dao -name $${NAME} -root . -force; else $(GAP) dao -root . -force; fi
	@if find internal/platform/storage/ent/schema -type f -name '*.go' -print -quit | grep -q .; then $(MAKE) ent-gen; else echo "Ent schema is empty; skipped Ent generation"; fi

fmt:
	gofmt -w $$(find . -name '*.go' -print)

test:
	GOCACHE=/tmp/go-agent-platform-cache go test ./...

run:
	GOCACHE=/tmp/go-agent-platform-cache go run .

deps:
	go get github.com/gin-gonic/gin
	go get github.com/spf13/viper
	go get entgo.io/ent
	go get github.com/redis/go-redis/v9
	go get github.com/jackc/pgx/v5
	go get go.opentelemetry.io/otel
	go get go.uber.org/zap
	go get gopkg.in/natefinch/lumberjack.v2
	go get github.com/prometheus/client_golang/prometheus
	go get google.golang.org/grpc
	go get google.golang.org/protobuf
	go mod tidy

ent-new:
	@test -n "$${NAME}" || (echo "NAME is required, example: make ent-new NAME=User" && exit 1)
	go get entgo.io/ent
	go get entgo.io/ent/cmd/ent
	go mod tidy
	go run -mod=mod entgo.io/ent/cmd/ent init --target internal/platform/storage/ent/schema $${NAME}

ent-gen:
	@test -n "$$(find internal/platform/storage/ent/schema -type f -name '*.go' -print -quit)" || (echo "Ent schema is empty; create one with: make ent-new NAME=Commission" && exit 1)
	go get entgo.io/ent
	go get entgo.io/ent/cmd/ent
	go mod tidy
	GOFLAGS=-mod=mod GOCACHE=/tmp/go-agent-platform-cache go run -mod=mod entgo.io/ent/cmd/ent generate ./internal/platform/storage/ent/schema

proto-gen:
	@command -v buf >/dev/null 2>&1 || (echo "buf is required; install it from https://buf.build/docs/installation" && exit 1)
	buf generate
	go mod tidy

proto-lint:
	@command -v buf >/dev/null 2>&1 || (echo "buf is required; install it from https://buf.build/docs/installation" && exit 1)
	buf lint

compose-up:
	docker compose -f manifest/docker/docker-compose.yml up -d

compose-down:
	docker compose -f manifest/docker/docker-compose.yml down

clean:
	rm -rf bin
`
}

func projectEnvExample() string {
	return `APP_ENV=local
CONFIG_PATH=manifest/config/config.local.yaml
DATABASE_DSN=postgres://agent:agent@localhost:5432/agent_platform?sslmode=disable
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
HTTP_ALLOWED_ORIGINS=http://localhost:3000
LOG_LEVEL=info
LOG_FILE=logs/app.log
JWT_SECRET=local-dev-secret
EINO_PROVIDER=mock
EINO_MODEL=mock-chat
OTLP_ENDPOINT=localhost:4317
`
}

func projectConfig(env string) string {
	return fmt.Sprintf(`app:
  name: go-agent-platform
  env: %s
  port: 8080
  shutdown_timeout: 10s

http:
  read_timeout: 5s
  write_timeout: 60s
  idle_timeout: 60s
  request_timeout: 30s
  body_limit: 10485760
  allowed_origins:
    - http://localhost:3000

database:
  driver: postgres
  dsn: postgres://agent:agent@localhost:5432/agent_platform?sslmode=disable
  max_open_conns: 20
  max_idle_conns: 10
  conn_max_lifetime: 1h
  connect_timeout: 5s

redis:
  addr: localhost:6379
  password: ""
  db: 0
  connect_timeout: 3s
  pool_size: 10

auth:
  jwt_secret: local-dev-secret
  access_token_ttl: 2h

worker:
  worker_id: local-worker-1
  concurrency: 8
  poll_interval: 1s
  lock_ttl: 60s
  max_retries: 3

eino:
  provider: mock
  model: mock-chat
  timeout: 60s
  max_steps: 8

observability:
  service_name: go-agent-platform
  otlp_endpoint: localhost:4317
  metrics_port: 9090

logging:
  level: info
  format: json
  file: logs/app.log
  max_size_mb: 100
  max_backups: 10
  max_age_days: 30
  compress: true
`, env)
}

func projectCompose() string {
	return `services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_USER: agent
      POSTGRES_PASSWORD: agent
      POSTGRES_DB: agent_platform
    ports:
      - "5432:5432"

  redis:
    image: redis:7
    ports:
      - "6379:6379"
`
}

func docsReadme(name string) string {
	return "# " + name + " Docs\n\n这里放接口、架构、状态机、异步任务、统计口径和运维文档。\n"
}

func standaloneMainFile(modulePath string) string {
	return fmt.Sprintf(`package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"%s/internal/platform/cmd"
	_ "%s/internal/logic"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := cmd.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
`, modulePath, modulePath)
}

func mainFile(modulePath string, appName string, runFunc string) string {
	return fmt.Sprintf(`package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"%s/internal/platform/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := cmd.%s(ctx); err != nil {
		log.Fatal(err)
	}
}
`, modulePath, runFunc)
}

func serverCommandFile(modulePath string) string {
	return fmt.Sprintf(`package cmd

import (
	"context"
	"fmt"

	"%s/internal/platform/cache"
	"%s/internal/platform/config"
	"%s/internal/platform/observability"
	"%s/internal/platform/server"
	"%s/internal/platform/storage"
	entstore "%s/internal/platform/storage/ent"
)

func Run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger, err := observability.NewLogger(cfg.Logging, cfg.Observability.ServiceName)
	if err != nil {
		return fmt.Errorf("initialize logger: %%w", err)
	}
	defer logger.Sync()
	db, err := storage.OpenPostgres(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("open postgres: %%w", err)
	}
	defer db.Close()
	entClient, err := entstore.Open(cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		return fmt.Errorf("open ent client: %%w", err)
	}
	defer entClient.Close()
	redisClient, err := cache.OpenRedis(ctx, cfg.Redis)
	if err != nil {
		return fmt.Errorf("open redis: %%w", err)
	}
	defer redisClient.Close()

	app := server.NewHTTPServer(cfg, logger, server.Dependencies{
		Ready: func(checkCtx context.Context) error {
			if err := storage.Ping(checkCtx, db); err != nil {
				return err
			}
			return cache.Ping(checkCtx, redisClient)
		},
	})
	return app.Run(ctx)
}
`, modulePath, modulePath, modulePath, modulePath, modulePath, modulePath)
}

func workerCommandFile(modulePath string) string {
	return fmt.Sprintf(`package cmd

import (
	"context"

	"%s/internal/platform/config"
)

func RunWorker(ctx context.Context) error {
	_, err := config.Load()
	_ = ctx
	return err
}
`, modulePath)
}

func mcpCommandFile(modulePath string) string {
	return fmt.Sprintf(`package cmd

import (
	"context"

	"%s/internal/platform/config"
)

func RunMCPServer(ctx context.Context) error {
	_, err := config.Load()
	_ = ctx
	return err
}
`, modulePath)
}

func logicRegistryFile(root string, modulePath string) string {
	entries, err := os.ReadDir(filepath.Join(root, "internal", "logic"))
	if err != nil {
		return logicRegistryContent(nil, modulePath)
	}
	modules := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() && sanitizeIdentifier(entry.Name()) != "" {
			modules = append(modules, entry.Name())
		}
	}
	sort.Strings(modules)
	return logicRegistryContent(modules, modulePath)
}

func logicRegistryContent(modules []string, modulePath string) string {
	var builder strings.Builder
	builder.WriteString("// Code generated by gap. DO NOT EDIT.\n\npackage logic\n")
	if len(modules) == 0 {
		return builder.String()
	}
	builder.WriteString("\nimport (\n")
	for _, module := range modules {
		builder.WriteString("\t_ ")
		builder.WriteString(strconv.Quote(modulePath + "/internal/logic/" + sanitizeIdentifier(module)))
		builder.WriteString("\n")
	}
	builder.WriteString(")\n")
	return builder.String()
}

func appEntry(modulePath string, runFunc string) string {
	return fmt.Sprintf(`package app

import (
	"context"
	"fmt"

	"%s/internal/platform/config"
)

func %s(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	_ = ctx
	fmt.Printf("%%s starting on :%%d\n", cfg.App.Name, cfg.App.Port)
	return nil
}
`, modulePath, runFunc)
}

func configTypes() string {
	return `package config

import "time"

type Config struct {
	App           AppConfig           ` + "`mapstructure:\"app\"`" + `
	HTTP          HTTPConfig          ` + "`mapstructure:\"http\"`" + `
	Database      DatabaseConfig      ` + "`mapstructure:\"database\"`" + `
	Redis         RedisConfig         ` + "`mapstructure:\"redis\"`" + `
	Auth          AuthConfig          ` + "`mapstructure:\"auth\"`" + `
	Worker        WorkerConfig        ` + "`mapstructure:\"worker\"`" + `
	Eino          EinoConfig          ` + "`mapstructure:\"eino\"`" + `
	Observability ObservabilityConfig ` + "`mapstructure:\"observability\"`" + `
	Logging       LoggingConfig       ` + "`mapstructure:\"logging\"`" + `
}

type AppConfig struct {
	Name            string        ` + "`mapstructure:\"name\"`" + `
	Env             string        ` + "`mapstructure:\"env\"`" + `
	Port            int           ` + "`mapstructure:\"port\"`" + `
	ShutdownTimeout time.Duration ` + "`mapstructure:\"shutdown_timeout\"`" + `
}

type HTTPConfig struct {
	ReadTimeout    time.Duration ` + "`mapstructure:\"read_timeout\"`" + `
	WriteTimeout   time.Duration ` + "`mapstructure:\"write_timeout\"`" + `
	IdleTimeout    time.Duration ` + "`mapstructure:\"idle_timeout\"`" + `
	RequestTimeout time.Duration ` + "`mapstructure:\"request_timeout\"`" + `
	BodyLimit      int64         ` + "`mapstructure:\"body_limit\"`" + `
	AllowedOrigins []string      ` + "`mapstructure:\"allowed_origins\"`" + `
}

type DatabaseConfig struct {
	Driver          string        ` + "`mapstructure:\"driver\"`" + `
	DSN             string        ` + "`mapstructure:\"dsn\"`" + `
	MaxOpenConns    int           ` + "`mapstructure:\"max_open_conns\"`" + `
	MaxIdleConns    int           ` + "`mapstructure:\"max_idle_conns\"`" + `
	ConnMaxLifetime time.Duration ` + "`mapstructure:\"conn_max_lifetime\"`" + `
	ConnectTimeout  time.Duration ` + "`mapstructure:\"connect_timeout\"`" + `
}

type RedisConfig struct {
	Addr            string        ` + "`mapstructure:\"addr\"`" + `
	Password        string        ` + "`mapstructure:\"password\"`" + `
	DB              int           ` + "`mapstructure:\"db\"`" + `
	ConnectTimeout  time.Duration ` + "`mapstructure:\"connect_timeout\"`" + `
	PoolSize        int           ` + "`mapstructure:\"pool_size\"`" + `
}

type AuthConfig struct {
	JWTSecret      string        ` + "`mapstructure:\"jwt_secret\"`" + `
	AccessTokenTTL time.Duration ` + "`mapstructure:\"access_token_ttl\"`" + `
}

type WorkerConfig struct {
	WorkerID     string
	Concurrency  int
	PollInterval time.Duration
	LockTTL      time.Duration
	MaxRetries   int
}

type EinoConfig struct {
	Provider string
	Model    string
	Timeout  time.Duration
	MaxSteps int
}

type ObservabilityConfig struct {
	ServiceName  string ` + "`mapstructure:\"service_name\"`" + `
	OTLPEndpoint string ` + "`mapstructure:\"otlp_endpoint\"`" + `
	MetricsPort  int    ` + "`mapstructure:\"metrics_port\"`" + `
}

type LoggingConfig struct {
	Level      string ` + "`mapstructure:\"level\"`" + `
	Format     string ` + "`mapstructure:\"format\"`" + `
	File       string ` + "`mapstructure:\"file\"`" + `
	MaxSizeMB  int    ` + "`mapstructure:\"max_size_mb\"`" + `
	MaxBackups int    ` + "`mapstructure:\"max_backups\"`" + `
	MaxAgeDays int    ` + "`mapstructure:\"max_age_days\"`" + `
	Compress   bool   ` + "`mapstructure:\"compress\"`" + `
}
`
}

func configLoader() string {
	return `package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

func Load() (Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath())
	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if err := applyEnv(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, cfg.Validate()
}

func (c Config) Validate() error {
	if c.App.Port <= 0 || c.Database.DSN == "" || c.Redis.Addr == "" {
		return fmt.Errorf("app.port, database.dsn and redis.addr are required")
	}
	return nil
}

func configPath() string {
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}
	return "manifest/config/config.local.yaml"
}

func applyEnv(cfg *Config) error {
	cfg.App.Env = envString("APP_ENV", cfg.App.Env)
	cfg.App.Name = envString("APP_NAME", cfg.App.Name)
	cfg.Database.DSN = envString("DATABASE_DSN", cfg.Database.DSN)
	cfg.Redis.Addr = envString("REDIS_ADDR", cfg.Redis.Addr)
	cfg.Redis.Password = envString("REDIS_PASSWORD", cfg.Redis.Password)
	if origins := os.Getenv("HTTP_ALLOWED_ORIGINS"); origins != "" { cfg.HTTP.AllowedOrigins = strings.Split(origins, ",") }
	cfg.Auth.JWTSecret = envString("JWT_SECRET", cfg.Auth.JWTSecret)
	cfg.Eino.Provider = envString("EINO_PROVIDER", cfg.Eino.Provider)
	cfg.Eino.Model = envString("EINO_MODEL", cfg.Eino.Model)
	cfg.Observability.OTLPEndpoint = envString("OTLP_ENDPOINT", cfg.Observability.OTLPEndpoint)
	cfg.Logging.Level = envString("LOG_LEVEL", cfg.Logging.Level)
	cfg.Logging.File = envString("LOG_FILE", cfg.Logging.File)
	var err error
	if cfg.App.Port, err = envInt("APP_PORT", cfg.App.Port); err != nil { return err }
	if cfg.Redis.DB, err = envInt("REDIS_DB", cfg.Redis.DB); err != nil { return err }
	return nil
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" { return value }
	return fallback
}

func envInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" { return fallback, nil }
	parsed, err := strconv.Atoi(value)
	if err != nil { return 0, fmt.Errorf("parse %s: %w", key, err) }
	return parsed, nil
}
`
}

func postgresFile(modulePath string) string {
	return fmt.Sprintf(`package storage

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"%s/internal/platform/config"
)

func OpenPostgres(ctx context.Context, cfg config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	pingCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping postgres: %%w", err)
	}
	return db, nil
}

func Ping(ctx context.Context, db *sql.DB) error {
	return db.PingContext(ctx)
}
`, modulePath)
}

func redisFile(modulePath string) string {
	return fmt.Sprintf(`package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"%s/internal/platform/config"
)

func OpenRedis(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
	pingCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping redis: %%w", err)
	}
	return client, nil
}

func Ping(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}
`, modulePath)
}

func loggerFile(modulePath string) string {
	return fmt.Sprintf(`package observability

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"%s/internal/platform/config"
)

func NewLogger(cfg config.LoggingConfig, serviceName string) (*zap.Logger, error) {
	level := zap.NewAtomicLevel()
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil { return nil, err }
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	if cfg.Format == "console" { encoder = zapcore.NewConsoleEncoder(encoderConfig) }
	cores := []zapcore.Core{zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)}
	if cfg.File != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.File), 0o755); err != nil { return nil, err }
		fileWriter := &lumberjack.Logger{Filename: cfg.File, MaxSize: cfg.MaxSizeMB, MaxBackups: cfg.MaxBackups, MaxAge: cfg.MaxAgeDays, Compress: cfg.Compress}
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(fileWriter), level))
	}
	return zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).With(zap.String("service", serviceName)), nil
}
`, modulePath)
}

func routerFile() string {
	return `package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func registerRoutes(engine *gin.Engine, ready Readiness) {
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, Response{Code: "OK", Message: "healthy"})
	})
	engine.GET("/readyz", func(c *gin.Context) {
		if err := ready(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, Response{Code: "NOT_READY", Message: err.Error()})
			return
		}
		c.JSON(http.StatusOK, Response{Code: "OK", Message: "ready"})
	})
}
`
}

func responseServerFile() string {
	return `package server

type Response struct {
	Code    string ` + "`json:\"code\"`" + `
	Message string ` + "`json:\"message\"`" + `
	Data    any    ` + "`json:\"data,omitempty\"`" + `
}
`
}

func httpServerFile(modulePath string) string {
	return fmt.Sprintf(`package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"%s/internal/platform/config"
	"%s/internal/platform/middleware"
)

type Readiness func(context.Context) error

type Dependencies struct { Ready Readiness }

type HTTPServer struct {
	server          *http.Server
	shutdownTimeout time.Duration
}

func NewHTTPServer(cfg config.Config, logger *zap.Logger, deps Dependencies) *HTTPServer {
	if cfg.App.Env == "production" { gin.SetMode(gin.ReleaseMode) }
	engine := gin.New()
	engine.Use(
		middleware.RequestID(),
		middleware.Recovery(logger),
		middleware.AccessLog(logger),
		middleware.Timeout(cfg.HTTP.RequestTimeout),
		middleware.BodyLimit(cfg.HTTP.BodyLimit),
		middleware.CORS(cfg.HTTP.AllowedOrigins),
		middleware.OTel(cfg.Observability.ServiceName),
	)
	ready := deps.Ready
	if ready == nil { ready = func(context.Context) error { return nil } }
	registerRoutes(engine, ready)
	return &HTTPServer{shutdownTimeout: cfg.App.ShutdownTimeout, server: &http.Server{
		Addr: fmt.Sprintf(":%%d", cfg.App.Port), Handler: engine,
		ReadTimeout: cfg.HTTP.ReadTimeout, WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout: cfg.HTTP.IdleTimeout,
	}}
}

func (s *HTTPServer) Run(ctx context.Context) error {
	errs := make(chan error, 1)
	go func() {
		err := s.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) { errs <- err; return }
		errs <- nil
	}()
	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()
		return s.server.Shutdown(shutdownCtx)
	}
}
`, modulePath, modulePath)
}

func requestIDMiddlewareFile() string {
	return `package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

const RequestIDHeader = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" { requestID = newRequestID() }
		c.Set(RequestIDHeader, requestID)
		c.Header(RequestIDHeader, requestID)
		c.Next()
	}
}

func newRequestID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil { return "request-id-unavailable" }
	return hex.EncodeToString(value)
}
`
}

func recoveryMiddlewareFile() string {
	return `package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.Error("panic recovered", zap.String("request_id", c.GetString(RequestIDHeader)), zap.Any("error", recovered))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL", "message": "internal server error"})
	})
}
`
}

func timeoutMiddlewareFile() string {
	return `package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if timeout <= 0 { c.Next(); return }
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
		if ctx.Err() == context.DeadlineExceeded && !c.Writer.Written() {
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{"code": "TIMEOUT", "message": "request timeout"})
		}
	}
}
`
}

func accessLogMiddlewareFile() string {
	return `package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AccessLog(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		logger.Info("http request", zap.String("request_id", c.GetString(RequestIDHeader)), zap.String("method", c.Request.Method), zap.String("path", c.Request.URL.Path), zap.Int("status", c.Writer.Status()), zap.Duration("latency", time.Since(started)), zap.String("client_ip", c.ClientIP()))
	}
}
`
}

func bodyLimitMiddlewareFile() string {
	return `package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func BodyLimit(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limit > 0 { c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit) }
		c.Next()
	}
}
`
}

func corsMiddlewareFile() string {
	return `package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins { allowed[origin] = struct{}{} }
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowed[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions && origin != "" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func ParseOrigins(value string) []string {
	if value == "" { return nil }
	return strings.Split(value, ",")
}
`
}

func authMiddlewareFile(modulePath string) string {
	return fmt.Sprintf(`package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"%s/internal/platform/auth"
)

type PrincipalResolver func(token string) (auth.Principal, error)

func Authenticate(resolve PrincipalResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := bearerToken(c.GetHeader("Authorization"))
		if err != nil { c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": err.Error()}); return }
		principal, err := resolve(token)
		if err != nil { c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "invalid token"}); return }
		c.Set("principal", principal)
		c.Next()
	}
}

func bearerToken(header string) (string, error) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" { return "", errors.New("bearer token required") }
	return parts[1], nil
}
`, modulePath)
}

func rateLimitMiddlewareFile() string {
	return `package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Limiter interface { Allow(key string) bool }

func RateLimit(limiter Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter != nil && !limiter.Allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"code": "RATE_LIMITED", "message": "too many requests"})
			return
		}
		c.Next()
	}
}
`
}

func otelMiddlewareFile() string {
	return `package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
)

func OTel(serviceName string) gin.HandlerFunc {
	tracer := otel.Tracer(serviceName)
	return func(c *gin.Context) {
		ctx, span := tracer.Start(c.Request.Context(), c.Request.Method+" "+c.Request.URL.Path)
		defer span.End()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
`
}

func middlewarePlaceholder(name string) string {
	return fmt.Sprintf("package middleware\n\nfunc %sName() string { return %q }\n", name, strings.ToLower(name))
}

func rbacFile() string {
	return `package auth

type Permission string

const (
	PermissionAgentRun    Permission = "agent:run"
	PermissionAgentCancel Permission = "agent:cancel"
	PermissionToolExecute Permission = "tool:execute"
	PermissionToolApprove Permission = "tool:approve"
	PermissionAdminAudit  Permission = "admin:audit"
)
`
}

func principalFile() string {
	return `package auth

type Principal struct {
	UserID      string
	Roles       []string
	Permissions []Permission
}
`
}

func jobFile() string {
	return `package queue

type JobStatus string

const (
	JobQueued    JobStatus = "Queued"
	JobRunning   JobStatus = "Running"
	JobSucceeded JobStatus = "Succeeded"
	JobFailed    JobStatus = "Failed"
)

type Job struct {
	ID     string
	Type   string
	Status JobStatus
}
`
}

func runStateMachineFile() string {
	return `package runs

type State string

const (
	StateQueued          State = "Queued"
	StateRunning         State = "Running"
	StateWaitingApproval State = "WaitingApproval"
	StateRetrying        State = "Retrying"
	StateSucceeded       State = "Succeeded"
	StateFailed          State = "Failed"
	StateCancelled       State = "Cancelled"
)

func CanTransit(from State, to State) bool {
	switch from {
	case StateQueued:
		return to == StateRunning || to == StateCancelled
	case StateRunning:
		return to == StateWaitingApproval || to == StateRetrying || to == StateSucceeded || to == StateFailed || to == StateCancelled
	case StateWaitingApproval:
		return to == StateRunning || to == StateCancelled
	case StateRetrying:
		return to == StateRunning || to == StateFailed || to == StateCancelled
	default:
		return false
	}
}
`
}

func toolRegistryFile() string {
	return `package tools

type Tool interface {
	Name() string
}

type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: map[string]Tool{}}
}

func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}
`
}

func entClientFile() string {
	return `package ent

type Client struct{}

func Open(driver string, dataSourceName string) (*Client, error) {
	_ = driver
	_ = dataSourceName
	return &Client{}, nil
}

func (c *Client) Close() error {
	return nil
}
`
}

func entTestRecordSchema() string {
	return `package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// TestRecord is a small verification entity for the Ent generation pipeline.
type TestRecord struct {
	ent.Schema
}

func (TestRecord) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("status").Default("pending"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (TestRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
	}
}
`
}

func errorsFile() string {
	return `package errors

type Code string

const (
	CodeOK           Code = "OK"
	CodeBadRequest   Code = "BAD_REQUEST"
	CodeUnauthorized Code = "UNAUTHORIZED"
	CodeForbidden    Code = "FORBIDDEN"
	CodeInternal     Code = "INTERNAL"
)
`
}

func responseFile() string {
	return `package response

type Envelope struct {
	Code    string ` + "`json:\"code\"`" + `
	Message string ` + "`json:\"message\"`" + `
	Data    any    ` + "`json:\"data,omitempty\"`" + `
}
`
}
