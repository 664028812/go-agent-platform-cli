package scaffold

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// fieldInfo 结构体字段（导出字段）
type fieldInfo struct {
	Name string
	Type string
}

// v1Pair 一对 v1 DTO：XxxReq + XxxRes，推导出方法 Xxx
type v1Pair struct {
	Method    string // 如 Create、Get、Test
	ReqName   string
	ResName   string
	ReqFields []fieldInfo
	ResFields []fieldInfo
}

// defaultPairs 新模块默认的 Create/Get 两对 DTO。
// 只用于 module / project 的新建场景，重生成命令不得使用。
func defaultPairs() []v1Pair {
	return []v1Pair{
		{
			Method:    "Create",
			ReqName:   "CreateReq",
			ResName:   "CreateRes",
			ReqFields: []fieldInfo{{Name: "Name", Type: "string"}},
			ResFields: []fieldInfo{{Name: "ID", Type: "string"}},
		},
		{
			Method:    "Get",
			ReqName:   "GetReq",
			ResName:   "GetRes",
			ReqFields: []fieldInfo{{Name: "ID", Type: "string"}},
			ResFields: []fieldInfo{{Name: "ID", Type: "string"}, {Name: "Name", Type: "string"}},
		},
	}
}

// dtoDefaultFiles 新模块的默认 DTO 文件（create.go / get.go）。
// 只用于 module / project 的新建场景，重生成命令不得使用。
func dtoDefaultFiles(name string) map[string]string {
	pkg := sanitizeIdentifier(name)
	return map[string]string{
		filepath.Join("api", pkg, "v1", "create.go"): `package v1

type CreateReq struct {
	Name string ` + "`json:\"name\" binding:\"required\"`" + `
}

type CreateRes struct {
	ID string ` + "`json:\"id\"`" + `
}
`,
		filepath.Join("api", pkg, "v1", "get.go"): `package v1

type GetReq struct {
	ID string ` + "`uri:\"id\" binding:\"required\"`" + `
}

type GetRes struct {
	ID   string ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}
`,
	}
}

// scanV1Pairs 扫描 api/<name>/v1 包，收集 XxxReq + XxxRes 配对。
// v1 目录不存在或为空时返回 nil。
func scanV1Pairs(root string, name string) ([]v1Pair, error) {
	dir := filepath.Join(root, "api", name, "v1")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil
	}

	structs := map[string][]fieldInfo{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, 0)
		if err != nil {
			return nil, err
		}
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				structs[ts.Name.Name] = structFields(fset, st)
			}
		}
	}

	var pairs []v1Pair
	for typeName, reqFields := range structs {
		if !strings.HasSuffix(typeName, "Req") {
			continue
		}
		method := strings.TrimSuffix(typeName, "Req")
		resName := method + "Res"
		resFields, ok := structs[resName]
		if !ok {
			continue
		}
		pairs = append(pairs, v1Pair{
			Method:    method,
			ReqName:   typeName,
			ResName:   resName,
			ReqFields: reqFields,
			ResFields: resFields,
		})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].Method < pairs[j].Method })
	return pairs, nil
}

// requireV1Pairs 重生成命令使用的配对获取入口：
// 只基于实际扫描到的 v1 DTO，没有配对时报错，绝不回落到写死的默认方法。
func requireV1Pairs(root string, name string) ([]v1Pair, error) {
	pairs, err := scanV1Pairs(root, name)
	if err != nil {
		return nil, err
	}
	if len(pairs) == 0 {
		return nil, fmt.Errorf("no v1 DTO pairs found for module %q: add XxxReq/XxxRes to api/%s/v1 or run gap module first", name, name)
	}
	return pairs, nil
}

func structFields(fset *token.FileSet, st *ast.StructType) []fieldInfo {
	var fields []fieldInfo
	for _, f := range st.Fields.List {
		if len(f.Names) == 0 {
			continue // 嵌入式字段跳过
		}
		typ := exprString(fset, f.Type)
		for _, n := range f.Names {
			if !n.IsExported() {
				continue
			}
			fields = append(fields, fieldInfo{Name: n.Name, Type: typ})
		}
	}
	return fields
}

// methodFile 生成单方法文件名，如 agents_v1_create.go。
// Go 会把 *_test.go 作为测试文件排除在正常构建外，因此 Test 方法需要
// 生成 agents_v1_test_handler.go，避免 ControllerV1 在生产构建中缺少方法。
func methodFile(prefix string, method string) string {
	base := prefix + strings.ToLower(method)
	if strings.HasSuffix(base, "_test") {
		base += "_handler"
	}
	return base + ".go"
}
