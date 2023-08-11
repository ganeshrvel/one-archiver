package onearchiver

import (
	"fmt"
	ignore "github.com/sabhiram/go-gitignore"
	"github.com/yeka/zip"
	"io"
	"os"
	"path/filepath"
)

func startUnpackingZip(arc zipArchive, ph *ProgressHandler) error {
	sourceFilepath := arc.meta.Filename
	password := arc.meta.Password
	destinationPath := arc.unpack.Destination
	gitIgnorePattern := arc.meta.GitIgnorePattern
	fileList := arc.unpack.FileList

	allowFileFiltering := len(fileList) > 0

	sourceReader, err := zip.OpenReader(sourceFilepath)
	if err != nil {
		return err
	}
	defer func() {
		if err := sourceReader.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	var ignoreList []string
	ignoreList = append(ignoreList, GlobalPatternDenylist...)
	ignoreList = append(ignoreList, gitIgnorePattern...)

	ignoreMatches := ignore.CompileIgnoreLines(ignoreList...)
	zipFilePathListMap := make(map[string]extractZipFileInfo)

	for _, file := range sourceReader.File {
		if file.IsEncrypted() {
			file.SetPassword(password)
		}

		fileName := filepath.ToSlash(file.Name)
		fileInfo := file.FileInfo()

		if allowFileFiltering {
			matched := StringFilter(fileList, func(s string) bool {
				filterFName := fixDirSlash(fileInfo.IsDir(), fileName)

				return subpathExists(s, filterFName)
			})

			if len(matched) < 1 {
				continue
			}
		}

		if ignoreMatches.MatchesPath(fileName) {
			continue
		}

		destinationFileAbsPath := filepath.Join(destinationPath, fileName)

		zipFilePathListMap[destinationFileAbsPath] = extractZipFileInfo{
			absFilepath: destinationFileAbsPath,
			name:        fileName,
			fileInfo:    &fileInfo,
			zipFileInfo: file,
		}
	}

	totalFiles := len(sourceReader.File)
	pInfo, ch := initProgress(totalFiles, ph)

	count := 0
	for destinationFileAbsPath, file := range zipFilePathListMap {
		count += 1
		pInfo.progress(ch, totalFiles, destinationFileAbsPath, count)

		if err := addFileFromZipToDisk(file.zipFileInfo, destinationFileAbsPath); err != nil {
			return err
		}
	}

	pInfo.endProgress(ch, totalFiles)

	if !exists(destinationPath) {
		if err := os.Mkdir(destinationPath, 0755); err != nil {
			return err
		}
	}

	return nil
}

func addFileFromZipToDisk(file *zip.File, destinationFileAbsPath string) error {
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
		if err := os.MkdirAll(destinationFileAbsPath, os.ModePerm); err != nil {
			return err
		}

		return nil
	} else {
		parent := filepath.Dir(destinationFileAbsPath)

		if err := os.MkdirAll(parent, os.ModePerm); err != nil {
			return err
		}
	}

	writer, err := os.Create(destinationFileAbsPath)
	if err != nil {
		return err
	}

	_, _ = io.Copy(writer, fileToExtract)

	return err
}
