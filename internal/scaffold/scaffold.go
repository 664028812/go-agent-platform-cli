package scaffold

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ---------- 命令规格 ----------

type ProjectSpec struct {
	Name   string
	Module string
	Dir    string
	Force  bool
}

type ModuleSpec struct {
	Name   string
	Root   string
	Module string
	Force  bool
}

type CtrlSpec = ModuleSpec
type ServiceSpec = ModuleSpec
type DaoSpec = ModuleSpec

// ---------- 生成结果 ----------

type FileResult struct {
	Path   string
	Status string
}

type Result struct {
	Root  string
	Files []FileResult
}

func (r Result) WrittenCount() int {
	total := 0
	for _, file := range r.Files {
		if file.Status == "write" {
			total++
		}
	}
	return total
}

func (r Result) SkippedCount() int {
	total := 0
	for _, file := range r.Files {
		if file.Status == "skip" {
			total++
		}
	}
	return total
}

// ---------- 生成入口 ----------

// GenerateProject 生成完整项目骨架（含 6 个种子模块）
func GenerateProject(spec ProjectSpec) (Result, error) {
	name := strings.TrimSpace(spec.Name)
	if name == "" {
		name = "go-agent-platform"
	}
	if name == "." || name == ".." || filepath.Base(name) != name || strings.ContainsAny(name, `/\\`) {
		return Result{}, fmt.Errorf("invalid project name %q: use a directory name without path separators", name)
	}
	dir := spec.Dir
	if dir == "" {
		dir = name
	}
	modulePath := strings.TrimSpace(spec.Module)
	if modulePath == "" {
		modulePath = name
	}

	result := Result{Root: dir}
	for _, dirName := range projectDirs() {
		if err := os.MkdirAll(filepath.Join(dir, dirName), 0o755); err != nil {
			return result, err
		}
	}

	files := projectFiles(name, modulePath)
	// 种子模块只保留 agents 一个示例；更多模块用 gap module 按需添加
	for _, moduleName := range []string{"agents"} {
		for rel, content := range seedModuleFiles(moduleName, modulePath) {
			files[rel] = content
		}
	}

	paths := sortedKeys(files)
	for _, rel := range paths {
		status, err := writeFile(filepath.Join(dir, rel), files[rel], spec.Force)
		if err != nil {
			return result, err
		}
		result.Files = append(result.Files, FileResult{Path: rel, Status: status})
	}
	return result, nil
}

// GenerateModule 生成一个完整新模块（含默认 Create/Get DTO）
func GenerateModule(spec ModuleSpec) (Result, error) {
	root, name, modulePath, err := normalizeModuleSpec(spec)
	if err != nil {
		return Result{Root: root}, err
	}
	pairs := defaultPairs()
	files := map[string]string{}
	for rel, content := range dtoDefaultFiles(name) {
		files[rel] = content
	}
	for rel, content := range ctrlFilesFromPairs(name, modulePath, pairs) {
		files[rel] = content
	}
	for rel, content := range serviceFilesFromPairs(name, modulePath, pairs) {
		files[rel] = content
	}
	for rel, content := range modelFilesFromPairs(name, pairs) {
		files[rel] = content
	}
	for rel, content := range daoFilesFromPairs(name, modulePath, pairs) {
		files[rel] = content
	}
	for rel, content := range logicFilesFromPairs(name, modulePath, pairs) {
		files[rel] = content
	}
	files[filepath.Join("internal", "platform", "storage", "ent", "client.go")] = entClientFile()
	files[filepath.Join("internal", "platform", "storage", "ent", "tx.go")] = "package ent\n\ntype Tx struct{}\n"

	result, err := writeFiles(root, files, spec.Force)
	if err != nil {
		return result, err
	}
	// 注册表必须在 logic 目录落盘后生成，才能扫到新模块
	status, err := writeFile(filepath.Join(root, "internal", "logic", "logic.go"), logicRegistryFile(root, modulePath), true)
	if err != nil {
		return result, err
	}
	result.Files = append(result.Files, FileResult{Path: filepath.Join("internal", "logic", "logic.go"), Status: status})
	return result, nil
}

// GenerateCtrl 生成 API 聚合接口与 Controller。
// 方法完全由 api/<name>/v1 的 DTO 配对驱动，Name 为空时遍历所有模块。
// api 聚合接口是契约文件，总是刷新；省略 Name 时表示全量强制刷新，
// controller 方法等生成文件也会覆盖。
func GenerateCtrl(spec CtrlSpec) (Result, error) {
	root, modulePath, err := resolveRootModule(spec.Root)
	if err != nil {
		return Result{Root: root}, err
	}
	names, err := resolveNames(spec.Name, root)
	if err != nil {
		return Result{Root: root}, err
	}
	implFiles := map[string]string{}
	contractFiles := map[string]string{}
	for _, name := range names {
		pairs, err := requireV1Pairs(root, name)
		if err != nil {
			return Result{Root: root}, err
		}
		for rel, content := range ctrlFilesFromPairs(name, modulePath, pairs) {
			if strings.HasPrefix(rel, "api/") {
				contractFiles[rel] = content
			} else {
				implFiles[rel] = content
			}
		}
	}
	force := spec.Force || strings.TrimSpace(spec.Name) == ""
	result, err := writeFiles(root, implFiles, force)
	if err != nil {
		return result, err
	}
	for _, rel := range sortedKeys(contractFiles) {
		status, err := writeFile(filepath.Join(root, rel), contractFiles[rel], true)
		if err != nil {
			return result, err
		}
		result.Files = append(result.Files, FileResult{Path: rel, Status: status})
	}
	return result, nil
}

// GenerateService 生成 model 输入输出与 Service 接口。
// model 由 api/<name>/v1 的 DTO 配对驱动；
// 接口方法扫描 internal/logic/<name> 的公开方法（logic 不存在时回退 v1 推导），
// 因此用户在 logic 中手写的新方法会自动同步进接口。Name 为空时遍历所有模块。
// 生成内容均为契约文件，总是刷新。
func GenerateService(spec ServiceSpec) (Result, error) {
	root, modulePath, err := resolveRootModule(spec.Root)
	if err != nil {
		return Result{Root: root}, err
	}
	names, err := resolveNames(spec.Name, root)
	if err != nil {
		return Result{Root: root}, err
	}
	files := map[string]string{}
	for _, name := range names {
		pairs, err := scanV1Pairs(root, name)
		if err != nil {
			return Result{Root: root}, err
		}
		if len(pairs) > 0 {
			for rel, content := range modelFilesFromPairs(name, pairs) {
				files[rel] = content
			}
		}
		content, err := serviceFromLogic(root, name, modulePath)
		if err != nil {
			// logic 尚不存在（例如只跑了 ctrl）：按 v1 DTO 推导接口
			if len(pairs) > 0 {
				for rel, content := range serviceFilesFromPairs(name, modulePath, pairs) {
					files[rel] = content
				}
				continue
			}
			return Result{Root: root}, err
		}
		files[filepath.Join("internal", "service", name+".go")] = content
	}
	return writeFiles(root, files, true)
}

// GenerateDao 生成 dao 与 logic 桩代码，并同步 logic 注册表。
// 方法完全由 api/<name>/v1 的 DTO 配对驱动，Name 为空时遍历所有模块并强制覆盖生成文件。
func GenerateDao(spec DaoSpec) (Result, error) {
	root, modulePath, err := resolveRootModule(spec.Root)
	if err != nil {
		return Result{Root: root}, err
	}
	names, err := resolveNames(spec.Name, root)
	if err != nil {
		return Result{Root: root}, err
	}
	files := map[string]string{}
	for _, name := range names {
		pairs, err := requireV1Pairs(root, name)
		if err != nil {
			return Result{Root: root}, err
		}
		for rel, content := range daoFilesFromPairs(name, modulePath, pairs) {
			files[rel] = content
		}
		for rel, content := range logicFilesFromPairs(name, modulePath, pairs) {
			files[rel] = content
		}
	}
	force := spec.Force || strings.TrimSpace(spec.Name) == ""
	result, err := writeFiles(root, files, force)
	if err != nil {
		return result, err
	}
	status, err := writeFile(filepath.Join(root, "internal", "logic", "logic.go"), logicRegistryFile(root, modulePath), true)
	if err != nil {
		return result, err
	}
	result.Files = append(result.Files, FileResult{Path: filepath.Join("internal", "logic", "logic.go"), Status: status})
	return result, nil
}

// ---------- 路径与模块解析 ----------

func normalizeModuleSpec(spec ModuleSpec) (string, string, string, error) {
	root := spec.Root
	if root == "" {
		root = "."
	}
	name := sanitizeIdentifier(spec.Name)
	if name == "" {
		return root, "", "", fmt.Errorf("module name is required")
	}
	modulePath := strings.TrimSpace(spec.Module)
	if modulePath == "" {
		var err error
		modulePath, err = readModulePath(filepath.Join(root, "go.mod"))
		if err != nil {
			return root, name, "", err
		}
	}
	return root, name, modulePath, nil
}

// resolveRootModule 解析项目根与 go.mod 模块路径
func resolveRootModule(root string) (string, string, error) {
	if root == "" {
		root = "."
	}
	modulePath, err := readModulePath(filepath.Join(root, "go.mod"))
	if err != nil {
		return root, "", err
	}
	return root, modulePath, nil
}

// resolveNames 解析目标模块列表。
// name 为空时扫描 api/ 下所有模块，用于全量重生成。
func resolveNames(name string, root string) ([]string, error) {
	if strings.TrimSpace(name) != "" {
		clean := sanitizeIdentifier(name)
		if clean == "" {
			return nil, fmt.Errorf("invalid module name %q", name)
		}
		return []string{clean}, nil
	}
	entries, err := os.ReadDir(filepath.Join(root, "api"))
	if err != nil {
		return nil, fmt.Errorf("cannot discover modules under %s/api: %w", root, err)
	}
	var names []string
	for _, entry := range entries {
		// openapi 为历史遗留目录，不作为业务模块
		if entry.IsDir() && sanitizeIdentifier(entry.Name()) != "" && entry.Name() != "openapi" {
			names = append(names, entry.Name())
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no modules found under %s/api", root)
	}
	sort.Strings(names)
	return names, nil
}

// ---------- 文件写入 ----------

func writeFiles(root string, files map[string]string, force bool) (Result, error) {
	result := Result{Root: root}
	for _, rel := range sortedKeys(files) {
		status, err := writeFile(filepath.Join(root, rel), files[rel], force)
		if err != nil {
			return result, err
		}
		result.Files = append(result.Files, FileResult{Path: rel, Status: status})
	}
	return result, nil
}

func writeFile(path string, content string, force bool) (string, error) {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return "skip", nil
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	return "write", os.WriteFile(path, []byte(content), 0o644)
}

func readModulePath(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && fields[0] == "module" {
			return fields[1], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("module path not found in %s", path)
}

func sortedKeys(files map[string]string) []string {
	keys := make([]string, 0, len(files))
	for key := range files {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// ---------- 命名工具 ----------

func sanitizeIdentifier(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	name = strings.ReplaceAll(name, "-", "_")
	re := regexp.MustCompile(`[^a-z0-9_]+`)
	name = re.ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")
	if name == "" {
		return ""
	}
	if name[0] >= '0' && name[0] <= '9' {
		name = "m_" + name
	}
	return name
}

func pascal(name string) string {
	parts := strings.Split(sanitizeIdentifier(name), "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}

// exprString 用 go/printer 把 AST 表达式还原为源码字符串
func exprString(fset *token.FileSet, expr ast.Expr) string {
	var buf bytes.Buffer
	_ = printer.Fprint(&buf, fset, expr)
	return buf.String()
}
