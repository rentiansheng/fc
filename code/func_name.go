package code

import (
	"fmt"
	"github.com/pkg/errors"

	"go/ast"
	"go/token"
)

/***************************
    @author: tiansheng.ren
    @date: 2025/3/12
    @desc:

***************************/

func posToLocation(p token.Position) Location {

	return Location{
		Line: p.Line,
		Col:  p.Column,
	}
}

func toLocation(fSet *token.FileSet, body *ast.BlockStmt) (start, end Location) {

	startBody, endBody := fSet.Position(body.Lbrace), fSet.Position(body.Rbrace)
	return posToLocation(startBody), posToLocation(endBody)
}

func globalFuncSelector(fSet *token.FileSet, decl *ast.FuncDecl) (function *Function) {
	if decl.Body == nil {
		return nil
	}

	start, end := toLocation(fSet, decl.Body)
	return &Function{
		Name:  decl.Name.Name,
		Decl:  decl,
		Start: start,
		End:   end,
	}
}

func methodFuncSelector(fSet *token.FileSet, decl *ast.FuncDecl) (function *Function, err error) {
	if len(decl.Recv.List) != 1 {
		return nil, errors.WithStack(fmt.Errorf("method func decl.Recv.List length != 1? decl=%#v", *decl))
	}
	structName := ""
	switch v := decl.Recv.List[0].Type.(type) {
	case *ast.Ident:
		structName = v.Name
	case *ast.StarExpr:
		if vv, ok := v.X.(*ast.Ident); ok {
			structName = "*" + vv.Name
		}
	}
	if structName == "" {
		return nil, errors.WithStack(fmt.Errorf("method struct not found? decl=%#v", *decl))
	}

	fName := decl.Name.Name
	start, end := toLocation(fSet, decl.Body)
	return &Function{
		Name:  structName + "." + fName,
		Decl:  decl,
		Start: start,
		End:   end,
	}, nil
}

func globalAnonymousFuncSelector(fSet *token.FileSet, decl *ast.GenDecl) (functions []*Function) {
	if len(decl.Specs) != 1 {
		return nil
	}
	var spec *ast.ValueSpec
	var ok bool
	if spec, ok = decl.Specs[0].(*ast.ValueSpec); !ok {
		return nil
	}
	for i, value := range spec.Values {
		if v, ok := value.(*ast.FuncLit); ok {
			newDecl := *decl
			newSpec := *spec
			newSpec.Names = []*ast.Ident{newSpec.Names[i]}
			newSpec.Values = []ast.Expr{newSpec.Values[i]}
			newDecl.Specs = []ast.Spec{&newSpec}
			start, end := toLocation(fSet, v.Body)
			functions = append(functions, &Function{
				Name:  spec.Names[i].Name,
				Decl:  &newDecl,
				Start: start,
				End:   end,
			})

		}
	}
	return functions
}
