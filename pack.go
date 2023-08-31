package onearchiver

import (
	"fmt"
	"github.com/ganeshrvel/archiver"
	ignore "github.com/sabhiram/go-gitignore"
	"github.com/wesovilabs/koazee"
	"os"
	"path/filepath"
	"strings"
)

func (arc zipArchive) doPack(ph *ProgressHandler) error {
	fileList := arc.pack.FileList

	commonParentPath := GetCommonParentPath(os.PathSeparator, fileList...)

	if indexExists(&fileList, 0) && commonParentPath == fileList[0] {
		commonParentPathSplitted := strings.Split(fileList[0], PathSep)

		commonParentPath = strings.Join(commonParentPathSplitted[:len(commonParentPathSplitted)-1], PathSep)
	}

	if err := createZipFile(&arc, fileList, commonParentPath, ph); err != nil {
		return err
	}

	return nil
}

func (arc commonArchive) doPack(ph *ProgressHandler) error {
	filename := arc.meta.Filename
	fileList := arc.pack.FileList

	arcFileObj, err := archiver.ByExtension(filename)

	if err != nil {
		return err
	}

	err = archiveFormat(&arcFileObj, "", OverwriteExisting)

	if err != nil {
		return err
	}

	commonParentPath := GetCommonParentPath(os.PathSeparator, fileList...)

	if indexExists(&fileList, 0) && commonParentPath == fileList[0] {
		commonParentPathSplitted := strings.Split(fileList[0], PathSep)

		commonParentPath = strings.Join(commonParentPathSplitted[:len(commonParentPathSplitted)-1], PathSep)
	}

	switch archFileWrite := arcFileObj.(type) {
	case *archiver.Tar:
		err = packTarballs(&arc, archFileWrite, &fileList, commonParentPath, ph)
	case *archiver.TarGz:
		err = packTarballs(&arc, archFileWrite, &fileList, commonParentPath, ph)
	case *archiver.TarBz2:
		err = packTarballs(&arc, archFileWrite, &fileList, commonParentPath, ph)
	case *archiver.TarBrotli:
		err = packTarballs(&arc, archFileWrite, &fileList, commonParentPath, ph)
	case *archiver.TarLz4:
		err = packTarballs(&arc, archFileWrite, &fileList, commonParentPath, ph)
	case *archiver.TarSz:
		err = packTarballs(&arc, archFileWrite, &fileList, commonParentPath, ph)
	case *archiver.TarXz:
		err = packTarballs(&arc, archFileWrite, &fileList, commonParentPath, ph)
	case *archiver.TarZstd:
		err = packTarballs(&arc, archFileWrite, &fileList, commonParentPath, ph)
	case *archiver.Gz:
		err = packCompressedFile(&arc, archFileWrite, &fileList, ph)
	case *archiver.Brotli:
		err = packCompressedFile(&arc, archFileWrite, &fileList, ph)
	case *archiver.Bz2:
		err = packCompressedFile(&arc, archFileWrite, &fileList, ph)
	case *archiver.Lz4:
		err = packCompressedFile(&arc, archFileWrite, &fileList, ph)
	case *archiver.Snappy:
		err = packCompressedFile(&arc, archFileWrite, &fileList, ph)
	case *archiver.Xz:
		err = packCompressedFile(&arc, archFileWrite, &fileList, ph)
	case *archiver.Zstd:
		err = packCompressedFile(&arc, archFileWrite, &fileList, ph)

	default:
		return fmt.Errorf(string(ErrorFormatSupported))
	}

	if err != nil {
		return err
	}

	return nil
}

func StartPacking(meta *ArchiveMeta, pack *ArchivePack, ph *ProgressHandler) error {
	_meta := *meta
	_pack := *pack

	var arcPackObj ArchivePacker

	ext := extension(_meta.Filename)

	if OverwriteExisting && FileExists(_meta.Filename) {
		if err := os.Remove(_meta.Filename); err != nil {
			return err
		}
	}

	switch ext {
	case "zip":
		arcPackObj = zipArchive{meta: _meta, pack: _pack}

		break

	default:
		arcPackObj = commonArchive{meta: _meta, pack: _pack}

		break
	}

	return arcPackObj.doPack(ph)
}

func getArchiveFilesRelativePath(absFilepath string, commonParentPath string) string {
	splittedFilepath := strings.Split(absFilepath, commonParentPath)

	koazeeStream := koazee.StreamOf(splittedFilepath)
	lastItem := koazeeStream.Last()

	return lastItem.String()
}

func processFilesForPackingArchives(zipFilePathListMap *map[string]createArchiveFileInfo, fileList *[]string, commonParentPath string, gitIgnorePattern *[]string) error {
	_zipFilePathListMap := *zipFilePathListMap
	_fileList := *fileList

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

					_koazeeStream := koazee.StreamOf(splittedFilepath)
					lastItem := _koazeeStream.Last()
					lastPartOfFilename := lastItem.String()

					// then the selected folder name should be the root directory in the archive
					if isDirectory(_fileList[0]) {
						archiveFilesRelativePath := getArchiveFilesRelativePath(absFilepath, commonParentPath)

						relativeFilePath = filepath.Join(lastPartOfFilename, archiveFilesRelativePath)
					} else {
						// then the selected file should be in the root directory in the archive
						relativeFilePath = lastPartOfFilename
					}

				} else {
					relativeFilePath = getArchiveFilesRelativePath(absFilepath, commonParentPath)
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

			_zipFilePathListMap[absFilepath] = createArchiveFileInfo{
				absFilepath:      absFilepath,
				relativeFilePath: relativeFilePath,
				isDir:            isFileADir,
				fileInfo:         &fileInfo,
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func processFilesForPackingCompressedFile(zipFilePathListMap *map[string]createArchiveFileInfo, fileSourcePath string, gitIgnorePattern *[]string) error {
	var ignoreList []string
	ignoreList = append(ignoreList, GlobalPatternDenylist...)
	ignoreList = append(ignoreList, *gitIgnorePattern...)

	absFilepath := filepath.ToSlash(fileSourcePath)
	ignoreMatches := ignore.CompileIgnoreLines(ignoreList...)

	fileInfo, err := os.Lstat(absFilepath)
	if err != nil {
		return err
	}

	relativeFilePath := strings.TrimLeft(fileInfo.Name(), PathSep)

	// ignore the file if pattern matches
	if ignoreMatches.MatchesPath(relativeFilePath) {
		return nil
	}

	(*zipFilePathListMap)[absFilepath] = createArchiveFileInfo{
		absFilepath:      absFilepath,
		relativeFilePath: relativeFilePath,
		isDir:            false,
		fileInfo:         &fileInfo,
	}

	return nil
}
