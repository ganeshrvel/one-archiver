package onearchiver

// TODO proper error handling

import (
	"fmt"
	"github.com/ganeshrvel/yeka_zip"
	ignore "github.com/sabhiram/go-gitignore"
	"os"
	"path/filepath"
)

// list files in zip archives
// yeka package is used here to list encrypted zip files
func (arc zipArchive) list() ([]ArchiveFileInfo, error) {
	filename := arc.meta.Filename
	listDirectoryPath := arc.read.ListDirectoryPath
	recursive := arc.read.Recursive
	orderBy := arc.read.OrderBy
	orderDir := arc.read.OrderDir
	gitIgnorePattern := arc.meta.GitIgnorePattern

	arcFileStat, err := os.Lstat(filename)
	if err != nil {
		return nil, err
	}
	reader, err := zip.OpenReader(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = reader.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	var filteredPaths []ArchiveFileInfo

	isListDirectoryPathExist := listDirectoryPath == ""

	var ignoreList []string
	ignoreList = append(ignoreList, GlobalPatternDenylist...)
	ignoreList = append(ignoreList, gitIgnorePattern...)
	compiledGitIgnoreLines := ignore.CompileIgnoreLines(ignoreList...)

	for _, file := range reader.File {
		fullPath := filepath.ToSlash(file.Name)
		isDir := file.FileInfo().IsDir()
		name := file.FileInfo().Name()

		fileInfo := ArchiveFileInfo{
			Mode:       file.FileInfo().Mode(),
			Size:       file.FileInfo().Size(),
			IsDir:      isDir,
			ModTime:    sanitizeTime(file.FileInfo().ModTime(), arcFileStat.ModTime()),
			Name:       name,
			FullPath:   fullPath,
			ParentPath: GetParentDirectory(fullPath),
			Extension:  extension(name),
		}

		fileInfo.FullPath = fixDirSlash(fileInfo.IsDir, fileInfo.FullPath)

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
	}

	if !isListDirectoryPathExist {
		return filteredPaths, fmt.Errorf("%s: %s", string(ErrorNoPathToFilter), listDirectoryPath)
	}

	sortedPaths := sortFiles(filteredPaths, orderBy, orderDir)

	return sortedPaths, err
}
