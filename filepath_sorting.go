package onearchiver

import (
	"path/filepath"
	"sort"
)

func sortPath(list []ArchiveFileInfo, orderDir ArchiveOrderDir) []ArchiveFileInfo {
	var filePathList []filePathListSortInfo

	for _, x := range list {
		var splittedPaths [2]string

		if !x.IsDir {
			splittedPaths = [2]string{filepath.Dir(x.FullPath), x.Name}
		} else {
			splittedPaths = [2]string{filepath.Dir(x.FullPath), ""}
		}

		filePathList = append(filePathList, filePathListSortInfo{
			IsDir:         x.IsDir,
			FullPath:      x.FullPath,
			splittedPaths: splittedPaths,
			Mode:          x.Mode,
			Size:          x.Size,
			ModTime:       x.ModTime,
			Name:          x.Name,
			ParentPath:    x.ParentPath,
			Extension:     x.Extension,
			Kind:          x.Kind(),
		})
	}

	_sortPath(&filePathList, orderDir)

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

func _sortPath(pathList *[]filePathListSortInfo, orderDir ArchiveOrderDir) {
	_pathList := *pathList

	sort.SliceStable(_pathList, func(i, j int) bool {
		if orderDir == OrderDirDesc {
			return _pathList[i].splittedPaths[0] > _pathList[j].splittedPaths[0]
		}

		return _pathList[i].splittedPaths[0] < _pathList[j].splittedPaths[0]
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

			if _pathList[currentIndex].splittedPaths[0] != _pathList[nextIndex].splittedPaths[0] {

				end = nextIndex
				count = currentIndex

				break
			}

		}

		trimmedPathList := _pathList[start:end]

		sort.SliceStable(trimmedPathList, func(i, j int) bool {
			if orderDir == OrderDirDesc {
				return trimmedPathList[i].splittedPaths[1] > trimmedPathList[j].splittedPaths[1]
			}

			return trimmedPathList[i].splittedPaths[1] < trimmedPathList[j].splittedPaths[1]
		})

		count += 1
	}
}
