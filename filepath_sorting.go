package onearchiver

import (
	"path/filepath"
	"sort"
)

// ConvertAndSortByPath transforms a list of ArchiveFileInfo into a sortable format,
// sorts it based on the directory and file structure, and then converts it back
// to a list of ArchiveFileInfo. Sorting can be in ascending or descending order
// based on the orderDir parameter.
func ConvertAndSortByPath(list []ArchiveFileInfo, orderDir ArchiveOrderDir) []ArchiveFileInfo {
	var filePathList []FilePathListSortInfo

	for _, x := range list {
		var splittedPaths [2]string

		if !x.IsDir {
			splittedPaths = [2]string{filepath.Dir(x.FullPath), x.Name}
		} else {
			splittedPaths = [2]string{filepath.Dir(x.FullPath), ""}
		}

		filePathList = append(filePathList, FilePathListSortInfo{
			IsDir:         x.IsDir,
			FullPath:      x.FullPath,
			SplittedPaths: splittedPaths,
			Mode:          x.Mode,
			Size:          x.Size,
			ModTime:       x.ModTime,
			Name:          x.Name,
			ParentPath:    x.ParentPath,
			Extension:     x.Extension,
			Kind:          x.Kind(),
		})
	}

	SortBySplittedPath(&filePathList, orderDir)

	var resultList []ArchiveFileInfo

	for _, x := range filePathList {
		resultList = append(resultList, ArchiveFileInfo{
			Mode:       x.Mode,
			Size:       x.Size,
			IsDir:      x.IsDir,
			ModTime:    x.ModTime,
			Name:       x.Name,
			FullPath:   x.FullPath,
			ParentPath: x.ParentPath,
			Extension:  x.Extension,
		})
	}

	return resultList
}

// SortBySplittedPath sorts the list of FilePathListSortInfo based on its split path structure.
// It first sorts by directory names and then by file names within those directories.
func SortBySplittedPath(pathList *[]FilePathListSortInfo, orderDir ArchiveOrderDir) {
	_pathList := *pathList

	sort.SliceStable(_pathList, func(i, j int) bool {
		if orderDir == OrderDirDesc {
			return _pathList[i].SplittedPaths[0] > _pathList[j].SplittedPaths[0]
		}

		return _pathList[i].SplittedPaths[0] < _pathList[j].SplittedPaths[0]
	})

	count := 0
	for count < len(_pathList)-1 {
		start := count
		end := len(_pathList)

		for k := range _pathList[count:] {
			currentIndex := count + k
			nextIndex := count + k + 1

			if nextIndex >= len(_pathList) {
				break
			}

			if _pathList[currentIndex].SplittedPaths[0] != _pathList[nextIndex].SplittedPaths[0] {

				end = nextIndex
				count = currentIndex

				break
			}

		}

		trimmedPathList := _pathList[start:end]

		sort.SliceStable(trimmedPathList, func(i, j int) bool {
			if orderDir == OrderDirDesc {
				return trimmedPathList[i].SplittedPaths[1] > trimmedPathList[j].SplittedPaths[1]
			}

			return trimmedPathList[i].SplittedPaths[1] < trimmedPathList[j].SplittedPaths[1]
		})

		count += 1
	}
}
