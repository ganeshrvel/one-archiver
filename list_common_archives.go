package one_archiver

import (
	"archive/tar"
	"fmt"
	"github.com/ganeshrvel/archiver"
	"github.com/nwaples/rardecode"
	"path/filepath"
)

// List files in the common archives
func (arc commonArchive) list() ([]ArchiveFileInfo, error) {
	_filename := arc.meta.Filename
	_password := arc.meta.Password
	_listDirectoryPath := arc.read.ListDirectoryPath
	_recursive := arc.read.Recursive
	_orderBy := arc.read.OrderBy
	_orderDir := arc.read.OrderDir
	_gitIgnorePattern := arc.meta.GitIgnorePattern

	arcFileObj, err := archiver.ByExtension(_filename)

	if err != nil {
		return nil, err
	}

	err = archiveFormat(&arcFileObj, _password, OverwriteExisting)

	if err != nil {
		return nil, err
	}

	var arcWalker, ok = arcFileObj.(archiver.Walker)
	if !ok {
		return nil, fmt.Errorf("some error occured while reading the archive")
	}

	var filteredPaths []ArchiveFileInfo

	isListDirectoryPathExist := _listDirectoryPath == ""

	err = arcWalker.Walk(_filename, func(file archiver.File) error {
		var fileInfo ArchiveFileInfo

		switch fileHeader := file.Header.(type) {
		case *tar.Header:
			fileInfo = ArchiveFileInfo{
				Mode:     file.Mode(),
				Size:     file.Size(),
				IsDir:    file.IsDir(),
				ModTime:  file.ModTime(),
				Name:     file.Name(),
				FullPath: filepath.ToSlash(fileHeader.Name),
			}

			break

		case *rardecode.FileHeader:
			fileInfo = ArchiveFileInfo{
				Mode:     file.Mode(),
				Size:     file.Size(),
				IsDir:    file.IsDir(),
				ModTime:  file.ModTime(),
				Name:     file.Name(),
				FullPath: filepath.ToSlash(fileHeader.Name),
			}

			break

		// not being used
		default:
			fileInfo = ArchiveFileInfo{
				Mode:     file.Mode(),
				Size:     file.Size(),
				IsDir:    file.IsDir(),
				ModTime:  file.ModTime(),
				Name:     file.Name(),
				FullPath: filepath.ToSlash(file.FileInfo.Name()),
			}

			break
		}

		fileInfo.FullPath = fixDirSlash(fileInfo.IsDir, fileInfo.FullPath)

		allowIncludeFile := getFilteredFiles(fileInfo, _listDirectoryPath, _recursive, _gitIgnorePattern)

		if allowIncludeFile {
			filteredPaths = append(filteredPaths, fileInfo)
		}

		if !isListDirectoryPathExist && subpathExists(_listDirectoryPath, fileInfo.FullPath) {
			isListDirectoryPathExist = true
		}

		return nil
	})

	if !isListDirectoryPathExist {
		return filteredPaths, fmt.Errorf("path not found to filter: %s", _listDirectoryPath)
	}

	if arc.read.OrderDir == OrderDirNone {
		return filteredPaths, err
	}

	sortedPaths := sortFiles(filteredPaths, _orderBy, _orderDir)

	return sortedPaths, err
}
