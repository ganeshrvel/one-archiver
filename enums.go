package onearchiver

type ArchiveOrderBy string

const (
	OrderBySize     ArchiveOrderBy = "size"
	OrderByModTime  ArchiveOrderBy = "modTime"
	OrderByName     ArchiveOrderBy = "name"
	OrderByFullPath ArchiveOrderBy = "fullPath"
)

type ArchiveOrderDir string

const (
	OrderDirAsc  ArchiveOrderDir = "asc"
	OrderDirDesc ArchiveOrderDir = "desc"
	OrderDirNone ArchiveOrderDir = "none"
)
