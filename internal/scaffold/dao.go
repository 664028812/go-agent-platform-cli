package scaffold

import (
	"fmt"
	"path/filepath"
	"strings"
)

// daoFilesFromPairs 生成 dao 桩，方法完全由 v1 DTO 配对驱动。
func daoFilesFromPairs(name string, modulePath string, pairs []v1Pair) map[string]string {
	pkg := sanitizeIdentifier(name)
	title := pascal(pkg)
	var b strings.Builder
	b.WriteString("package dao\n\n")
	b.WriteString("import (\n\t\"context\"\n\n")
	b.WriteString("\t\"" + modulePath + "/internal/model\"\n")
	b.WriteString("\tentstore \"" + modulePath + "/internal/platform/storage/ent\"\n)\n\n")
	b.WriteString("type " + title + " struct {\n\tclient *entstore.Client\n}\n\n")
	b.WriteString("func New" + title + "(client *entstore.Client) " + title + " {\n\treturn " + title + "{client: client}\n}\n\n")
	for _, p := range pairs {
		b.WriteString("func (d " + title + ") " + p.Method + "(ctx context.Context, input model." + p.Method + title + "Input) (model." + title + "Output, error) {\n")
		b.WriteString("\t_ = ctx\n\t_ = d.client\n\t_ = input\n")
		b.WriteString("\treturn model." + title + "Output{}, nil\n}\n\n")
	}
	return map[string]string{filepath.Join("internal", "dao", pkg+".go"): b.String()}
}

// logicFilesFromPairs 生成 logic 注册与桩方法。
func logicFilesFromPairs(name string, modulePath string, pairs []v1Pair) map[string]string {
	pkg := sanitizeIdentifier(name)
	title := pascal(pkg)
	files := map[string]string{
		filepath.Join("internal", "logic", pkg, pkg+".go"): fmt.Sprintf(`package %s

import (
	"%s/internal/dao"
	"%s/internal/service"
)

type s%s struct {
	dao dao.%s
}

func New() *s%s {
	return &s%s{}
}

func init() {
	service.Register%s(New())
}
`, pkg, modulePath, modulePath, title, title, title, title, title),
	}
	for _, p := range pairs {
		files[filepath.Join("internal", "logic", pkg, methodFile(pkg+"_", p.Method))] = fmt.Sprintf(`package %s

import (
	"context"

	"%s/internal/model"
)

func (l *s%s) %s(ctx context.Context, input model.%s%sInput) (model.%sOutput, error) {
	return l.dao.%s(ctx, input)
}
`, pkg, modulePath, title, p.Method, p.Method, title, title, p.Method)
	}
	return files
}

// seedModuleFiles 生成新模块默认 Create/Get 两对的完整分层文件。
// 只用于 project 命令的种子模块和 module 命令（新建场景）。
func seedModuleFiles(name string, modulePath string) map[string]string {
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
	return files
}
