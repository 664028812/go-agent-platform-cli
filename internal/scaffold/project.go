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
		"internal/platform/auth",
		"internal/platform/cache",
		"internal/platform/queue",
		"internal/platform/observability",
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
		"go.mod":                             projectGoMod(modulePath),
		"go.sum":                             projectGoSum(),
		"README.md":                          projectReadme(name),
		"Makefile":                           projectMakefile(),
		"hack/hack.mk":                       projectHackMakefile(),
		"hack/hack-cli.mk":                   projectHackCLIMakefile(),
		".env.example":                       projectEnvExample(),
		"manifest/config/config.local.yaml":  projectConfig("local"),
		"manifest/config/config.dev.yaml":    projectConfig("dev"),
		"manifest/config/config.prod.yaml":   projectConfig("prod"),
		"manifest/docker/docker-compose.yml": projectCompose(),
		"docs/README.md":                     docsReadme(name),
		"migrations/.gitkeep":                "",
		"main.go":                            standaloneMainFile(modulePath),
		"cmd/worker/main.go":                 mainFile(modulePath, "Worker", "RunWorker"),
		"cmd/mcp-server/main.go":             mainFile(modulePath, "MCPServer", "RunMCPServer"),
		"internal/platform/cmd/server.go":    serverCommandFile(modulePath),
		"internal/platform/consts/app.go":    "package consts\n\nconst ServiceName = \"" + name + "\"\n",
		"internal/platform/cmd/worker.go":    workerCommandFile(modulePath),
		"internal/platform/cmd/mcp.go":       mcpCommandFile(modulePath),
		"internal/platform/config/config.go": configTypes(),
		"internal/platform/config/loader.go": configLoader(),
		"internal/platform/server/router.go": `package server

type Route struct {
	Method string
	Path   string
}

func Routes() []Route {
	return []Route{
		{Method: "GET", Path: "/healthz"},
		{Method: "GET", Path: "/readyz"},
		{Method: "POST", Path: "/api/v1/runs"},
	}
}
`,
		"internal/platform/server/http_server.go": `package server

type HTTPServer struct {
	Addr string
}

func NewHTTPServer(addr string) HTTPServer {
	return HTTPServer{Addr: addr}
}
`,
		"internal/platform/server/response.go": `package server

type Response struct {
	Code    string ` + "`json:\"code\"`" + `
	Message string ` + "`json:\"message\"`" + `
	Data    any    ` + "`json:\"data,omitempty\"`" + `
}
`,
		"internal/platform/middleware/request_id.go": `package middleware` + "\n\n" + `const RequestIDHeader = "X-Request-ID"` + "\n",
		"internal/platform/middleware/recover.go":    middlewarePlaceholder("Recover"),
		"internal/platform/middleware/timeout.go":    middlewarePlaceholder("Timeout"),
		"internal/platform/middleware/access_log.go": middlewarePlaceholder("AccessLog"),
		"internal/platform/middleware/auth.go":       middlewarePlaceholder("Auth"),
		"internal/platform/middleware/rate_limit.go": middlewarePlaceholder("RateLimit"),
		"internal/platform/middleware/otel.go":       middlewarePlaceholder("OTel"),
		"internal/platform/auth/jwt.go":              `package auth` + "\n\n" + `type TokenManager struct { Secret string }` + "\n",
		"internal/platform/auth/rbac.go":             rbacFile(),
		"internal/platform/auth/principal.go":        principalFile(),
		"internal/platform/cache/redis.go":           `package cache` + "\n\n" + `type RedisClient struct { Addr string }` + "\n",
		"internal/platform/cache/lock.go":            `package cache` + "\n\n" + `type Lock struct { Key string }` + "\n",
		"internal/platform/cache/rate_limiter.go":    `package cache` + "\n\n" + `type RateLimiter struct { RequestsPerMinute int }` + "\n",
		"internal/platform/cache/cache_aside.go":     `package cache` + "\n\n" + `type CacheAside struct{}` + "\n",
		"internal/platform/queue/job.go":             jobFile(),
		"internal/platform/queue/job_store.go":       `package queue` + "\n\n" + `type JobStore interface { Enqueue(job Job) error }` + "\n",
		"internal/platform/queue/dispatcher.go":      `package queue` + "\n\n" + `type Dispatcher struct{}` + "\n",
		"internal/platform/queue/worker_pool.go":     `package queue` + "\n\n" + `type WorkerPool struct { Concurrency int }` + "\n",
		"internal/platform/queue/retry.go":           `package queue` + "\n\n" + `type RetryPolicy struct { MaxRetries int }` + "\n",
		"internal/platform/queue/dead_letter.go":     `package queue` + "\n\n" + `type DeadLetter struct { JobID string; Error string }` + "\n",
		"internal/platform/observability/logger.go":  `package observability` + "\n\n" + `type Logger struct { ServiceName string }` + "\n",
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
		"internal/platform/storage/ent/client.go": entClientFile(),
		"internal/platform/storage/ent/tx.go":     `package ent` + "\n\n" + `type Tx struct{}` + "\n",
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

require entgo.io/ent v0.14.6
`, modulePath)
}

func projectGoSum() string {
	return `entgo.io/ent v0.14.6 h1:/f2696BpwuWAEEG6PVGWflg6+Inrpq4pRWuNlWz/Skk=
entgo.io/ent v0.14.6/go.mod h1:z46QBUdGC+BATwsedbDuREfSS0oSCV+csdEYlL4p73s=
`
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
Observability: slog + OpenTelemetry + Prometheus
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
manifest/                # config and Docker Compose
hack/                    # Make targets, following earn/app/server
~~~

`+"运行 `make deps` 后可执行 `make ent-gen`，验证默认的 `TestRecord` schema 已生成 Ent 代码。\n", name)
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

.PHONY: build ctrl service dao fmt test run deps ent-new ent-gen compose-up compose-down clean

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
	go get go.opentelemetry.io/otel
	go get github.com/prometheus/client_golang/prometheus

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

database:
  driver: postgres
  dsn: postgres://agent:agent@localhost:5432/agent_platform?sslmode=disable
  max_open_conns: 20
  max_idle_conns: 10
  conn_max_lifetime: 1h

redis:
  addr: localhost:6379
  db: 0

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

	"%s/internal/platform/cmd"
	_ "%s/internal/logic"
)

func main() {
	if err := cmd.Run(context.Background()); err != nil {
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

	"%s/internal/platform/cmd"
)

func main() {
	if err := cmd.%s(context.Background()); err != nil {
		log.Fatal(err)
	}
}
`, modulePath, runFunc)
}

func serverCommandFile(modulePath string) string {
	return fmt.Sprintf(`package cmd

import (
	"context"

	"%s/internal/platform/config"
)

func Run(ctx context.Context) error {
	_, err := config.Load()
	_ = ctx
	return err
}
`, modulePath)
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
	App           AppConfig
	HTTP          HTTPConfig
	Database      DatabaseConfig
	Redis         RedisConfig
	Auth          AuthConfig
	Worker        WorkerConfig
	Eino          EinoConfig
	Observability ObservabilityConfig
}

type AppConfig struct {
	Name            string
	Env             string
	Port            int
	ShutdownTimeout time.Duration
}

type HTTPConfig struct {
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	RequestTimeout time.Duration
	BodyLimit      int64
}

type DatabaseConfig struct {
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Addr string
	DB   int
}

type AuthConfig struct {
	JWTSecret      string
	AccessTokenTTL time.Duration
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
	ServiceName  string
	OTLPEndpoint string
	MetricsPort  int
}
`
}

func configLoader() string {
	return `package config

import (
	"os"
	"strconv"
	"time"
)

func Load() (Config, error) {
	port := envInt("APP_PORT", 8080)
	return Config{
		App: AppConfig{
			Name:            env("APP_NAME", "go-agent-platform"),
			Env:             env("APP_ENV", "local"),
			Port:            port,
			ShutdownTimeout: 10 * time.Second,
		},
		Database: DatabaseConfig{
			Driver: "postgres",
			DSN:    env("DATABASE_DSN", "postgres://agent:agent@localhost:5432/agent_platform?sslmode=disable"),
		},
		Redis: RedisConfig{Addr: env("REDIS_ADDR", "localhost:6379")},
		Auth:  AuthConfig{JWTSecret: env("JWT_SECRET", "local-dev-secret"), AccessTokenTTL: 2 * time.Hour},
		Worker: WorkerConfig{
			WorkerID:     env("WORKER_ID", "local-worker-1"),
			Concurrency:  8,
			PollInterval: time.Second,
			LockTTL:      time.Minute,
			MaxRetries:   3,
		},
		Eino: EinoConfig{
			Provider: env("EINO_PROVIDER", "mock"),
			Model:    env("EINO_MODEL", "mock-chat"),
			Timeout:  time.Minute,
			MaxSteps: 8,
		},
		Observability: ObservabilityConfig{
			ServiceName:  env("OTEL_SERVICE_NAME", "go-agent-platform"),
			OTLPEndpoint: env("OTLP_ENDPOINT", "localhost:4317"),
			MetricsPort:  9090,
		},
	}, nil
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
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

func Open(dsn string) (*Client, error) {
	_ = dsn
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
