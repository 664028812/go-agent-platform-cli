package scaffold

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateProject(t *testing.T) {
	root := filepath.Join(t.TempDir(), "platform")
	result, err := GenerateProject(ProjectSpec{
		Name:   "go-agent-platform",
		Module: "example.com/go-agent-platform",
		Dir:    root,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.WrittenCount() == 0 {
		t.Fatal("expected generated files")
	}
	mustExist(t, root, "go.mod")
	mustExist(t, root, "go.sum")
	mustExist(t, root, "Makefile")
	mustExist(t, root, "main.go")
	mustExist(t, root, "internal/platform/cmd/server.go")
	mustExist(t, root, "internal/platform/storage/postgres.go")
	mustExist(t, root, "internal/platform/cache/redis.go")
	mustExist(t, root, "internal/platform/server/http_server.go")
	mustExist(t, root, "internal/platform/middleware/body_limit.go")
	mustExist(t, root, "internal/platform/storage/ent/generate.go")
	mustExist(t, root, "internal/platform/storage/ent/schema/test_record.go")
	mustExist(t, root, "internal/platform/storage/ent/tools.go")
	mustExist(t, root, "internal/platform/tools/tools.go")
	mustExist(t, root, "api/proto/platform/v1/health.proto")
	mustExist(t, root, "buf.yaml")
	mustExist(t, root, "buf.gen.yaml")
	mustExist(t, root, "hack/hack.mk")
	mustExist(t, root, "manifest/config/config.local.yaml")
	mustExist(t, root, "internal/controller/agents/agents_new.go")
	// 种子模块只保留 agents
	mustNotExist(t, root, "internal/controller/runs/runs_new.go")
	mustNotExist(t, root, "internal/controller/auth/auth_new.go")
}

func TestGenerateProjectPreservesHyphenatedName(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "gap-test")
	_, err := GenerateProject(ProjectSpec{Name: "gap-test", Dir: root})
	if err != nil {
		t.Fatal(err)
	}
	mustExist(t, root, "go.mod")

	content, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "module gap-test") {
		t.Fatalf("expected hyphenated module name, got:\n%s", content)
	}
	if !strings.Contains(string(content), "entgo.io/ent v0.14.6") {
		t.Fatalf("expected the default Ent dependency, got:\n%s", content)
	}
	for _, dependency := range []string{
		"github.com/gin-gonic/gin v1.12.0",
		"github.com/jackc/pgx/v5 v5.10.0",
		"github.com/redis/go-redis/v9 v9.22.0",
		"github.com/spf13/viper v1.21.0",
		"go.uber.org/zap v1.27.1",
		"gopkg.in/natefinch/lumberjack.v2 v2.2.1",
	} {
		if !strings.Contains(string(content), dependency) {
			t.Fatalf("expected dependency %q, got:\n%s", dependency, content)
		}
	}
}

func TestGenerateProjectRejectsPathAsName(t *testing.T) {
	_, err := GenerateProject(ProjectSpec{Name: "nested/gap-test", Dir: t.TempDir()})
	if err == nil {
		t.Fatal("expected path-like project name to be rejected")
	}
}

func TestGenerateModule(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/platform\n\ngo 1.25.6\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := GenerateModule(ModuleSpec{Name: "orders", Root: root})
	if err != nil {
		t.Fatal(err)
	}
	if result.WrittenCount() < 10 {
		t.Fatalf("expected a complete standalone module, got %d files", result.WrittenCount())
	}
	mustExist(t, root, "api/orders/orders.go")
	mustExist(t, root, "api/orders/v1/create.go")
	mustExist(t, root, "api/orders/v1/get.go")
	mustExist(t, root, "internal/controller/orders/orders_new.go")
	mustExist(t, root, "internal/controller/orders/orders_v1_create.go")
	mustExist(t, root, "internal/service/orders.go")
	mustExist(t, root, "internal/logic/orders/orders.go")
	mustExist(t, root, "internal/logic/logic.go")
	mustExist(t, root, "internal/model/orders.go")
	mustExist(t, root, "internal/dao/orders.go")
	mustExist(t, root, "internal/platform/storage/ent/client.go")
	mustExist(t, root, "internal/platform/storage/ent/tx.go")
}

func TestGeneratedEntCommandsUsePlatformStoragePath(t *testing.T) {
	root := filepath.Join(t.TempDir(), "platform")
	if _, err := GenerateProject(ProjectSpec{Name: "platform", Dir: root}); err != nil {
		t.Fatal(err)
	}

	hackMakefile, err := os.ReadFile(filepath.Join(root, "hack", "hack.mk"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(hackMakefile)
	for _, expected := range []string{
		"go run -mod=mod entgo.io/ent/cmd/ent init --target internal/platform/storage/ent/schema $${NAME}",
		"go run -mod=mod entgo.io/ent/cmd/ent generate ./internal/platform/storage/ent/schema",
		"Ent schema is empty; skipped Ent generation",
		"proto-gen:",
		"buf generate",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected Ent command %q in generated hack.mk:\n%s", expected, content)
		}
	}

	generateFile, err := os.ReadFile(filepath.Join(root, "internal", "platform", "storage", "ent", "generate.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(generateFile), "//go:generate env GOFLAGS=-mod=mod go run -mod=mod entgo.io/ent/cmd/ent generate ./schema") {
		t.Fatalf("unexpected Ent generator entrypoint:\n%s", generateFile)
	}
}

func TestGenerateCtrlServiceAndDao(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/platform\n\ngo 1.25.6\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := GenerateModule(ModuleSpec{Name: "orders", Root: root}); err != nil {
		t.Fatal(err)
	}
	// ctrl/dao 重生成前必须先有 v1 DTO，这里先 module 建好 payments 再重生成
	if _, err := GenerateModule(ModuleSpec{Name: "payments", Root: root}); err != nil {
		t.Fatal(err)
	}
	if _, err := GenerateCtrl(CtrlSpec{Name: "payments", Root: root, Force: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := GenerateService(ServiceSpec{Name: "orders", Root: root, Force: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := GenerateDao(DaoSpec{Name: "payments", Root: root, Force: true}); err != nil {
		t.Fatal(err)
	}

	mustExist(t, root, "api/payments/payments.go")
	mustExist(t, root, "api/payments/v1/create.go")
	mustExist(t, root, "internal/controller/payments/payments_new.go")
	mustExist(t, root, "internal/dao/payments.go")

	content, err := os.ReadFile(filepath.Join(root, "internal/service/orders.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "type IOrders interface") {
		t.Fatal("expected generated service interface")
	}
	if !strings.Contains(string(content), "Create(ctx context.Context, input model.CreateOrdersInput) (model.OrdersOutput, error)") {
		t.Fatal("expected service interface to include logic method signature")
	}
	if !strings.Contains(string(content), "func RegisterOrders(service IOrders)") {
		t.Fatal("expected GF-style service registration function")
	}
}

func TestGenerateCtrlRequiresExistingDTOs(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/platform\n\ngo 1.25.6\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// 没有 v1 DTO 的模块：重生成必须报错，而不是写入写死的默认 Create/Get
	if _, err := GenerateCtrl(CtrlSpec{Name: "ghost", Root: root}); err == nil {
		t.Fatal("expected error when module has no v1 DTO pairs")
	}
}

func TestGenerateCtrlAllModulesWithoutName(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/platform\n\ngo 1.25.6\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// 两个模块
	for _, m := range []string{"orders", "payments"} {
		if _, err := GenerateModule(ModuleSpec{Name: m, Root: root}); err != nil {
			t.Fatal(err)
		}
	}

	// 不带 -name：全量重生成
	result, err := GenerateCtrl(CtrlSpec{Root: root, Force: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.WrittenCount() == 0 {
		t.Fatal("expected regenerated files")
	}
	mustExist(t, root, "api/orders/orders.go")
	mustExist(t, root, "api/payments/payments.go")
	mustExist(t, root, "internal/controller/orders/orders_new.go")
	mustExist(t, root, "internal/controller/payments/payments_new.go")
}

func TestGenerateAllWithoutNameForcesGeneratedFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/platform\n\ngo 1.25.6\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"orders", "payments"} {
		if _, err := GenerateModule(ModuleSpec{Name: name, Root: root}); err != nil {
			t.Fatal(err)
		}
	}

	ctrlPath := filepath.Join(root, "internal", "controller", "orders", "orders_new.go")
	daoPath := filepath.Join(root, "internal", "dao", "orders.go")
	if err := os.WriteFile(ctrlPath, []byte("package orders\n\nconst staleController = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(daoPath, []byte("package dao\n\nconst staleDAO = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := GenerateCtrl(CtrlSpec{Root: root}); err != nil {
		t.Fatal(err)
	}
	if _, err := GenerateService(ServiceSpec{Root: root}); err != nil {
		t.Fatal(err)
	}
	if _, err := GenerateDao(DaoSpec{Root: root}); err != nil {
		t.Fatal(err)
	}

	ctrlContent, err := os.ReadFile(ctrlPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(ctrlContent), "staleController") || !strings.Contains(string(ctrlContent), "func NewV1") {
		t.Fatalf("ctrl all mode did not refresh %s:\n%s", ctrlPath, ctrlContent)
	}
	daoContent, err := os.ReadFile(daoPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(daoContent), "staleDAO") || !strings.Contains(string(daoContent), "func (d Orders) Create") {
		t.Fatalf("dao all mode did not refresh %s:\n%s", daoPath, daoContent)
	}
	mustExist(t, root, "internal/service/orders.go")
	mustExist(t, root, "internal/service/payments.go")
}

func TestGenerateServiceSyncsLogicMethods(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/platform\n\ngo 1.25.6\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := GenerateModule(ModuleSpec{Name: "orders", Root: root}); err != nil {
		t.Fatal(err)
	}

	// 用户在 logic 中手写一个新方法
	if err := os.WriteFile(filepath.Join(root, "internal", "logic", "orders", "orders_list.go"), []byte(`package orders

import (
	"context"

	"example.com/platform/internal/model"
)

func (l *sOrders) List(ctx context.Context, page int, size int) ([]model.OrdersOutput, error) {
	return nil, nil
}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := GenerateService(ServiceSpec{Name: "orders", Root: root, Force: true}); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(root, "internal", "service", "orders.go"))
	if err != nil {
		t.Fatal(err)
	}
	svc := string(content)
	if !strings.Contains(svc, "List(ctx context.Context, page int, size int) ([]model.OrdersOutput, error)") {
		t.Fatalf("service interface missing hand-written logic method List:\n%s", svc)
	}
	if !strings.Contains(svc, "Create(ctx context.Context, input model.CreateOrdersInput) (model.OrdersOutput, error)") {
		t.Fatalf("service interface missing generated logic method Create:\n%s", svc)
	}
}

func TestGenerateCtrlAddsMethodFromNewDTO(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/platform\n\ngo 1.25.6\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := GenerateModule(ModuleSpec{Name: "agents", Root: root}); err != nil {
		t.Fatal(err)
	}

	// TestReq/TestRes 会生成 Test 方法；文件名不能落为 *_test.go，
	// 否则 Go 会在正常构建中排除 Controller 方法。
	if err := os.WriteFile(filepath.Join(root, "api", "agents", "v1", "test.go"), []byte(`package v1

type TestReq struct {
	ID string
}

type TestRes struct {
	ID string
}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	// 不带 -force：api 接口是契约文件，也必须刷新
	if _, err := GenerateCtrl(CtrlSpec{Name: "agents", Root: root}); err != nil {
		t.Fatal(err)
	}

	// api 接口应包含 Update 方法
	apiContent, err := os.ReadFile(filepath.Join(root, "api", "agents", "agents.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(apiContent), "Test(ctx context.Context, req *v1.TestReq) (res *v1.TestRes, err error)") {
		t.Fatalf("api interface missing Test method:\n%s", apiContent)
	}

	// controller 方法文件必须避开 Go 的 *_test.go 特殊文件名。
	if _, err := os.Stat(filepath.Join(root, "internal", "controller", "agents", "agents_v1_test_handler.go")); err != nil {
		t.Fatalf("expected agents_v1_test_handler.go: %v", err)
	}

	// 全套重生成后应能编译
	if _, err := GenerateService(ServiceSpec{Name: "agents", Root: root}); err != nil {
		t.Fatal(err)
	}
	if _, err := GenerateDao(DaoSpec{Name: "agents", Root: root, Force: true}); err != nil {
		t.Fatal(err)
	}
	assertBuild(t, root)
}

func assertBuild(t *testing.T, root string) {
	t.Helper()
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOCACHE=/tmp/go-agent-platform-cli-test-cache")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
}

func mustExist(t *testing.T, root string, rel string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
		t.Fatalf("%s missing: %v", rel, err)
	}
}

func mustNotExist(t *testing.T, root string, rel string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, rel)); err == nil {
		t.Fatalf("%s should not exist", rel)
	}
}
