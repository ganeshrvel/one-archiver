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
	ProgressStreamDebounceTime = 1 * time.Second // 1 second
)

var allowedSecondExtensions allowedSecondExtMap = map[string]string{"tar": "tar"}
