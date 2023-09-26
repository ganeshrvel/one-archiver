package onearchiver

// archiveProgressMetrices is used for tracking the progress of file archiving.
// The type parameter T allows for customization of how archive paths are represented.
type archiveProgressMetrices[T any] struct {
	totalFiles, totalSize, filesProgressCount int64
}

func newArchiveProgressMetrices[T any]() *archiveProgressMetrices[T] {
	return &archiveProgressMetrices[T]{
		totalFiles:         0,
		totalSize:          0,
		filesProgressCount: 0,
	}
}

// updateArchiveProgressMetrices updates the counters based on the path, size, and whether it's a directory.
func (metrics *archiveProgressMetrices[T]) updateArchiveProgressMetrices(m map[string]T, path string, size int64, isDir bool) {
	// - If the path is already present in the map or is a directory, it returns without updating the metrics.
	_, exists := m[path]
	if exists || isDir {
		return
	}

	// - Otherwise, it increments the totalSize by the given size and the totalFiles by 1.
	metrics.totalSize += size
	metrics.totalFiles += 1
}

// updateArchiveFilesProgressCount updates the file progress count based on whether the input represents a directory or not.
// If the input is a directory, the function returns without any update.
// Otherwise, it increments the filesProgressCount by 1.
func (metrics *archiveProgressMetrices[T]) updateArchiveFilesProgressCount(isDir bool) {
	if isDir {
		return
	}

	metrics.filesProgressCount += 1
}
