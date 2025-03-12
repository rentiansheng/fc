package code

import (
	"github.com/pkg/errors"
	"go/ast"
	"go/token"
)

/***************************
    @author: tiansheng.ren
    @date: 2025/3/12
    @desc:

***************************/

type Function struct {
	Name  string
	Decl  ast.Decl
	Start Location
	End   Location
}

type Location struct {
	Line int
	Col  int
}

type FuncCombination struct {
	Name string
	F1   *Function
	F2   *Function
	Same bool
}

type CodeDiff struct {
	f1        *ast.File
	f2        *ast.File
	f1FileSet *token.FileSet
	f2FileSet *token.FileSet
	res       map[string]*FuncCombination
}

func NewCodeDiff(f1, f2 *ast.File, f1FileSet, f2FileSet *token.FileSet) *CodeDiff {
	return &CodeDiff{
		f1:        f1,
		f2:        f2,
		f1FileSet: f1FileSet,
		f2FileSet: f2FileSet,
		res:       make(map[string]*FuncCombination),
	}
}

func (cd *CodeDiff) Diff() (sames []*FuncCombination, diffs []*FuncCombination, err error) {
	if len(cd.res) == 0 {
		err := cd.compare()
		if err != nil {
			return nil, nil, err
		}
	}

	for _, v := range cd.res {
		if v.Same {
			sames = append(sames, v)
		} else {
			diffs = append(diffs, v)
		}
	}

	return sames, diffs, nil
}

func (cd *CodeDiff) compare() error {
	if len(cd.res) > 0 {
		// 已经对比过了。使用缓存结果
		return nil
	}
	f1FnRela, err := cd.fileFunction(cd.f1FileSet, cd.f1.Decls)
	if err != nil {
		return errors.WithMessage(err, "get f1 file function")
	}
	f2FnRela, err := cd.fileFunction(cd.f2FileSet, cd.f2.Decls)
	if err != nil {
		return errors.WithMessage(err, "get f2 file function")
	}
	cd.res = make(map[string]*FuncCombination)
	for k, v := range f1FnRela {
		cd.res[k] = &FuncCombination{
			Name: k,
			F1:   v,
			F2:   f2FnRela[k],
		}
		delete(f2FnRela, k)
	}
	for k, v := range f2FnRela {
		if _, ok := cd.res[k]; !ok {
			cd.res[k] = &FuncCombination{
				Name: k,
				F1:   &Function{},
				F2:   v,
			}
		}
	}
	for k, v := range cd.res {
		v.Same = cd.compareFunc(v.F1, v.F2)
		cd.res[k] = v

	}

	return nil
}

func (cd *CodeDiff) compareFunc(f1, f2 *Function) bool {
	if f1.Name != f2.Name {
		return false
	}
	return DeepEqualForAst(f1.Decl, f2.Decl)
}

func (cd *CodeDiff) fileFunction(fSet *token.FileSet, decls []ast.Decl) (map[string]*Function, error) {
	var e error
	res := make(map[string]*Function)
	for _, decl := range decls {
		switch v := decl.(type) {
		case *ast.FuncDecl:
			var function *Function
			if v.Recv == nil {
				function = globalFuncSelector(fSet, v)
				if function == nil {
					continue
				}
			} else {
				function, e = methodFuncSelector(fSet, v)
				if e != nil {
					return nil, e
				}
			}
			res[function.Name] = function

		case *ast.GenDecl:
			functions := globalAnonymousFuncSelector(fSet, v)
			for _, f := range functions {
				res[f.Name] = f
			}

		}
	}

	return res, nil
}
