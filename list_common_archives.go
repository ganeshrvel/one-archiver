package onearchiver

import (
	"archive/tar"
	"fmt"
	"github.com/ganeshrvel/archiver"
	"github.com/nwaples/rardecode"
	ignore "github.com/sabhiram/go-gitignore"
	"path/filepath"
)

// List files in the common archives
func (arc commonArchive) list() ([]ArchiveFileInfo, error) {
	filename := arc.meta.Filename
	password := arc.meta.Password
	listDirectoryPath := arc.read.ListDirectoryPath
	recursive := arc.read.Recursive
	orderBy := arc.read.OrderBy
	orderDir := arc.read.OrderDir
	gitIgnorePattern := arc.meta.GitIgnorePattern

	arcFileObj, err := archiver.ByExtension(filename)

	if err != nil {
		return nil, err
	}

	err = archiveFormat(&arcFileObj, password, OverwriteExisting)
	if err != nil {
		return nil, err
	}

	var arcWalker, ok = arcFileObj.(archiver.Walker)
	if !ok {
		return nil, fmt.Errorf(string(ErrorArchiverList))
	}

	var filteredPaths []ArchiveFileInfo

	isListDirectoryPathExist := listDirectoryPath == ""

	var ignoreList []string
	ignoreList = append(ignoreList, GlobalPatternDenylist...)
	ignoreList = append(ignoreList, gitIgnorePattern...)
	compiledGitIgnoreLines := ignore.CompileIgnoreLines(ignoreList...)

	err = arcWalker.Walk(filename, func(file archiver.File) error {
		var fileInfo ArchiveFileInfo

		switch fileHeader := file.Header.(type) {
		case *tar.Header:
			fullPath := filepath.ToSlash(fileHeader.Name)
			isDir := file.IsDir()
			name := file.Name()

			fileInfo = ArchiveFileInfo{
				Mode:       file.Mode(),
				Size:       file.Size(),
				IsDir:      isDir,
				ModTime:    file.ModTime(),
				Name:       name,
				FullPath:   fullPath,
				ParentPath: GetParentDirectory(fullPath),
				Extension:  extension(name),
			}

		case *rardecode.FileHeader:

			isDir := file.IsDir()
			fullPath := fixDirSlash(isDir, filepath.ToSlash(file.Name()))
			name := filepath.Base(fullPath)

			fileInfo = ArchiveFileInfo{
				Mode:       file.Mode(),
				Size:       file.Size(),
				IsDir:      isDir,
				ModTime:    file.ModTime(),
				Name:       name,
				FullPath:   fullPath,
				ParentPath: GetParentDirectory(fullPath),
				Extension:  extension(name),
			}

		// not being used
		default:
			fullPath := filepath.ToSlash(file.FileInfo.Name())
			isDir := file.IsDir()
			name := file.Name()

			fileInfo = ArchiveFileInfo{
				Mode:       file.Mode(),
				Size:       file.Size(),
				IsDir:      isDir,
				ModTime:    file.ModTime(),
				Name:       name,
				FullPath:   fullPath,
				ParentPath: GetParentDirectory(fullPath),
				Extension:  extension(name),
			}
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

		return nil
	})

	if !isListDirectoryPathExist {
		return filteredPaths, fmt.Errorf("%s: %s", string(ErrorNoPathToFilter), listDirectoryPath)
	}

	if arc.read.OrderDir == OrderDirNone {
		return filteredPaths, err
	}

	sortedPaths := sortFiles(filteredPaths, orderBy, orderDir)

	return sortedPaths, err
}
