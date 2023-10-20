package onearchiver

import (
	"os"
)

var (
	GlobalPatternDenylist = []string{"pax_global_header", "__MACOSX/*", "*.DS_Store"}
	PathSep               = string(os.PathSeparator)
)

const (
	OverwriteExisting = true
)

var allowedSecondExtensions allowedSecondExtMap = map[string]string{"tar": "tar"}
