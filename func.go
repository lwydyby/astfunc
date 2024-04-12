package astfunc

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/mod/modfile"
)

var BuiltinFuncs = map[string]bool{
	"len":     true,
	"cap":     true,
	"make":    true,
	"new":     true,
	"append":  true,
	"copy":    true,
	"delete":  true,
	"close":   true,
	"complex": true,
	"real":    true,
	"imag":    true,
	"panic":   true,
	"recover": true,
	"print":   true,
	"println": true,
	"int32":   true,
}

var BuiltinModels = map[string]bool{
	"log": true,
	"fmt": true,
}

func GetFuncList(fn *ast.FuncDecl) []string {
	funcName := make([]string, 0)
	hasFunc := map[string]bool{}
	ast.Inspect(fn, func(n ast.Node) bool {
		switch call := n.(type) {
		case *ast.CallExpr:
			switch fun := call.Fun.(type) {
			case *ast.Ident:
				if !BuiltinFuncs[fun.Name] && !hasFunc[fun.Name] {
					hasFunc[fun.Name] = true
					funcName = append(funcName, fun.Name)
				}
			case *ast.SelectorExpr:
				if pkgIdent, ok := fun.X.(*ast.Ident); ok {
					if !BuiltinFuncs[fun.Sel.Name] && !BuiltinModels[pkgIdent.Name] {
						name := fmt.Sprintf("%s.%s", pkgIdent.Name, fun.Sel.Name)
						if hasFunc[name] {
							return true
						}
						hasFunc[name] = true
						funcName = append(funcName, name)
					}
				}
			}
		}
		return true
	})
	return funcName
}

type Value struct {
	Name string
	Type ast.Expr
}

func GetParamList(fn *ast.FuncDecl) []Value {
	paramList := make([]Value, 0)
	if fn == nil || fn.Type == nil || fn.Type.Params == nil {
		return paramList
	}
	for _, param := range fn.Type.Params.List {
		if param.Type == nil {
			continue
		}
		paramList = append(paramList, Value{
			Name: ExprToString(param.Type),
			Type: param.Type,
		})
	}
	return paramList
}

func GetReturnList(fn *ast.FuncDecl) []Value {
	returnList := make([]Value, 0)
	if fn == nil || fn.Type == nil || fn.Type.Params == nil {
		return returnList
	}
	if fn.Type.Results != nil {
		for _, result := range fn.Type.Results.List {
			returnList = append(returnList, Value{
				Name: ExprToString(result.Type),
				Type: result.Type,
			})
		}
	}
	return returnList
}

func ExprToString(expr ast.Expr) string {
	if expr != nil {
		switch e := expr.(type) {
		case *ast.Ident:
			return e.Name
		case *ast.SelectorExpr:
			return ExprToString(e.X) + "." + e.Sel.Name
		case *ast.StarExpr:
			return ExprToString(e.X)
		case *ast.ArrayType:
			return ExprToString(e.Elt)
		default:
			return reflect.TypeOf(expr).String()
		}
	}
	return ""
}

type FuncInfo struct {
	MethodName string
	Params     []string
	Returns    []string
}

func ProcessFunc(names []string, modPath, path string, modFile *modfile.File, imports []*ast.ImportSpec) []FuncInfo {
	code := make([]FuncInfo, 0)
	for i := range names {
		var fn *ast.FuncDecl
		var dir string
		if strings.Contains(names[i], ".") {
			ss := strings.Split(names[i], ".")
			importPath := FindImport(ss[0], imports)
			// 如果找不到说是变量的函数 目前不支持打印
			if len(importPath) == 0 {
				continue
			}
			dir = GetModCachePath(modFile, modPath, importPath)
			if len(dir) == 0 {
				dir = path
			}
			_, fn = FindUsageInDir(dir, ss[1])
		} else {
			dir = path
			_, fn = FindUsageInDir(dir, names[i])
		}
		code = append(code, FuncInfo{
			MethodName: names[i],
			Params:     ProcessStruct(GetParamList(fn), modPath, dir, modFile, imports),
			Returns:    ProcessStruct(GetReturnList(fn), modPath, dir, modFile, imports),
		})
	}
	return code
}

func ProcessStruct(names []Value, modPath, path string, modFile *modfile.File, imports []*ast.ImportSpec) []string {
	code := make([]string, 0)
	for i := range names {
		if strings.Contains(names[i].Name, ".") {
			ss := strings.Split(names[i].Name, ".")
			importPath := FindImport(ss[0], imports)
			if len(importPath) == 0 {
				code = append(code, GetCode(names[i].Type))
				continue
			}
			dir := GetModCachePath(modFile, modPath, importPath)
			if len(dir) == 0 {
				continue
			}
			c := FindStructCode(dir, ss[1])
			if len(c) == 0 {
				code = append(code, GetCode(names[i].Type))
				continue
			}
			code = append(code, c)
			continue
		}
		c := FindStructCode(path, names[i].Name)
		if len(c) == 0 {
			code = append(code, GetCode(names[i].Type))
			continue
		}
		code = append(code, c)
	}
	return code
}

func FindStructCode(dir string, targetVarName string) string {
	var code string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			fmt.Println("Parse error:", err)
			return nil
		}
		var skip bool
		ast.Inspect(node, func(node ast.Node) bool {
			ts, ok := node.(*ast.TypeSpec)
			if !ok {
				return true
			}
			if _, ok := ts.Type.(*ast.StructType); ok && ts.Name.Name == targetVarName {
				skip = true
				code = GetCode(node)
				return false
			}
			return true
		})
		if skip {
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return code
}

func FindImport(name string, imports []*ast.ImportSpec) string {
	for i := range imports {
		imp := imports[i]
		path := strings.ReplaceAll(imp.Path.Value, "\"", "")
		if imp.Name != nil {
			if imp.Name.Name == name {
				return path
			}
			continue
		}
		if path[strings.LastIndex(path, "/")+1:] == name {
			return path
		}
	}
	return ""
}
