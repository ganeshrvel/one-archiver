package onearchiver

// TODO proper error handling

import (
	"fmt"
	ignore "github.com/sabhiram/go-gitignore"
	"github.com/yeka/zip"
	"path/filepath"
)

// list files in zip archives
// yeka package is used here to list encrypted zip files
func (arc zipArchive) list() ([]ArchiveFileInfo, error) {
	_filename := arc.meta.Filename
	_listDirectoryPath := arc.read.ListDirectoryPath
	_password := arc.meta.Password
	_recursive := arc.read.Recursive
	_orderBy := arc.read.OrderBy
	_orderDir := arc.read.OrderDir
	_gitIgnorePattern := arc.meta.GitIgnorePattern

	reader, err := zip.OpenReader(_filename)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = reader.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	var filteredPaths []ArchiveFileInfo

	isListDirectoryPathExist := _listDirectoryPath == ""

	var ignoreList []string
	ignoreList = append(ignoreList, GlobalPatternDenylist...)
	ignoreList = append(ignoreList, _gitIgnorePattern...)
	compiledGitIgnoreLines, _ := ignore.CompileIgnoreLines(ignoreList...)

	for _, file := range reader.File {
		if _password != "" {
			file.SetPassword(_password)
		}

		fileInfo := ArchiveFileInfo{
			Mode:     file.FileInfo().Mode(),
			Size:     file.FileInfo().Size(),
			IsDir:    file.FileInfo().IsDir(),
			ModTime:  file.FileInfo().ModTime(),
			Name:     file.FileInfo().Name(),
			FullPath: filepath.ToSlash(file.Name),
		}

		fileInfo.FullPath = fixDirSlash(fileInfo.IsDir, fileInfo.FullPath)

		allowIncludeFile := getFilteredFiles(fileInfo, _listDirectoryPath, _recursive, compiledGitIgnoreLines)

		if allowIncludeFile {
			filteredPaths = append(filteredPaths, fileInfo)
		}

		if !isListDirectoryPathExist && subpathExists(_listDirectoryPath, fileInfo.FullPath) {
			isListDirectoryPathExist = true
		}
	}

	if !isListDirectoryPathExist {
		return filteredPaths, fmt.Errorf("path not found to filter: %s", _listDirectoryPath)
	}

	sortedPaths := sortFiles(filteredPaths, _orderBy, _orderDir)

	return sortedPaths, err
}
