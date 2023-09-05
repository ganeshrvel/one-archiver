package onearchiver

// TODO proper error handling

import (
	"fmt"
	"sort"
	"strings"
)

func sortFiles(list []ArchiveFileInfo, orderBy ArchiveOrderBy, orderDir ArchiveOrderDir) []ArchiveFileInfo {
	if orderDir == OrderDirNone {
		return list
	}

	switch orderBy {
	case OrderByFullPath:
		return sortPath(list, orderDir)

	case OrderByName:
		sort.SliceStable(list, func(i, j int) bool {
			if orderDir == OrderDirDesc {
				return list[i].Name > list[j].Name
			}

			return list[i].Name < list[j].Name
		})
	case OrderByModTime:
		sort.SliceStable(list, func(i, j int) bool {
			if orderDir == OrderDirDesc {
				return list[i].ModTime.After(list[j].ModTime)
			}

			return list[i].ModTime.Before(list[j].ModTime)
		})
	case OrderBySize:
		sort.SliceStable(list, func(i, j int) bool {
			if orderDir == OrderDirDesc {
				return list[i].Size > list[j].Size
			}

			return list[i].Size < list[j].Size
		})
	case OrderByKind:
		sort.SliceStable(list, func(i, j int) bool {
			if orderDir == OrderDirDesc {
				return list[i].Kind() > list[j].Kind()
			}

			return list[i].Kind() < list[j].Kind()
		})
	}

	return list
}

func getFilteredFiles(fileInfo ArchiveFileInfo, listDirectoryPath string, recursive bool) (include bool) {
	isInPath := strings.HasPrefix(fileInfo.FullPath, listDirectoryPath)

	if isInPath {
		// dont return the directory path if it's listDirectoryPath. This will make sure that only files and sub directories are returned
		if listDirectoryPath == fileInfo.FullPath {
			return false
		}

		// if recursive mode is true return all files and subdirectories under the filtered path
		if recursive {
			return true
		}

		slashSplitListDirectoryPath := strings.Split(listDirectoryPath, PathSep)
		slashSplitListDirectoryPathLength := len(slashSplitListDirectoryPath)

		slashSplitFullPath := strings.Split(fileInfo.FullPath, PathSep)
		slashSplitFullPathLength := len(slashSplitFullPath)

		// if directory allow an extra '/' to figure out the subdirectory
		if fileInfo.IsDir && slashSplitFullPathLength < slashSplitListDirectoryPathLength+2 {
			return true
		}

		if !fileInfo.IsDir && slashSplitFullPathLength < slashSplitListDirectoryPathLength+1 {
			return true
		}
	}

	return false
}

func GetArchiveFileList(meta *ArchiveMeta, read *ArchiveRead) ([]ArchiveFileInfo, error) {
	_meta := *meta
	_read := *read

	pctx := _read.passwordContext()

	var arcObj ArchiveReader

	// check whether the archive is encrypted
	// if yes, check whether the password is valid
	prep, err := PrepareArchive(meta, _read.Passwords)
	if err != nil {
		return nil, err
	}

	/// if an archive requires password(s) to read it and if password field is empty
	/// then return 'password is required' error
	if prep.IsPasswordRequired && !pctx.hasPasswords() {
		return nil, fmt.Errorf(string(ErrorPasswordRequired))
	}

	/// if an archive requires password(s) to read it and if the password is invalid
	/// then return 'invalid password' error
	if prep.IsPasswordRequired && !prep.IsValidPassword {
		return nil, fmt.Errorf(string(ErrorInvalidPassword))
	}

	ext := extension(meta.Filename)

	// add a trailing slash to [listDirectoryPath] if missing
	if _read.ListDirectoryPath != "" && !strings.HasSuffix(_read.ListDirectoryPath, PathSep) {
		_read.ListDirectoryPath = fmt.Sprintf("%s%s", _read.ListDirectoryPath, PathSep)
	}

	switch ext {
	case "zip":
		arcObj = zipArchive{meta: _meta, read: _read}
	case "zst":
		fallthrough
	case "xz":
		fallthrough
	case "sz":
		fallthrough
	case "lz4":
		fallthrough
	case "bz2":
		fallthrough
	case "br":
		fallthrough
	case "gz":
		arcObj = compressedFile{meta: _meta, read: _read}
	case "tar.zst":
		fallthrough
	case "tar.xz":
		fallthrough
	case "tar.sz":
		fallthrough
	case "tar.lz4":
		fallthrough
	case "tar.bz2":
		fallthrough
	case "tar.br":
		fallthrough
	case "tar.gz":
		fallthrough
	default:
		arcObj = commonArchive{meta: _meta, read: _read}
	}

	return arcObj.list()
}
