package code_anaylsis

import (
	"go/types"
	"golang.org/x/tools/go/ssa"
)

/***************************
    @author: tiansheng.ren
    @date: 2024/12/19
    @desc:

***************************/

type Package struct {
	Name  string `json:"name"`
	Index uint32 `json:"index"`
}

type File struct {
	Index   uint32 `json:"index"`
	Package uint32 `json:"package"`
	Name    string `json:"name"`
}

type Struct struct {
	Index   uint32       `json:"index"`
	File    uint32       `json:"file"`
	Package uint32       `json:"package"`
	Name    string       `json:"name"`
	Metas   []string     `json:"metas"`
	Inst    *types.Named `json:"-"`
}

type Interface struct {
	Index   uint32           `json:"index"`
	File    uint32           `json:"file"`
	Package uint32           `json:"package"`
	Name    string           `json:"name"`
	Inst    *types.Interface `json:"-"`

	// 这里新加一个字段，用于 interface 绑定的 struct
}

type CallType uint8

const (
	CallTypeStatic CallType = iota + 1
	CallTypeInterface
)

type Call struct {
	Line   uint32   `json:"line"`
	Column uint32   `json:"column"`
	Type   CallType `json:"type"`
}

type CallInterfacePath struct {
	// name  idx 互斥，优先用name
	Name string
	Idx  int
}

type CallInterface struct {
	Call
	Interface uint32              `json:"interface"`
	Paths     []CallInterfacePath `json:"paths"`
}

type Function struct {
	Index         uint32          `json:"index"`
	File          uint32          `json:"file"`
	Package       uint32          `json:"package"`
	CallStatic    []Call          `json:"static"`
	CallInterface []CallInterface `json:"interface"`
	Name          string          `json:"name"`

	Inst *ssa.Function `json:"-"`
	// 标识是否已经解析过，无需再次解析
	Ready bool `json:"-"`
}

