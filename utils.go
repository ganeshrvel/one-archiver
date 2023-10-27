package onearchiver

import (
	"fmt"
	"github.com/ganeshrvel/rardecode"
	"github.com/mitchellh/go-homedir"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)

	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)

	return !os.IsNotExist(err)
}

func GetDesktopFile(filename string) string {
	_home, _ := homedir.Dir()

	return filepath.Join(_home, "Desktop", filename)
}

func GetHomeDirFile(filename string) string {
	_home, _ := homedir.Dir()

	return filepath.Join(_home, filename)
}

// GetCommonParentPath - Get Parent path of a list of directories and files
func GetCommonParentPath(sep byte, paths ...string) string {
	// Handle special cases.
	switch len(paths) {
	case 0:
		return ""
	case 1:
		return path.Clean(paths[0])
	}

	// Note, we treat string as []byte, not []rune as is often
	// done in Go. (And sep as byte, not rune). This is because
	// most/all supported OS' treat paths as string of non-zero
	// bytes. A filename may be displayed as a sequence of Unicode
	// runes (typically encoded as UTF-8) but paths are
	// not required to be valid UTF-8 or in any normalized form
	// (e.g. "é" (U+00C9) and "é" (U+0065,U+0301) are different
	// file names.
	c := []byte(path.Clean(paths[0]))

	// We add a trailing sep to handle the case where the
	// common prefix directory is included in the path list
	// (e.g. /home/user1, /home/user1/foo, /home/user1/bar).
	// path.Clean will have cleaned off trailing / separators with
	// the exception of the root directory, "/" (in which case we
	// make it "//", but this will get fixed up to "/" bellow).
	c = append(c, sep)

	// Ignore the first path since it's already in c
	for _, v := range paths[1:] {
		// Clean up each path before testing it
		v = path.Clean(v) + string(sep)

		// Find the first non-common byte and truncate c
		if len(v) < len(c) {
			c = c[:len(v)]
		}
		for i := 0; i < len(c); i++ {
			if v[i] != c[i] {
				c = c[:i]
				break
			}
		}
	}

	// Remove trailing non-separator characters and the final separator
	for i := len(c) - 1; i >= 0; i-- {
		if c[i] == sep {
			c = c[:i]
			break
		}
	}

	return string(c)
}

func isFile(name string) bool {
	if fi, err := os.Stat(name); err == nil {
		if fi.Mode().IsRegular() {
			return true
		}
	}

	return false
}

func IsDirectory(name string) bool {
	if fi, err := os.Stat(name); err == nil {
		if fi.Mode().IsDir() {
			return true
		}
	}
	return false
}

func FixDirSlash(isDir bool, absFilepath string) string {
	if isDir && !strings.HasSuffix(absFilepath, PathSep) {
		absFilepath = fmt.Sprintf("%s%s", absFilepath, PathSep)
	}

	return absFilepath
}

func indexExists(arr interface{}, index int) bool {
	switch value := arr.(type) {
	case *[]string:
		return len(*value) > index

	case []string:
		return len(value) > index

	default:
		log.Panic("invalid type in 'indexExists'")
	}

	return false
}

func IsSymlink(fi os.FileInfo) bool {
	return fi.Mode()&os.ModeSymlink != 0
}

func IsRarSymlink(h *rardecode.FileHeader) bool {
	return h.Attributes == 1056
}

func Percent(partial, total float64) float64 {
	if total == 0 {
		return 0
	}

	return (partial / total) * 100
}

// TransferRatePercent calculates the transfer rate percentage.
// If both the total and partial file sizes are 0, it returns 100% as completed.
func TransferRatePercent(partial, total float64) float64 {
	if total == 0 && partial == 0 {
		return 100
	}

	return Percent(partial, total)
}

func StringFilter(x []string, f func(string) bool) []string {
	a := make([]string, 0)

	for _, v := range x {
		if f(v) && len(v) > 7 {
			a = append(a, v)
		}
	}

	return a
}

func subpathExists(path string, searchPath string) bool {
	return path != "" && strings.HasPrefix(searchPath, path)
}

func GetParentDirectory(fullPath string) string {
	// return if [fullPath] = [PathSep]
	if fullPath == PathSep {
		return fullPath
	}

	// return if [fullPath] = ""
	if fullPath == "" {
		return ""
	}

	// return if [fullPath] = .
	if fullPath == "." {
		return ""
	}

	// append '/' to the fullPath just so that parent directory is parsed correctly
	fullPath = fmt.Sprintf("%s/", fullPath)
	pd := filepath.Dir(fullPath)

	pdSplit, _ := filepath.Split(pd)

	return pdSplit
}

func Extension(filename string) string {
	_, _filename := filepath.Split(filename)

	f := strings.Split(_filename, ".")
	var extension string

	length := len(f)

	if length < 1 {
		return extension
	}

	if length > 2 {
		exts := f[length-2:]
		if _, ok := allowedSecondExtensions[exts[0]]; ok {
			return strings.Join(exts, ".")
		}
	}

	if length > 1 {
		exts := f[length-1:]

		return exts[0]
	}

	return extension
}

func sanitizeTime(objTime time.Time, arcTime time.Time) time.Time {
	zeroTime := time.Time{}
	if objTime == zeroTime {
		return arcTime
	}

	return objTime
}
