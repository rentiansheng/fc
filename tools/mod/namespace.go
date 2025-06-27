package mod

import (
	"bytes"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"os/fileapth"
)

/***************************
    @author: tiansheng.ren
    @date: 2025/3/19
    @desc:

***************************/

// MainModulePath returns the main module path from the gomod file text of go repo.
func MainModulePath(repoPath string) (mainModulePath string, err error) {
	// Check go.mod exist
	goModPath := filepath.Join(repoPath, "go.mod")
	if _, err = os.Stat(goModPath); os.IsNotExist(err) {
		return
	}

	// Read go.mod
	src, err := ioutil.ReadFile(goModPath)
	if err != nil {
		return
	}
	mainModulePath = modulePath(src)
	if mainModulePath == "" {
		err = errors.New("go.mod missing module path")
	}
	return
}

func modulePath(mod []byte) string {
	for len(mod) > 0 {
		line := mod
		mod = nil
		if i := bytes.IndexByte(line, '\n'); i >= 0 {
			line, mod = line[:i], line[i+1:]
		}
		if i := bytes.Index(line, slashSlash); i >= 0 {
			line = line[:i]
		}
		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, moduleStr) {
			continue
		}
		line = line[len(moduleStr):]
		n := len(line)
		line = bytes.TrimSpace(line)
		if len(line) == n || len(line) == 0 {
			continue
		}

		if line[0] == '"' || line[0] == '`' {
			p, err := strconv.Unquote(string(line))
			if err != nil {
				return "" // malformed quoted string or multiline module path
			}
			return p
		}

		return string(line)
	}
	return "" // missing module path
}
