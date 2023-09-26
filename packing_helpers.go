package onearchiver

import (
	"fmt"
	ignore "github.com/sabhiram/go-gitignore"
	"github.com/samber/lo"
	"os"
	"path/filepath"
	"strings"
)

func getArchiveFilesRelativePath(absFilepath string, commonParentPath string) (string, error) {
	splittedFilepath := strings.Split(absFilepath, commonParentPath)

	return lo.Last(splittedFilepath)
}

func processFilesForPackingArchives(zipFilePathListMap *map[string]createArchiveFileInfo, fileList *[]string, commonParentPath string, gitIgnorePattern *[]string) (*archiveProgressMetrices[createArchiveFileInfo], error) {
	_zipFilePathListMap := *zipFilePathListMap
	_fileList := *fileList
	progressMetrices := newArchiveProgressMetrices[createArchiveFileInfo]()

	var ignoreList []string
	ignoreList = append(ignoreList, GlobalPatternDenylist...)
	ignoreList = append(ignoreList, *gitIgnorePattern...)

	ignoreMatches := ignore.CompileIgnoreLines(ignoreList...)

	for _, item := range _fileList {
		err := filepath.Walk(item, func(absFilepath string, fileInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			absFilepath = filepath.ToSlash(absFilepath)
			relativeFilePath := absFilepath

			if commonParentPath != "" {
				// if there is only one filepath in [_fileList]
				if len(_fileList) < 2 && _fileList[0] == commonParentPath {
					splittedFilepath := strings.Split(_fileList[0], PathSep)
					lastPartOfFilename, err := lo.Last(splittedFilepath)
					if err != nil {
						return fmt.Errorf(string(ErrorInternalErrorOccured)) //todo add and return the error as well to this
					}

					// then the selected folder name should be the root directory in the archive
					if isDirectory(_fileList[0]) {
						archiveFilesRelativePath, err := getArchiveFilesRelativePath(absFilepath, commonParentPath)
						if err != nil {
							return fmt.Errorf(string(ErrorInternalErrorOccured)) //todo add and return the error as well to this
						}

						relativeFilePath = filepath.Join(lastPartOfFilename, archiveFilesRelativePath)
					} else {
						// then the selected file should be in the root directory in the archive
						relativeFilePath = lastPartOfFilename
					}

				} else {
					relativeFilePath, err = getArchiveFilesRelativePath(absFilepath, commonParentPath)
					if err != nil {
						return err
					}
				}
			}

			isFileADir := fileInfo.IsDir()
			relativeFilePath = fixDirSlash(isFileADir, relativeFilePath)

			relativeFilePath = strings.TrimLeft(relativeFilePath, PathSep)

			// ignore the files if pattern matches
			if ignoreMatches.MatchesPath(relativeFilePath) {
				return nil
			}

			// when the commonpath is used to construct the relative path, the parent directories in the filepath list doesnt get written into the archive file
			if commonParentPath != "" && absFilepath != commonParentPath {
				if item == absFilepath {
					splittedPaths := strings.Split(relativeFilePath, PathSep)
					for pathIndex := range splittedPaths {
						_relativeFilePath := strings.Join(splittedPaths[:pathIndex+1], PathSep)

						// skip if filename is blank
						if _relativeFilePath == "" {
							continue
						}

						_absFilepath := filepath.Join(commonParentPath, _relativeFilePath)

						_fileInfo, err := os.Lstat(_absFilepath)
						if err != nil {
							return err
						}

						isDir := _fileInfo.IsDir()

						_absFilepath = fixDirSlash(isDir, _absFilepath)
						_relativeFilePath = fixDirSlash(isDir, _relativeFilePath)

						progressMetrices.updateArchiveProgressMetrices(_zipFilePathListMap, _absFilepath, _fileInfo.Size(), isDir)
						_zipFilePathListMap[_absFilepath] = createArchiveFileInfo{
							absFilepath:      _absFilepath,
							relativeFilePath: _relativeFilePath,
							isDir:            isDir,
							fileInfo:         &_fileInfo,
						}

					}

					return nil
				}
			}

			absFilepath = fixDirSlash(isFileADir, absFilepath)

			progressMetrices.updateArchiveProgressMetrices(_zipFilePathListMap, absFilepath, fileInfo.Size(), isFileADir)
			_zipFilePathListMap[absFilepath] = createArchiveFileInfo{
				absFilepath:      absFilepath,
				relativeFilePath: relativeFilePath,
				isDir:            isFileADir,
				fileInfo:         &fileInfo,
			}

			return nil
		})

		if err != nil {
			return progressMetrices, nil
		}
	}

	return progressMetrices, nil
}

func processFilesForPackingCompressedFile(zipFilePathListMap *map[string]createArchiveFileInfo, fileSourcePath string, gitIgnorePattern *[]string) (*archiveProgressMetrices[createArchiveFileInfo], error) {
	progressMetrices := newArchiveProgressMetrices[createArchiveFileInfo]()

	var ignoreList []string
	ignoreList = append(ignoreList, GlobalPatternDenylist...)
	ignoreList = append(ignoreList, *gitIgnorePattern...)

	absFilepath := filepath.ToSlash(fileSourcePath)
	ignoreMatches := ignore.CompileIgnoreLines(ignoreList...)

	fileInfo, err := os.Lstat(absFilepath)
	if err != nil {
		return progressMetrices, err
	}

	relativeFilePath := strings.TrimLeft(fileInfo.Name(), PathSep)

	// ignore the file if ignore files pattern matches
	if ignoreMatches.MatchesPath(relativeFilePath) {
		return progressMetrices, nil
	}

	progressMetrices.updateArchiveProgressMetrices(*zipFilePathListMap, absFilepath, fileInfo.Size(), false)
	(*zipFilePathListMap)[absFilepath] = createArchiveFileInfo{
		absFilepath:      absFilepath,
		relativeFilePath: relativeFilePath,
		isDir:            false,
		fileInfo:         &fileInfo,
	}

	return progressMetrices, nil
}
