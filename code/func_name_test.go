package code

import (
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

/***************************
    @author: tiansheng.ren
    @date: 2025/3/12
    @desc:

***************************/

func Test_globalFuncSelector(t *testing.T) {
	fileBody := `package code
import "fmt"
func globalFun()  {
	fmt.Println("hello")
}`
	fSet := token.NewFileSet()
	f1, err := parser.ParseFile(fSet, "file.go", fileBody, parser.ParseComments)
	require.NoError(t, err)
	var fn *Function
	for _, decl := range f1.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); !ok {
			continue
		} else {
			fn = globalFuncSelector(fSet, funcDecl)
			break
		}
	}
	if fn == nil {
		t.Errorf("not found global func")
		return
	}

	require.Equal(t, "globalFun", fn.Name)
	require.Equal(t, 3, fn.Start.Line)
	require.Equal(t, 19, fn.Start.Col)
	require.Equal(t, 5, fn.End.Line)
	require.Equal(t, 1, fn.End.Col)
}

func Test_methodFuncSelector(t *testing.T) {
	fileBody := `package code
type  methodFunc struct {
}
func (m *methodFunc) methodFunc()  {
}
func (m methodFunc) methodFunc2()  {
}`
	fSet := token.NewFileSet()
	f1, err := parser.ParseFile(fSet, "file.go", fileBody, parser.ParseComments)
	require.NoError(t, err)
	var fns []*Function
	for _, decl := range f1.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); !ok {
			continue
		} else {
			fn, err := methodFuncSelector(fSet, funcDecl)
			require.NoError(t, err)
			fns = append(fns, fn)
		}
	}

	require.Len(t, fns, 2)
	res := []struct {
		Name string
		val  []int
	}{
		{"*methodFunc.methodFunc", []int{4, 36, 5, 1}},
		{"methodFunc.methodFunc2", []int{6, 36, 7, 1}},
	}
	for idx, fn := range fns {
		val := res[idx]
		require.Equal(t, val.Name, fn.Name)
		require.Equal(t, val.val[0], fn.Start.Line)
		require.Equal(t, val.val[1], fn.Start.Col)
		require.Equal(t, val.val[2], fn.End.Line)
		require.Equal(t, val.val[3], fn.End.Col)
	}

}

func Test_globalAnonymousFuncSelector(t *testing.T) {
	fileBody := `package code
var globalAnonymousFunc = func()  {
}`
	fSet := token.NewFileSet()
	f1, err := parser.ParseFile(fSet, "file.go", fileBody, parser.ParseComments)
	require.NoError(t, err)
	var fns []*Function
	for _, decl := range f1.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); !ok {
			continue
		} else {
			fns = globalAnonymousFuncSelector(fSet, genDecl)
		}
	}
	require.Equal(t, 1, len(fns))
	fn := fns[0]
	require.Equal(t, "globalAnonymousFunc", fn.Name)
	require.Equal(t, 2, fn.Start.Line)
	require.Equal(t, 35, fn.Start.Col)
	require.Equal(t, 3, fn.End.Line)
	require.Equal(t, 1, fn.End.Col)
}
