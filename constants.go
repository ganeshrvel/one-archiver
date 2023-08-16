package onearchiver

import (
	"os"
	"time"
)

var (
	GlobalPatternDenylist = []string{"pax_global_header", "__MACOSX/*", "*.DS_Store"}
	PathSep               = string(os.PathSeparator)
)

const (
	OverwriteExisting          = true
	ProgressStreamDebounceTime = 500 * time.Millisecond // 500 ms
)

var allowedSecondExtensions allowedSecondExtMap = map[string]string{"tar": "tar"}
