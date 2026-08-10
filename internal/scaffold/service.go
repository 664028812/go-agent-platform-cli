package scaffold

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// serviceMethod 从 logic 扫描出的一个公开方法
type serviceMethod struct {
	name      string
	signature string
	imports   map[string]string
}

// serviceFromLogic 扫描 internal/logic/<name> 的公开方法，生成 Service 接口。
// 用户在 logic 中手写或新增的方法会自动同步进接口。
func serviceFromLogic(root string, name string, modulePath string) (string, error) {
	logicDir := filepath.Join(root, "internal", "logic", name)
	entries, err := os.ReadDir(logicDir)
	if err != nil {
		return "", fmt.Errorf("read logic dir %s: %w", logicDir, err)
	}

	fset := token.NewFileSet()
	methods := make([]serviceMethod, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(logicDir, entry.Name()), nil, 0)
		if err != nil {
			return "", err
		}
		imports := fileImportMap(file)
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || !isLogicMethod(fn, name) || !ast.IsExported(fn.Name.Name) {
				continue
			}
			signature := fn.Name.Name + funcTypeString(fset, fn.Type)
			methods = append(methods, serviceMethod{
				name:      fn.Name.Name,
				signature: signature,
				imports:   importsForSignature(signature, imports),
			})
		}
	}
	if len(methods) == 0 {
		return "", fmt.Errorf("no exported Logic methods found in %s", logicDir)
	}

	sort.Slice(methods, func(i, j int) bool { return methods[i].name < methods[j].name })

	neededImports := map[string]string{}
	for _, method := range methods {
		for qualifier, line := range method.imports {
			neededImports[qualifier] = line
		}
	}
	if strings.Contains(strings.Join(methodSignatures(methods), "\n"), "model.") {
		neededImports["model"] = strconv.Quote(modulePath + "/internal/model")
	}

	var builder strings.Builder
	builder.WriteString("package service\n\n")
	if len(neededImports) > 0 {
		builder.WriteString("import (\n")
		for _, qualifier := range sortedKeys(neededImports) {
			builder.WriteString("\t")
			builder.WriteString(neededImports[qualifier])
			builder.WriteString("\n")
		}
		builder.WriteString(")\n\n")
	}
	title := pascal(name)
	builder.WriteString("type I" + title + " interface {\n")
	for _, method := range methods {
		builder.WriteString("\t" + method.signature + "\n")
	}
	builder.WriteString("}\n\nvar local" + title + " I" + title + "\n\n")
	builder.WriteString("func " + title + "() I" + title + " {\n\tif local" + title + " == nil {\n\t\tpanic(\"service implementation I" + title + " is not registered\")\n\t}\n\treturn local" + title + "\n}\n\n")
	builder.WriteString("func Register" + title + "(service I" + title + ") {\n\tlocal" + title + " = service\n}\n")

	formatted, err := format.Source([]byte(builder.String()))
	if err != nil {
		return builder.String(), nil
	}
	return string(formatted), nil
}

// serviceFilesFromPairs 按 v1 DTO 配对推导 Service 接口。
// 仅在 logic 目录尚不存在时作为回退（新建模块先跑 ctrl 的场景）。
func serviceFilesFromPairs(name string, modulePath string, pairs []v1Pair) map[string]string {
	pkg := sanitizeIdentifier(name)
	title := pascal(pkg)
	var b strings.Builder
	b.WriteString("package service\n\n")
	b.WriteString("import (\n\t\"context\"\n\n\t\"" + modulePath + "/internal/model\"\n)\n\n")
	b.WriteString("type I" + title + " interface {\n")
	for _, p := range pairs {
		b.WriteString("\t" + p.Method + "(ctx context.Context, input model." + p.Method + title + "Input) (model." + title + "Output, error)\n")
	}
	b.WriteString("}\n\n")
	b.WriteString("var local" + title + " I" + title + "\n\n")
	b.WriteString("func " + title + "() I" + title + " {\n\tif local" + title + " == nil {\n\t\tpanic(\"service implementation I" + title + " is not registered\")\n\t}\n\treturn local" + title + "\n}\n\n")
	b.WriteString("func Register" + title + "(service I" + title + ") {\n\tlocal" + title + " = service\n}\n")
	return map[string]string{filepath.Join("internal", "service", pkg+".go"): b.String()}
}

// modelFilesFromPairs 生成每个方法一个输入结构 + 模块统一输出结构。
func modelFilesFromPairs(name string, pairs []v1Pair) map[string]string {
	pkg := sanitizeIdentifier(name)
	title := pascal(pkg)
	var b strings.Builder
	b.WriteString("package model\n\n")
	for _, p := range pairs {
		b.WriteString("type " + p.Method + title + "Input struct {\n")
		for _, f := range p.ReqFields {
			b.WriteString("\t" + f.Name + " " + f.Type + "\n")
		}
		b.WriteString("}\n\n")
	}
	b.WriteString("type " + title + "Output struct {\n")
	seen := map[string]bool{}
	for _, p := range pairs {
		for _, f := range p.ResFields {
			if seen[f.Name] {
				continue
			}
			seen[f.Name] = true
			b.WriteString("\t" + f.Name + " " + f.Type + "\n")
		}
	}
	b.WriteString("}\n")
	return map[string]string{filepath.Join("internal", "model", pkg+".go"): b.String()}
}

// ---------- logic 方法扫描辅助 ----------

func isLogicMethod(fn *ast.FuncDecl, name string) bool {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return false
	}
	switch recv := fn.Recv.List[0].Type.(type) {
	case *ast.Ident:
		return recv.Name == "Logic" || recv.Name == "s"+pascal(name)
	case *ast.StarExpr:
		ident, ok := recv.X.(*ast.Ident)
		return ok && (ident.Name == "Logic" || ident.Name == "s"+pascal(name))
	default:
		return false
	}
}

func funcTypeString(fset *token.FileSet, typ *ast.FuncType) string {
	params := fieldListString(fset, typ.Params)
	results := resultListString(fset, typ.Results)
	if results == "" {
		return "(" + params + ")"
	}
	return "(" + params + ") " + results
}

func fieldListString(fset *token.FileSet, fields *ast.FieldList) string {
	if fields == nil || len(fields.List) == 0 {
		return ""
	}
	parts := make([]string, 0, len(fields.List))
	for _, field := range fields.List {
		typ := exprString(fset, field.Type)
		if len(field.Names) == 0 {
			parts = append(parts, typ)
			continue
		}
		names := make([]string, 0, len(field.Names))
		for _, name := range field.Names {
			names = append(names, name.Name)
		}
		parts = append(parts, strings.Join(names, ", ")+" "+typ)
	}
	return strings.Join(parts, ", ")
}

func resultListString(fset *token.FileSet, fields *ast.FieldList) string {
	if fields == nil || len(fields.List) == 0 {
		return ""
	}
	if len(fields.List) == 1 && len(fields.List[0].Names) == 0 {
		return exprString(fset, fields.List[0].Type)
	}
	return "(" + fieldListString(fset, fields) + ")"
}

func fileImportMap(file *ast.File) map[string]string {
	imports := map[string]string{}
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		qualifier := importQualifier(spec, path)
		if qualifier == "" {
			continue
		}
		imports[qualifier] = importLine(spec, path)
	}
	return imports
}

func importsForSignature(signature string, imports map[string]string) map[string]string {
	needed := map[string]string{}
	for qualifier, line := range imports {
		if strings.Contains(signature, qualifier+".") {
			needed[qualifier] = line
		}
	}
	return needed
}

func importQualifier(spec *ast.ImportSpec, path string) string {
	if spec.Name != nil {
		if spec.Name.Name == "." || spec.Name.Name == "_" {
			return ""
		}
		return spec.Name.Name
	}
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	return strings.ReplaceAll(base, "-", "_")
}

func importLine(spec *ast.ImportSpec, path string) string {
	if spec.Name != nil {
		return spec.Name.Name + " " + strconv.Quote(path)
	}
	return strconv.Quote(path)
}

func methodSignatures(methods []serviceMethod) []string {
	signatures := make([]string, 0, len(methods))
	for _, method := range methods {
		signatures = append(signatures, method.signature)
	}
	return signatures
}
