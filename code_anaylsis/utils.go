package code_anaylsis

import (
	"context"
	"github.com/pkg/errors"
	"go/types"
	"golang.org/x/tools/go/ssa"
)

/***************************
    @author: tiansheng.ren
    @date: 2024/12/19
    @desc:

***************************/

func getFuncFullName(f *ssa.Function) string {
	if f.Object() == nil {
		return f.String()
	}

	if fn := f.Object().(*types.Func); fn != nil {
		return fn.FullName()
	}

	return ""
}

func interfaceCallPaths(ctx context.Context, f ssa.Value) []CallInterfacePath {

	switch fxt := f.(type) {
	case *ssa.UnOp:
		return interfaceCallPaths(ctx, fxt.X)
	case *ssa.Alloc:
		return []CallInterfacePath{{Name: fxt.Comment}}
	case *ssa.FieldAddr:
		return append(interfaceCallPaths(ctx, fxt.X), CallInterfacePath{Idx: fxt.Field})
	default:
		println("call fn field addr path not found")
	}

	return nil
}
