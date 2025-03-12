package code

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_DeepEqualForAstSame(t *testing.T) {
	fileBody1 := `package code
import "fmt"
func globalFun()  {
}
`
	fileBody2 := `package code
import "fmt"



func globalFun()  {


// test loc and  annotate
}
`
	f1Set := token.NewFileSet()
	f2Set := token.NewFileSet()
	f1, err := parser.ParseFile(f1Set, "file1.go", fileBody1, parser.ParseComments)
	require.NoError(t, err)
	f2, err := parser.ParseFile(f2Set, "file2.go", fileBody2, parser.ParseComments)
	require.NoError(t, err)
	require.True(t, DeepEqualForAst(f1.Decls[1], f2.Decls[1]))

}

func Test_DeepEqualForAstSameButCallFnChange(t *testing.T) {

	baseBody := `package code
 
func globalFun()  {
	changeFn()
}
`
	fileBody1 := baseBody + `func changeFn()  {
	 var a int
     _ = a
}
`
	fileBody2 := baseBody + `func changeFn()  {
	 }`

	f1Set := token.NewFileSet()
	f2Set := token.NewFileSet()
	f1, err := parser.ParseFile(f1Set, "file1.go", fileBody1, parser.ParseComments)
	require.NoError(t, err)
	f2, err := parser.ParseFile(f2Set, "file2.go", fileBody2, parser.ParseComments)
	require.NoError(t, err)

	require.True(t, DeepEqualForAst(f1.Decls[0].(*ast.FuncDecl).Body, f2.Decls[0].(*ast.FuncDecl).Body))
	require.False(t, DeepEqualForAst(f1.Decls[1].(*ast.FuncDecl).Body, f2.Decls[1].(*ast.FuncDecl).Body))
}

func Test_DeepEqualForAstDiffClosure(t *testing.T) {

	fileBody1 := `package code
func globalFun()  {
 var fn = func() {
}
fn()
}
`
	fileBody2 := `package code
 


func globalFun()  {

var fn = func() {
   var a int
   _ = a
}
fn()
// test loc and  annotate
}
`

	f1Set := token.NewFileSet()
	f2Set := token.NewFileSet()
	f1, err := parser.ParseFile(f1Set, "file1.go", fileBody1, parser.ParseComments)
	require.NoError(t, err)
	f2, err := parser.ParseFile(f2Set, "file2.go", fileBody2, parser.ParseComments)
	require.NoError(t, err)

	require.False(t, DeepEqualForAst(f1.Decls[0].(*ast.FuncDecl).Body, f2.Decls[0].(*ast.FuncDecl).Body))

}
