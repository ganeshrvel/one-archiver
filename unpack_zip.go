package one_archiver

import (
	"fmt"
	ignore "github.com/sabhiram/go-gitignore"
	"github.com/yeka/zip"
	"io"
	"os"
	"path/filepath"
)

func startUnpackingZip(arc zipArchive, ph *ProgressHandler) error {
	_filename := arc.meta.Filename
	_password := arc.meta.Password
	_destination := arc.unpack.Destination
	_gitIgnorePattern := arc.meta.GitIgnorePattern
	_fileList := arc.unpack.FileList

	allowFileFiltering := len(_fileList) > 0

	reader, err := zip.OpenReader(_filename)
	if err != nil {
		return err
	}

	var ignoreList []string
	ignoreList = append(ignoreList, GlobalPatternDenylist...)
	ignoreList = append(ignoreList, _gitIgnorePattern...)

	ignoreMatches, _ := ignore.CompileIgnoreLines(ignoreList...)

	zipFilePathListMap := make(map[string]extractZipFileInfo)

	for _, file := range reader.File {
		if file.IsEncrypted() {
			file.SetPassword(_password)
		}

		fileName := filepath.ToSlash(file.Name)
		_fileInfo := file.FileInfo()

		if allowFileFiltering {
			matched := StringFilter(_fileList, func(s string) bool {
				_filterFName := fixDirSlash(_fileInfo.IsDir(), fileName)

				return subpathExists(s, _filterFName)
			})

			if len(matched) < 1 {
				continue
			}
		}

		if ignoreMatches.MatchesPath(fileName) {
			continue
		}

		_absPath := filepath.Join(_destination, fileName)

		zipFilePathListMap[_absPath] = extractZipFileInfo{
			absFilepath: _absPath,
			name:        fileName,
			fileInfo:    &_fileInfo,
			zipFileInfo: file,
		}
	}

	totalFiles := len(reader.File)
	pInfo, ch := initProgress(totalFiles, ph)

	count := 0
	for absolutePath, file := range zipFilePathListMap {
		count += 1
		pInfo.progress(ch, totalFiles, absolutePath, count)

		if err := addFileFromZipToDisk(file.zipFileInfo, absolutePath); err != nil {
			return err
		}
	}

	pInfo.endProgress(ch, totalFiles)

	defer func() {
		if err := reader.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	return nil
}

func addFileFromZipToDisk(file *zip.File, filename string) error {
	fileToExtract, err := file.Open()

	if err != nil {
		return err
	}

	defer func() {
		if err := fileToExtract.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	if file.FileInfo().IsDir() {
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

	writer, err := os.Create(filename)
	if err != nil {
		return err
	}

	_, _ = io.Copy(writer, fileToExtract)

	return err
}
