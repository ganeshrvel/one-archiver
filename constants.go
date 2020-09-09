package one_archiver

import "os"

var (
	GlobalPatternDenylist = []string{"pax_global_header", "__MACOSX/*", "*.DS_Store"}
	PathSep               = string(os.PathSeparator)
)

const (
	OverwriteExisting = true
)
