package one_archiver

import (
	"archive/tar"
	"github.com/ganeshrvel/archiver"
	"github.com/nwaples/rardecode"
	ignore "github.com/sabhiram/go-gitignore"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func startUnpackingCommonArchives(arc commonArchive, arcWalker interface{ archiver.Walker }, ph *ProgressHandler) error {
	_filename := arc.meta.Filename
	_gitIgnorePattern := arc.meta.GitIgnorePattern
	_fileList := arc.unpack.FileList
	_destination := arc.unpack.Destination

	allowFileFiltering := len(_fileList) > 0

	var ignoreList []string
	ignoreList = append(ignoreList, GlobalPatternDenylist...)
	ignoreList = append(ignoreList, _gitIgnorePattern...)

	ignoreMatches, _ := ignore.CompileIgnoreLines(ignoreList...)

	commonArchiveFilePathListMap := make(map[string]extractCommonArchiveFileInfo)

	err := arcWalker.Walk(_filename, func(file archiver.File) error {
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

		// not currently being used
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

		if allowFileFiltering {
			matched := StringFilter(_fileList, func(s string) bool {
				_fName := fixDirSlash(fileInfo.IsDir, fileInfo.FullPath)

				return subpathExists(s, _fName)
			})

			if len(matched) < 1 {
				return nil
			}
		}

		if ignoreMatches.MatchesPath(fileInfo.FullPath) {
			return nil
		}

		_absPath := filepath.Join(_destination, fileInfo.FullPath)

		fileData := make([]byte, file.Size())
		numBytesRead, err := file.Read(fileData)
		if err != nil && !(numBytesRead == int(file.Size()) && err == io.EOF) {
			return err
		}

		commonArchiveFilePathListMap[_absPath] = extractCommonArchiveFileInfo{
			absFilepath: _absPath,
			name:        fileInfo.Name,
			fileInfo:    &fileInfo,
			fileBytes:   &fileData,
		}

		return nil
	})

	totalFiles := len(commonArchiveFilePathListMap)
	pInfo, ch := initProgress(totalFiles, ph)

	count := 0
	for absolutePath, file := range commonArchiveFilePathListMap {
		count += 1
		pInfo.progress(ch, totalFiles, absolutePath, count)

		if err := addFileFromCommonArchiveToDisk(&file, absolutePath); err != nil {
			return err
		}
	}

	pInfo.endProgress(ch, totalFiles)

	return err
}

func addFileFromCommonArchiveToDisk(file *extractCommonArchiveFileInfo, filename string) error {
	if file.fileInfo.IsDir {
		if err := os.MkdirAll(filename, os.ModePerm); err != nil {
			return err
		}

		return nil
	} else {
		_basename := filepath.Dir(filename)

		if err := os.MkdirAll(_basename, os.ModePerm); err != nil {
			return err
		}
	}

	return ioutil.WriteFile(file.absFilepath, *file.fileBytes, file.fileInfo.Mode)
}
