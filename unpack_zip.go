package onearchiver

import (
	"fmt"
	"github.com/ganeshrvel/yeka_zip"
	ignore "github.com/sabhiram/go-gitignore"
	"io"
	"os"
	"path/filepath"
)

func startUnpackingZip(arc zipArchive, ph *ProgressHandler) error {
	sourceFilepath := arc.meta.Filename
	destinationPath := arc.unpack.Destination
	gitIgnorePattern := arc.meta.GitIgnorePattern
	fileList := arc.unpack.FileList

	pctx := arc.unpack.passwordContext()

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

		if err := makeAddFileFromZipToDisk(file.zipFileInfo, destinationFileAbsPath, &pctx); err != nil {
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

func makeAddFileFromZipToDisk(zipFileInfo *zip.File, destinationFileAbsPath string, pctx *PasswordContext) error {
	isEncrypted := zipFileInfo.IsEncrypted()

	if !isEncrypted {
		return addFileFromZipToDisk(zipFileInfo, destinationFileAbsPath, "")
	}

	for _, password := range pctx.passwords {
		err := addFileFromZipToDisk(zipFileInfo, destinationFileAbsPath, password)

		if err == nil {
			return nil
		}

		if err == zip.ErrChecksum {
			continue
		}

		return err
	}

	return nil
}

func addFileFromZipToDisk(zipFileInfo *zip.File, destinationFileAbsPath string, password string) error {

	if len(password) > 0 {
		zipFileInfo.SetPassword(password)
	}

	fileToExtract, err := zipFileInfo.Open()
	if err != nil {
		return err
	}
	defer func() {
		if err := fileToExtract.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	if zipFileInfo.FileInfo().IsDir() {
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

	if isSymlink(zipFileInfo.FileInfo()) {
		targetBytes, err := io.ReadAll(fileToExtract)
		if err != nil {
			return err
		}

		targetPath := filepath.ToSlash(string(targetBytes))

		err = os.Symlink(targetPath, destinationFileAbsPath)

		if err == zip.ErrChecksum {
			return err
		}
		// todo add a check if continue of error then dont return
		return err
	}

	writer, err := os.Create(destinationFileAbsPath)
	if err != nil {
		return err
	}

	// todo add a check if continue of error then dont return
	_, err = io.Copy(writer, fileToExtract)
	if err == zip.ErrChecksum {
		return err
	}

	return err
}
