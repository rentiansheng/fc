package wire

import (
	"context"
	"fmt"
	"github.com/rentiansheng/fc/injection"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"path/filepath"
	"strings"
)

/***************************
    @author: tiansheng.ren
    @date: 2024/12/30
    @desc:

***************************/

func (w *wire) genFile(ctx context.Context, pkg *packages.Package, decls []ast.Decl, wireFilename string) {
	filePkgIdNameRela := w.genFileImport(ctx, decls, wireFilename)
	for _, decl := range decls {
		// 函数内部变量
		varNameRela := map[string]*injection.Inject{}
		switch declVal := decl.(type) {
		case *ast.GenDecl:
			//  单独处理 import 导入
			// 整个文件处理完成才可以缓存文件file imports ,因为import 可能会有多个语句
		case *ast.FuncDecl:
			// 函数处理

			fnRes := w.genFunc(ctx, declVal, pkg, filePkgIdNameRela, varNameRela)
			w.mergeSTypeMeta(fnRes...)
		default:
			println(fmt.Sprintf("gen file decl type not found hande %d", wireFilename))
		}
	}
}

func (w *wire) mergeSTypeMeta(res ...*injection.Inject) {
	for _, item := range res {
		if _, ok := w.inject[item.St]; !ok {
			w.inject[item.St] = []*injection.Inject{}
		}
		w.inject[item.St] = append(w.inject[item.St], item)
	}
}

func (w *wire) genFunc(ctx context.Context, declVal *ast.FuncDecl, pkg *packages.Package, filePkgIdNameRela map[string]string,
	varNameRela map[string]*injection.Inject) []*injection.Inject {

	for _, fnStmt := range declVal.Body.List {
		//println(fnStmt)
		switch fnStmtVal := fnStmt.(type) {
		case *ast.AssignStmt:
			w.genFuncAssignStmt(ctx, fnStmtVal, pkg, filePkgIdNameRela, varNameRela)
		case *ast.ReturnStmt:
			res := w.genFnReturn(ctx, pkg, filePkgIdNameRela, fnStmtVal, varNameRela)
			// 缓存所有struct 实现

			return res
		default:
			println(fnStmt)

		}

	}
	return nil
}

func (w *wire) genFuncAssignStmt(ctx context.Context, fnStmtVal *ast.AssignStmt, pkg *packages.Package,
	filePkgIdNameRela map[string]string,
	varNameRela map[string]*injection.Inject) {

	lValArr := []string{}
	for _, lVal := range fnStmtVal.Lhs {
		if lv, ok := lVal.(*ast.Ident); ok {
			lValArr = append(lValArr, lv.Name)
		} else {
			println("fn res")
		}
	}

	for _, r := range fnStmtVal.Rhs {
		switch rVal := r.(type) {
		case *ast.CallExpr:
			fnRes := w.genFuncCallExpr(ctx, rVal, pkg, filePkgIdNameRela, varNameRela)
			for fnResIdx, fnResItem := range fnRes {
				varNameRela[lValArr[fnResIdx]] = fnResItem
			}

		case *ast.UnaryExpr:
			// 考虑其他package 下struct
			if cls, ok := rVal.X.(*ast.CompositeLit); ok {
				w.genFuncAssignStmtCompositeLit(ctx, fnStmtVal, pkg, filePkgIdNameRela, cls, varNameRela)
			} else {
				println("----")
			}
		case *ast.CompositeLit:
			// 考虑其他package 下struct
			w.genFuncAssignStmtCompositeLit(ctx, fnStmtVal, pkg, filePkgIdNameRela, rVal, varNameRela)
		default:
			println(rVal)
		}
	}

}

func (w *wire) genFnReturn(ctx context.Context, pkg *packages.Package,
	filePkgIdNameRela map[string]string, slVal *ast.ReturnStmt, varNameRela map[string]*injection.Inject) []*injection.Inject {

	res := []*injection.Inject{}
	for _, fnRes := range slVal.Results {
		switch fnResVal := fnRes.(type) {
		case *ast.Ident:
			res = append(res, varNameRela[fnResVal.Name])
		case *ast.UnaryExpr:
			if cls, ok := fnResVal.X.(*ast.CompositeLit); ok {
				sKvRela := w.genFuncCompositeLit(ctx, cls, pkg, filePkgIdNameRela, varNameRela)
				if sKvRela != nil {
					res = append(res, sKvRela)
				}

			} else {
				println("----")
			}
		case *ast.CompositeLit:
			sKvRela := w.genFuncCompositeLit(ctx, fnResVal, pkg, filePkgIdNameRela, varNameRela)
			if sKvRela != nil {
				res = append(res, sKvRela)
			}
		case *ast.SelectorExpr:
			println(slVal)
		default:
			println(slVal)
		}
	}

	return res
}

func (w *wire) genFuncCallExpr(ctx context.Context, rVal *ast.CallExpr, pkg *packages.Package,
	filePkgIdNameRela map[string]string,
	varNameRela map[string]*injection.Inject) []*injection.Inject {
	//res := make(map[types.Type][]*sTypeMeta)
	switch rFnVal := rVal.Fun.(type) {
	case *ast.SelectorExpr:
		return w.genFuncCallSelectorExpr(ctx, rVal, rFnVal, pkg, filePkgIdNameRela, varNameRela)
	case *ast.Ident:
		// struct 需要特殊处理
		callPkgPath := pkg.PkgPath
		callPkgPath = strings.Trim(callPkgPath, "\"")
		asbFnPath := callPkgPath + "." + rFnVal.Name
		if fn, ok := w.s.FunctionRela[asbFnPath]; ok {
			// TODO: arg
			return w.genFuncCall(ctx, pkg, fn.Inst, nil)
		} else {
			println(rFnVal)
		}

	default:
		println(rFnVal)
	}

	return nil
}

func (w *wire) genFuncCall(ctx context.Context, pkg *packages.Package,
	fn *ssa.Function, args []*injection.Inject) []*injection.Inject {
	varNameRela := map[string]*injection.Inject{}

	fnName := fn.String()
	meFn, ok := w.s.FunctionRela[fnName]
	if !ok {
		println(fn.String() + "not found fn")
		return nil
	}

	if pkg == nil {
		return nil
	}
	if len(args) == 0 {

		if fnRes, ok := w.wireFnCallInject[meFn.Index]; ok {
			return fnRes
		}
	} else {
		for idx, arg := range args {
			if arg == nil {
				continue
			}
			varNameRela[fn.Params[idx].Name()] = arg
		}
	}

	if !strings.HasPrefix(fn.Pkg.Pkg.Path(), w.namespace) {
		// 不是当前项目数据，放弃调用
		return nil
	}

	// TODO:  参数需要出来
	fnFile := ""
	// 非当前调用不处理
	if pos := fn.Pkg.Prog.Fset.Position(fn.Pos()); pos.IsValid() {
		fnFile = pos.Filename
	}

	if strings.Contains(fnFile, vendorPart) {
		// vendor 目录中处理
		return nil
	}

	filePkgIdNameRela := make(map[string]string)
	for idx, filename := range pkg.GoFiles {
		if filename == fnFile {
			filePkgIdNameRela = w.genFileImport(ctx, pkg.Syntax[idx].Decls, filename)
			break
		}
	}

	if stmt, ok := fn.Syntax().(*ast.FuncDecl); ok {
		for _, stmtLine := range stmt.Body.List {
			switch slVal := stmtLine.(type) {
			case *ast.AssignStmt:
				if len(slVal.Lhs) == 0 {
					continue
				}
				w.genFuncAssignStmt(ctx, slVal, pkg, filePkgIdNameRela, varNameRela)
			case *ast.ReturnStmt:
				res := w.genFnReturn(ctx, pkg, filePkgIdNameRela, slVal, varNameRela)
				if len(args) == 0 {
					//这里仅处理无参数的函数调用，参数不同，初始化字段累行不行不同
					w.wireFnCallInject[meFn.Index] = res
				}
				return res
			default:
				println(slVal)

			}
		}
	}

	return nil
}

func (w *wire) genFuncAssignStmtCompositeLit(ctx context.Context, fnStmtVal *ast.AssignStmt, pkg *packages.Package, filePkgIdNameRela map[string]string, rVal *ast.CompositeLit, varNameRela map[string]*injection.Inject) {
	sKvRela := w.genFuncCompositeLit(ctx, rVal, pkg, filePkgIdNameRela, varNameRela)
	if varName, ok := fnStmtVal.Lhs[0].(*ast.Ident); ok {
		varNameRela[varName.Name] = sKvRela
	}
}

func (w *wire) genFuncCompositeLit(ctx context.Context, cls *ast.CompositeLit, pkg *packages.Package, filePkgIdNameRela map[string]string, varNameRela map[string]*injection.Inject) *injection.Inject {
	res := &injection.Inject{
		Meta: make(map[string][]*injection.Inject),
	}

	var keyNames []string
	res.St, keyNames = w.genFuncCompositeLitFields(ctx, pkg.PkgPath, cls.Type, filePkgIdNameRela)

	fnKV := func(clVal *ast.KeyValueExpr) {
		// TODO: 注意第三方struct
		// 不支持赋值语句使用函数调用
		if keyName, ok := clVal.Key.(*ast.Ident); ok {
			if valName, ok := clVal.Value.(*ast.Ident); ok {
				if valType, ok := varNameRela[valName.Name]; ok {
					res.Meta[keyName.Name] = append(res.Meta[keyName.Name], valType)
				}
			} else {
				println("CompositeLit KeyValueExpr not ident")
			}
		}
	}

	fnIdent := func(clVal *ast.Ident, idx int) {
		if valType, ok := varNameRela[clVal.Name]; ok {
			res.Meta[keyNames[idx]] = append(res.Meta[keyNames[idx]], valType)
		}
	}
	for idx, cl := range cls.Elts {
		switch clVal := cl.(type) {
		case *ast.KeyValueExpr:
			fnKV(clVal)
		case *ast.Ident:
			fnIdent(clVal, idx)
		case *ast.UnaryExpr:
			// 指针
			switch clVV := clVal.X.(type) {
			case *ast.Ident:
				fnIdent(clVV, idx)
			}
		case *ast.CallExpr:
			callRes := w.genFuncCallExpr(ctx, clVal, pkg, filePkgIdNameRela, varNameRela)
			res.Meta[keyNames[idx]] = append(res.Meta[keyNames[idx]], callRes[0])

		default:
			println("CompositeLit elts type not operate handle")
		}
	}
	return res
}

func (w *wire) genFuncCompositeLitFields(ctx context.Context, curPkg string, clsType ast.Expr, filePkgIdNameRela map[string]string) (
	stIdx uint32, keyNames []string) {

	pkg := curPkg
	var name string
	switch ct := clsType.(type) {
	case *ast.Ident:
		name = ct.Name
	case *ast.SelectorExpr:
		var path string
		path, name = getSelectorExprType(ct)
		if len(path) != 0 {
			if pkgPath, ok := filePkgIdNameRela[path]; ok {
				pkg = pkgPath
			} else {
				println("SelectorExpr pkg path not found ")
			}

		}

	default:
		println("SelectorExpr pkg path empty")
		return
	}

	stName := fmt.Sprintf("%s.%s", pkg, name)
	st := w.s.StructRela[stName]
	if st != nil {
		sType, sTypeOk := st.Inst.Underlying().(*types.Struct)
		if sTypeOk {
			for idx := 0; idx < sType.NumFields(); idx++ {
				keyNames = append(keyNames, sType.Field(idx).Name())
			}
		}
	}

	return
}

func (w *wire) genFuncCallSelectorExpr(ctx context.Context, callExpr *ast.CallExpr, rFnVal *ast.SelectorExpr, pkg *packages.Package,
	filePkgIdNameRela map[string]string,
	varNameRela map[string]*injection.Inject) []*injection.Inject {
	callPkg := pkg
	// 是否为调用点前pkg 对象
	callPkgPath := pkg.PkgPath
	//objectName := ""
	//isPtr := false
	if rFnVal.X != nil {
		if rFnValX, ok := rFnVal.X.(*ast.Ident); ok {
			// 是否为package 对象
			if pkgPath := filePkgIdNameRela[rFnValX.Name]; pkgPath != "" {
				callPkgPath = strings.Trim(pkgPath, "\"")
				callPkg = w.s.ASTPackageRela[callPkgPath]
			}
		}
	}

	callPkgPath = strings.Trim(callPkgPath, "\"")
	if fn, ok := w.s.FunctionRela[callPkgPath+"."+rFnVal.Sel.Name]; ok {
		args := buildCallArgs(callExpr, varNameRela)
		return w.genFuncCall(ctx, callPkg, fn.Inst, args)
	} else {
		println(rFnVal)
	}
	return nil
}

// 处理导入package 与 名字关系
func (s *wire) genImport(ctx context.Context, declVal *ast.GenDecl, filePkgIdNameRela map[string]string) {

	for _, spec := range declVal.Specs {
		if importSpec, ok := spec.(*ast.ImportSpec); ok {
			importPkg := strings.TrimSpace(strings.Trim(importSpec.Path.Value, "\""))
			name := filepath.Base(importPkg)
			if importSpec.Name != nil {
				name = importSpec.Name.Name
			}
			filePkgIdNameRela[name] = importPkg
		}
	}
}

func (w *wire) genFileImport(ctx context.Context, decls []ast.Decl, filename string) map[string]string {

	if cacheInfo, ok := w.fileImportCache[filename]; ok {
		return cacheInfo
	}
	filePkgIdNameRela := make(map[string]string)
	for _, decl := range decls {
		if declVal, ok := decl.(*ast.GenDecl); ok {
			w.genImport(ctx, declVal, filePkgIdNameRela)
		} else {
			break
		}
	}

	w.fileImportCache[filename] = filePkgIdNameRela
	return filePkgIdNameRela
}

func getSelectorExprType(typ *ast.SelectorExpr) (path, name string) {
	if rawName, ok := typ.X.(*ast.Ident); ok {
		path = rawName.Name
	}
	if typ.Sel != nil {
		name = typ.Sel.Name
	}
	return
}

func buildCallArgs(callExpr *ast.CallExpr, varNameRela map[string]*injection.Inject) []*injection.Inject {
	args := make([]*injection.Inject, 0)
	for _, arg := range callExpr.Args {
		if argNameIdent, ok := arg.(*ast.Ident); ok {
			args = append(args, varNameRela[argNameIdent.Name])
		} else {
			args = append(args, nil)
			println(arg)
		}

	}

	return args
}
