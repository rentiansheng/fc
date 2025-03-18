package code_anaylsis

import (
	"context"
	"golang.org/x/tools/go/ssa"
	"strings"
)

/***************************
    @author: tiansheng.ren
    @date: 2024/12/19
    @desc: 解析函数的代码

***************************/

func (s *Scan) parserFuncBodyCode(ctx context.Context, fn *ssa.Function) ([]Call, []CallInterface) {
	c1, c2 := s.parserFuncBodyBlocks(ctx, fn.Blocks)
	// 如果有匿名函数处理
	for _, closuresFn := range fn.AnonFuncs {
		itemC1, itemC2 := s.parserFuncBodyCode(ctx, closuresFn)
		c1 = append(c1, itemC1...)
		c2 = append(c2, itemC2...)
	}
	return c1, c2
}

func (s *Scan) parserFuncBodyBlocks(ctx context.Context, blocks []*ssa.BasicBlock) ([]Call, []CallInterface) {
	var calls = make([]Call, 0)
	var callInterfaces = make([]CallInterface, 0)
	for _, block := range blocks {
		for _, instr := range block.Instrs {
			switch call := instr.(type) {
			case *ssa.Call:
				pos := s.prog.Fset.Position(call.Pos())

				// call func or struct func
				staticCallee := call.Call.StaticCallee()
				if staticCallee == nil { // if is call interface method, staticCallee is nil
					if call.Call.Method == nil { // skip builtin function, like append, make
						continue
					}

					//
					s.prog.Package(call.Call.Method.Pkg())
					if !s.isProjectByTypesPkg(ctx, call.Call.Method.Pkg()) {
						continue
					}
					callInfo := CallInterface{}
					callInfo.Line = uint32(pos.Line)
					callInfo.Column = uint32(pos.Column)
					callInfo.Type = CallTypeInterface
					// 不支持实例其他仓库的interface，调用链路追踪
					if s.InterfaceRela[call.Call.Method.FullName()] == nil {
						continue
					}

					callInfo.Interface = s.InterfaceRela[call.Call.Method.FullName()].Index
					callInfo.Paths = interfaceCallPaths(ctx, call.Call.Args[0])

				} else {
					//
					if !s.isProject(ctx, staticCallee.Pkg) {
						continue
					}
					fnName := getFuncFullName(staticCallee)
					fnInfo := s.FunctionRela[fnName]
					if fnInfo == nil {
						continue
					}

					pos := s.prog.Fset.Position(call.Pos())
					calls = append(calls, Call{
						Line:   uint32(pos.Line),
						Column: uint32(pos.Column),
						Type:   CallTypeStatic,
					})

				}

				// 处理参数中的函数调用, 这里暂未实现, 仅有框架
				s.parserFuncArgsCall(ctx, call.Call.Args, instr)

			}
		}
	}

	return calls, callInterfaces
}

func (s *Scan) parserFuncArgsCall(ctx context.Context, fnArgs []ssa.Value, instr ssa.Instruction) {
	for _, arg := range fnArgs {
		switch argTyp := arg.(type) {
		case *ssa.ChangeType:
			s.parserFuncArgsCall(ctx, []ssa.Value{argTyp.X}, instr)
		case *ssa.MakeClosure:
			s.parserFuncArgsCall(ctx, []ssa.Value{argTyp.Fn}, instr)
		case *ssa.Function:
			fnName := strings.Split(argTyp.String(), "$")[0]
			if fn, ok := s.StructRela[fnName]; ok {
				// TODO:  这里需要处理
				_ = fn
			}

		case *ssa.Slice:
			// TODO:  数组的没有实现
		}
	}
}
