package astfunc

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func GetFuncWithDependency(dir string, methodName string) string {
	file, f := FindUsageInDir(dir, methodName)
	modPath, mod := FindGoMod(dir)
	params := ProcessStruct(GetParamList(f), modPath, dir, mod, file.Imports)
	returns := ProcessStruct(GetReturnList(f), modPath, dir, mod, file.Imports)
	funcs := ProcessFunc(GetFuncList(f), modPath, dir, mod, file.Imports)
	data := map[string]interface{}{
		"Mod":     mod.Module.Mod.Path,
		"Package": file.Package,
		"Code":    GetCode(f),
		"Params":  params,
		"Returns": returns,
		"Funcs":   funcs,
	}
	t, err := template.New("code").Parse(tpl)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

// GetCode 打印源码
func GetCode(node ast.Node) string {
	var buf bytes.Buffer
	err := format.Node(&buf, token.NewFileSet(), node)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

// FindUsageInDir 在指定目录中查找函数调用和变量使用
func FindUsageInDir(dir string, methodName string) (*ast.File, *ast.FuncDecl) {
	var fset = token.NewFileSet()
	var astFile *ast.File
	var f *ast.FuncDecl
	// 遍历目录中的所有.go文件
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		astFile, f = FindFunc(fset, path, methodName)
		if f != nil {
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return astFile, f
}

// FindFunc 查找指定名称的方法
func FindFunc(fset *token.FileSet, path string, methodName string) (*ast.File, *ast.FuncDecl) {
	// 解析.go文件
	node, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		panic(err)
	}
	var f *ast.FuncDecl
	// 遍历AST查找函数调用和变量使用
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			name := x.Name.Name
			// 可能会是闭包函数,所以这里必须检查是否含有recv
			if x.Recv != nil && len(x.Recv.List) > 0 {
				starExpr, ok := x.Recv.List[0].Type.(*ast.StarExpr)
				if ok {
					ident, ok := starExpr.X.(*ast.Ident)
					if ok {
						name = ident.Name + "." + name
					}
				}
			}
			// 寻找到匹配的func
			if name == methodName {
				f = x
				return false
			}
		}
		return true
	})

	return node, f
}
