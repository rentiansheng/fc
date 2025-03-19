package code

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"
)

/***************************
    @author: tiansheng.ren
    @date: 2025/3/12
    @desc:

***************************/

func Test_Diff(t *testing.T) {
	fileBody1 := `package code
type  methodFunc struct {}
func (m *methodFunc) same1()   { }
func (m methodFunc) same2()  {var a int; _ = a}
func (m methodFunc) diff1()  {}
func (m methodFunc) diff2()  {}
func  same1()  {}
func  same2()  {}
func diff1()  {}
func diff2()  {}
`

	fileBody2 := `package code
type  methodFunc struct {}
func (m methodFunc) same2()  {var a int; _ = a}
func (m *methodFunc) same1()  {}
func (m methodFunc) diff2()  { {var a int; _ = a}}
func (m methodFunc) diff1()  { {var a int; _ = a}}
func  same1()  {}
func  same2()  {}
func diff2()  { {var a int; _ = a}}
func diff1()  { {var a int; _ = a}}`

	f1Set := token.NewFileSet()
	f2Set := token.NewFileSet()
	f1, err := parser.ParseFile(f1Set, "file1.go", fileBody1, parser.ParseComments)
	require.NoError(t, err)
	f2, err := parser.ParseFile(f2Set, "file2.go", fileBody2, parser.ParseComments)
	require.NoError(t, err)

	cd := NewCodeDiff(f1, f2, f1Set, f2Set)
	sames, diffs, err := cd.Diff()
	require.NoError(t, err)
	require.Equal(t, 4, len(sames))
	require.Equal(t, 4, len(diffs))

}

func Test_DiffStruct(t *testing.T) {
	fileBody1 := `package code
type  methodFunc struct {}
func (m methodFunc) same1()  {var a int; _ = a}
func  diff1()  {a := struct{}{}; _ = a}
`

	fileBody2 := `package code
type  methodFunc struct {a int}
func (m methodFunc) same1()  {var a int; _ = a}
func  diff1()  {a := struct{a int}{}; _ = a}

`

	f1Set := token.NewFileSet()
	f2Set := token.NewFileSet()
	f1, err := parser.ParseFile(f1Set, "file1.go", fileBody1, parser.ParseComments)
	require.NoError(t, err)
	f2, err := parser.ParseFile(f2Set, "file2.go", fileBody2, parser.ParseComments)
	require.NoError(t, err)

	cd := NewCodeDiff(f1, f2, f1Set, f2Set)
	_, diffs, err := cd.Diff()
	require.NoError(t, err)
	require.Equal(t, 1, len(diffs))

}
