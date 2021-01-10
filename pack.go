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
	_fileList := arc.pack.FileList

	commonParentPath := GetCommonParentPath(os.PathSeparator, _fileList...)

	if indexExists(&_fileList, 0) && commonParentPath == _fileList[0] {
		commonParentPathSplitted := strings.Split(_fileList[0], PathSep)

		commonParentPath = strings.Join(commonParentPathSplitted[:len(commonParentPathSplitted)-1], PathSep)
	}

	if err := createZipFile(&arc, _fileList, commonParentPath, ph); err != nil {
		return err
	}

	return nil
}

func (arc commonArchive) doPack(ph *ProgressHandler) error {
	_filename := arc.meta.Filename
	_fileList := arc.pack.FileList

	arcFileObj, err := archiver.ByExtension(_filename)

	if err != nil {
		return err
	}

	err = archiveFormat(&arcFileObj, "", OverwriteExisting)

	if err != nil {
		return err
	}

	commonParentPath := GetCommonParentPath(os.PathSeparator, _fileList...)

	if indexExists(&_fileList, 0) && commonParentPath == _fileList[0] {
		commonParentPathSplitted := strings.Split(_fileList[0], PathSep)

		commonParentPath = strings.Join(commonParentPathSplitted[:len(commonParentPathSplitted)-1], PathSep)
	}

	switch archValue := arcFileObj.(type) {
	case *archiver.Tar:
		err = packTarballs(&arc, archValue, &_fileList, commonParentPath, ph)
	case *archiver.TarGz:
		err = packTarballs(&arc, archValue, &_fileList, commonParentPath, ph)
	case *archiver.TarBz2:
		err = packTarballs(&arc, archValue, &_fileList, commonParentPath, ph)
	case *archiver.TarBrotli:
		err = packTarballs(&arc, archValue, &_fileList, commonParentPath, ph)
	case *archiver.TarLz4:
		err = packTarballs(&arc, archValue, &_fileList, commonParentPath, ph)
	case *archiver.TarSz:
		err = packTarballs(&arc, archValue, &_fileList, commonParentPath, ph)
	case *archiver.TarXz:
		err = packTarballs(&arc, archValue, &_fileList, commonParentPath, ph)
	case *archiver.TarZstd:
		err = packTarballs(&arc, archValue, &_fileList, commonParentPath, ph)

	// Todo: parking the development of file compressors for now.
	// It requires a different logic for listing, compressing and uncompressing
	//case *archiver.Gz:
	//	err = packCompressFile(&arc, archValue, &_fileList)
	//case *archiver.Brotli:
	//	err = packCompressFile(&arc, archValue, &_fileList)
	//case *archiver.Bz2:
	//	err = packCompressFile(&arc, archValue, &_fileList)
	//case *archiver.Lz4:
	//	err = packCompressFile(&arc, archValue, &_fileList)
	//case *archiver.Snappy:
	//	err = packCompressFile(&arc, archValue, &_fileList)
	//case *archiver.Xz:
	//	err = packCompressFile(&arc, archValue, &_fileList)
	//case *archiver.Zstd:
	//	err = packCompressFile(&arc, archValue, &_fileList)

	default:
		return fmt.Errorf("archive file format is not supported")
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

	ext := filepath.Ext(_meta.Filename)

	if OverwriteExisting && FileExists(_meta.Filename) {
		if err := os.Remove(_meta.Filename); err != nil {
			return err
		}
	}

	switch ext {
	case ".zip":
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

	_koazeeStream := koazee.StreamOf(splittedFilepath)
	lastItem := _koazeeStream.Last()

	return lastItem.String()
}

func processFilesForPacking(zipFilePathListMap *map[string]createArchiveFileInfo, fileList *[]string, commonParentPath string, gitIgnorePattern *[]string) error {
	_zipFilePathListMap := *zipFilePathListMap
	_fileList := *fileList

	var ignoreList []string
	ignoreList = append(ignoreList, GlobalPatternDenylist...)
	ignoreList = append(ignoreList, *gitIgnorePattern...)

	ignoreMatches, _ := ignore.CompileIgnoreLines(ignoreList...)

	for _, item := range _fileList {
		err := filepath.Walk(item, func(absFilepath string, fileInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if isSymlink(fileInfo) {
				return nil
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
