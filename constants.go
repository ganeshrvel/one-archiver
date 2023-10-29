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
	OverwriteExisting                       = true
	DefaultProgressStreamDebounceTime int64 = int64(500 * time.Millisecond)
)

var allowedSecondExtensions allowedSecondExtMap = map[string]string{"tar": "tar"}
