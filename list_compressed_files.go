package onearchiver

import (
	"fmt"
	ignore "github.com/sabhiram/go-gitignore"
	"os"
	"path/filepath"
	"strings"
)

func readCompressedFiles(filename string) (fileInfo ArchiveFileInfo, error error) {
	afi := ArchiveFileInfo{}
	compressFileExt := filepath.Ext(filename)

	file, err := os.Lstat(filename)
	if err != nil {
		return afi, err
	}

	strippedFileName := strings.TrimRight(file.Name(), compressFileExt)
	fileExt := extension(strippedFileName)
	fullPath := filepath.ToSlash(strippedFileName)

	afi.ModTime = file.ModTime()
	afi.IsDir = false
	afi.Name = strippedFileName
	afi.FullPath = fullPath
	afi.ParentPath = GetParentDirectory(fullPath)
	afi.Extension = fileExt

	return afi, nil
}

// List files in the compressed file
func (arc compressedFile) list() ([]ArchiveFileInfo, error) {
	filename := arc.meta.Filename
	listDirectoryPath := arc.read.ListDirectoryPath
	recursive := arc.read.Recursive
	gitIgnorePattern := arc.meta.GitIgnorePattern

	fileInfo, err := readCompressedFiles(filename)
	if err != nil {
		return nil, fmt.Errorf(string(ErrorArchiverList))
	}

	isListDirectoryPathExist := listDirectoryPath == ""
	var filteredPaths []ArchiveFileInfo

	var ignoreList []string
	ignoreList = append(ignoreList, GlobalPatternDenylist...)
	ignoreList = append(ignoreList, gitIgnorePattern...)
	compiledGitIgnoreLines := ignore.CompileIgnoreLines(ignoreList...)

	includeFile := getFilteredFiles(
		fileInfo, listDirectoryPath, recursive,
	)

	if includeFile {
		if !compiledGitIgnoreLines.MatchesPath(fileInfo.FullPath) {
			filteredPaths = append(filteredPaths, fileInfo)
		}
	}

	if !isListDirectoryPathExist && subpathExists(listDirectoryPath, fileInfo.FullPath) {
		isListDirectoryPathExist = true
	}

	if !isListDirectoryPathExist {
		return filteredPaths, fmt.Errorf("%s: %s", string(ErrorNoPathToFilter), listDirectoryPath)
	}

	return filteredPaths, err
}
