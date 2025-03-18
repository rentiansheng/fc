package code_anaylsis

import (
	"context"
	"fmt"
	"go/types"
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"

	"regexp"
	"strings"
)

/***************************
    @author: tiansheng.ren
    @date: 2024/12/19
    @desc:

***************************/

type Scan struct {
	Namespace string `json:"namespace"`
	Dir       string `json:"dir"`

	cfg *packages.Config `json:"-"`

	prog    *ssa.Program
	ssaPkgs []*ssa.Package
	pkgs    []*packages.Package

	ignoreRegexp []*regexp.Regexp

	Packages   []Package   `json:"packages"`
	Files      []File      `json:"files"`
	Structs    []Struct    `json:"structs"`
	Interfaces []Interface `json:"interfaces"`
	Functions  []Function  `json:"functions"`
	scanMemCache
}

// scanMemCache is a memory cache for scan
type scanMemCache struct {
	PackageRela   map[string]*Package   `json:"-"`
	FileRela      map[string]*File      `json:"-"`
	StructRela    map[string]*Struct    `json:"-"`
	InterfaceRela map[string]*Interface `json:"-"`
	FunctionRela  map[string]*Function  `json:"-"`

	ASTPackageRela map[string]*packages.Package `json:"-"`
}

func newDefaultCfg() *packages.Config {
	return &packages.Config{

		Mode: packages.NeedFiles |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedName |
			packages.NeedDeps |
			packages.NeedTypesInfo |
			packages.NeedImports,
	}
}

func NewScan(dir string, namespace string, cfg *packages.Config) *Scan {
	if cfg == nil {
		cfg = newDefaultCfg()
	}
	cfg.Dir = dir
	mc := scanMemCache{
		PackageRela:   make(map[string]*Package),
		FileRela:      make(map[string]*File),
		StructRela:    make(map[string]*Struct),
		InterfaceRela: make(map[string]*Interface),
		FunctionRela:  make(map[string]*Function),

		ASTPackageRela: make(map[string]*packages.Package),
	}
	return &Scan{
		Namespace:    namespace,
		Dir:          dir,
		cfg:          cfg,
		Packages:     make([]Package, 0),
		Files:        make([]File, 0),
		Structs:      make([]Struct, 0),
		Interfaces:   make([]Interface, 0),
		Functions:    make([]Function, 0),
		scanMemCache: mc,
	}
}

func (s *Scan) isProject(ctx context.Context, pkg *ssa.Package) bool {
	if pkg == nil {
		return false
	}
	return s.isProjectByTypesPkg(ctx, pkg.Pkg)
}

func (s *Scan) isProjectByTypesPkg(ctx context.Context, pkg *types.Package) bool {
	if pkg == nil {
		return false
	}
	return strings.HasPrefix(pkg.Path(), s.Namespace)
}

func (s *Scan) Analysis(ctx context.Context) error {
	initial, err := packages.Load(s.cfg, s.Dir+"/...")
	if err != nil {
		return fmt.Errorf("load packages error: %w", err)
	}

	prog, pkgs := ssautil.Packages(initial, ssa.NaiveForm)
	for _, pkg := range pkgs {
		if pkg == nil {
			continue
		}
		if !s.isProject(ctx, pkg) {
			continue
		}
		pkg.Build()

	}
	prog.Build()
	s.prog, s.ssaPkgs = prog, pkgs
	s.pkgs = initial

	s.initCache(ctx)
	// notice:  analysis all functions, later will be optimized, only analysis use functions
	for ssaFn := range ssautil.AllFunctions(s.prog) {
		if !s.isProject(ctx, ssaFn.Pkg) {
			continue
		}
		fn := s.FunctionRela[ssaFn.String()]
		if fn == nil {
			continue
		}
		if fn.Ready {
			continue
		}
		call, interfaceCall := s.parserFuncBodyCode(ctx, ssaFn)
		fn.CallStatic = call
		fn.CallInterface = interfaceCall
		fn.Ready = true
	}

	return nil
}

func (s *Scan) GoPackages() []*packages.Package {
	return s.pkgs
}

func (s *Scan) GoSSAPackages() []*ssa.Package {
	return s.ssaPkgs
}

func (s *Scan) Func(name string) *Function {
	return s.FunctionRela[name]

}

func inStd(node *callgraph.Node) bool {
	return isStdPkgPath(node.Func.Pkg.Pkg.Path())
}

func isStdPkgPath(path string) bool {
	if strings.Contains(path, ".") {
		return false
	}
	return true
}
