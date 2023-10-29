package onearchiver

import (
	"archive/tar"
	"fmt"
	"github.com/ganeshrvel/archiver"
	"github.com/ganeshrvel/rardecode"
	ignore "github.com/sabhiram/go-gitignore"
	"os"
	"path/filepath"
)

// List files in the common archives
func (arc commonArchive) list() ([]ArchiveFileInfo, error) {
	filename := arc.meta.Filename
	pctx := arc.read.passwordContext()
	listDirectoryPath := arc.read.ListDirectoryPath
	recursive := arc.read.Recursive
	orderBy := arc.read.OrderBy
	orderDir := arc.read.OrderDir
	gitIgnorePattern := arc.meta.GitIgnorePattern

	arcFileObj, err := archiver.ByExtension(filename)
	if err != nil {
		return nil, err
	}
	arcFileStat, err := os.Lstat(filename)
	if err != nil {
		return nil, err
	}

	err = archiveFormat(&arcFileObj, pctx, OverwriteExisting)
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
			isDir := file.IsDir()
			fullPath := filepath.ToSlash(fileHeader.Name)
			name := file.Name()

			size := file.Size()
			if TarFileLinkType(file, fileHeader).isLink() {
				_, link, err := getCommonArchivesTargetLinkPath(&file)

				if err != nil {
					return err
				}

				size = int64(len(link))
			}

			fileInfo = ArchiveFileInfo{
				Mode:       file.Mode(),
				Size:       size,
				IsDir:      isDir,
				ModTime:    sanitizeTime(file.ModTime(), arcFileStat.ModTime()),
				Name:       name,
				FullPath:   fullPath,
				ParentPath: GetParentDirectory(fullPath),
				Extension:  Extension(name),
			}

		case *rardecode.FileHeader:
			isDir := file.IsDir()
			fullPath := FixDirSlash(isDir, filepath.ToSlash(file.Name()))
			name := filepath.Base(fullPath)

			size := file.Size()
			if RarFileLinkType(file, fileHeader).isLink() {
				_, link, err := getCommonArchivesTargetLinkPath(&file)

				if err != nil {
					return err
				}

				size = int64(len(link))
			}

			fileInfo = ArchiveFileInfo{
				Mode:       file.Mode(),
				Size:       size,
				IsDir:      isDir,
				ModTime:    sanitizeTime(file.ModTime(), arcFileStat.ModTime()),
				Name:       name,
				FullPath:   fullPath,
				ParentPath: GetParentDirectory(fullPath),
				Extension:  Extension(name),
			}
		}

		fileInfo.FullPath = FixDirSlash(fileInfo.IsDir, fileInfo.FullPath)

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
