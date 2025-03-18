package code_anaylsis

import (
	"context"
	"go/types"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
	"path"
	"strings"
)

/***************************
    @author: tiansheng.ren
    @date: 2024/12/19
    @desc: 处理代码基础类型，包含pkg,  文件，结构体，接口，函数

***************************/

var ignoreFile = []string{"vendor"}

func (s *Scan) cacheSSAPkg(ctx context.Context, name string) uint32 {
	if _, ok := s.PackageRela[name]; ok {
		return s.PackageRela[name].Index
	}
	pkgIdx := len(s.Packages)
	s.Packages = append(s.Packages, Package{
		Index: uint32(pkgIdx),
		Name:  name,
	})
	s.PackageRela[name] = &s.Packages[pkgIdx]
	return uint32(pkgIdx)
}

func (s *Scan) cacheFile(ctx context.Context, pkgIdx uint32, name string) uint32 {
	if _, ok := s.FileRela[name]; ok {
		return s.FileRela[name].Index
	}
	fileIdx := len(s.Files)
	s.Files = append(s.Files, File{
		Index:   uint32(fileIdx),
		Package: pkgIdx,
		Name:    name,
	})
	s.FileRela[name] = &s.Files[fileIdx]
	return uint32(fileIdx)
}

func (s *Scan) canCacheFile(ctx context.Context, filename string) bool {
	// 非当前项目忽律
	if !strings.HasPrefix(filename, s.Dir) {
		return false
	}
	// 忽律 vendor 文件
	for _, ig := range ignoreFile {
		if strings.HasPrefix(filename, path.Join(s.Dir, ig)) {
			return false
		}
	}
	for _, ig := range s.ignoreRegexp {
		if ig.MatchString(filename) {
			return false
		}

	}
	return true
}

func (s *Scan) initCache(ctx context.Context) {

	for _, pkg := range s.ssaPkgs {
		if !s.isProject(ctx, pkg) {
			continue
		}
		s.cacheSSAPkg(ctx, pkg.Pkg.Path())
	}

	for _, pkg := range s.pkgs {
		if !s.isProjectByTypesPkg(ctx, pkg.Types) {
			continue
		}
		s.ASTPackageRela[pkg.PkgPath] = pkg
		for _, f := range pkg.GoFiles {
			if !strings.HasPrefix(f, s.Dir) {
				continue
			}

			pkgInternalPath := path.Dir(strings.TrimPrefix(f, s.Dir))
			pkgName := path.Join(s.Namespace, pkgInternalPath)
			if _, ok := s.PackageRela[pkgName]; !ok {
				continue
			}
			pkgIdx := s.PackageRela[pkgName].Index
			s.cacheFile(ctx, pkgIdx, f)
		}

	}

	for _, pkg := range s.ssaPkgs {

		if !s.isProject(ctx, pkg) {
			continue
		}
		pkgIdx := s.cacheSSAPkg(ctx, pkg.Pkg.Path())
		for key := range pkg.Members {

			if t := pkg.Type(key); t != nil {
				filePos := pkg.Prog.Fset.Position(t.Pos())
				if !strings.HasPrefix(filePos.Filename, s.Dir) {
					continue
				}
				file := s.FileRela[filePos.Filename]
				if file == nil {
					println("file not found", filePos.Filename)
					continue
				}
				// struct or interface
				if nd, ok := t.Type().(*types.Named); ok {
					s.memberCache(ctx, pkgIdx, file.Index, nd)
				}
			}
		}

	}

	for fn := range ssautil.AllFunctions(s.prog) {
		if !s.isProject(ctx, fn.Pkg) {
			continue
		}
		s.cacheFunc(ctx, fn)
		// 这里需要先缓存函数，后续才能解析函数调用，为了提前占位置
		s.cacheSSAPkg(ctx, fn.Pkg.Pkg.Path())
	}

}

func (s *Scan) cacheFunc(ctx context.Context, fn *ssa.Function) {
	if !s.isProject(ctx, fn.Pkg) {
		return
	}
	pkgIdx := s.cacheSSAPkg(ctx, fn.Pkg.Pkg.Path())
	fileName := s.prog.Fset.Position(fn.Pos()).Filename
	fileIdx := s.cacheFile(ctx, pkgIdx, fileName)
	funcIdx := len(s.Functions)
	s.Functions = append(s.Functions, Function{
		Index:   uint32(funcIdx),
		Package: pkgIdx,
		File:    fileIdx,
		Name:    fn.String(),
		Inst:    fn,
	})
	s.FunctionRela[fn.String()] = &s.Functions[funcIdx]
}

func (s *Scan) memberCache(ctx context.Context, pkgIdx, fileIdx uint32, nd *types.Named) {

	if types.IsInterface(nd) {
		// interface
		interfaceIdx := len(s.Interfaces)
		it := nd.Underlying().(*types.Interface)
		injects := make(map[string]map[uint64]uint32)
		for i := 0; i < it.NumMethods(); i++ { // interface node need add now, case no func body
			injects[it.Method(i).Name()] = make(map[uint64]uint32)
		}
		s.Interfaces = append(s.Interfaces, Interface{
			Index:   uint32(interfaceIdx),
			Package: pkgIdx,
			File:    fileIdx,
			Name:    it.String(),
			Inst:    it,
		})
	} else {
		if !s.isProjectByTypesPkg(ctx, nd.Obj().Pkg()) {
			return
		}
		metas := make([]string, 0, nd.NumMethods())
		for idx := 0; idx < nd.NumMethods(); idx++ {
			metas = append(metas, nd.Method(idx).Name())
		}
		// struct
		structIdx := len(s.Structs)
		s.Structs = append(s.Structs, Struct{
			Index:   uint32(structIdx),
			Package: pkgIdx,
			File:    fileIdx,
			Name:    nd.String(),
			Metas:   metas,
			Inst:    nd,
		})
	}
	return
}
