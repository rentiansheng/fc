package wire

import (
	"github.com/rentiansheng/fc/code_anaylsis"
	"github.com/rentiansheng/fc/injection"
	"strings"
)

/***************************
    @author: tiansheng.ren
    @date: 2024/12/30
    @desc:

***************************/

const fileSuffix = "/wire_gen.go"
const vendorPart = "/vendor/"

type wire struct {
	namespace string
	dir       string
	s         *code_anaylsis.Scan

	// map[scan.SSAPackages index] -> []*injection.Inject
	inject map[uint32][]*injection.Inject

	// map[func index] -> []*injection.Inject
	wireFnCallInject map[uint32][]*injection.Inject

	// map[file name] -> [package name] -> package path
	fileImportCache map[string]map[string]string
}

func (w *wire) ParseGenFile() {
	for _, pkg := range w.s.GoPackages() {
		if pkg == nil {
			continue
		}

		for _, filename := range pkg.GoFiles {
			if strings.Contains(filename, vendorPart) {
				break
			}

			if !strings.HasPrefix(filename, fileSuffix) {
				continue
			}
		}

	}
}
