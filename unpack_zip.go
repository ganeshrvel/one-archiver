package onearchiver

import (
	"errors"
	"fmt"
	"github.com/ganeshrvel/yeka_zip"
	ignore "github.com/sabhiram/go-gitignore"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func startUnpackingZip(session *Session, arc zipArchive) error {
	sourceFilepath := arc.meta.Filename
	destinationPath := arc.unpack.Destination
	gitIgnorePattern := arc.meta.GitIgnorePattern
	fileList := arc.unpack.FileList
	progressStreamDebounceTime := arc.unpack.ProgressStreamDebounceTime

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

	progressMetrices := newArchiveProgressMetrices[extractZipFileInfo]()

	for _, file := range sourceReader.File {
		fileName := filepath.ToSlash(file.Name)
		fileInfo := file.FileInfo()

		if allowFileFiltering {
			matched := StringFilter(fileList, func(s string) bool {
				filterFName := FixDirSlash(fileInfo.IsDir(), fileName)

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

		progressMetrices.updateArchiveProgressMetrices(zipFilePathListMap, destinationFileAbsPath, fileInfo.Size(), fileInfo.IsDir())
		zipFilePathListMap[destinationFileAbsPath] = extractZipFileInfo{
			absFilepath: destinationFileAbsPath,
			name:        fileName,
			fileInfo:    &fileInfo,
			zipFileInfo: file,
		}
	}

	session.initializeProgress(progressMetrices.totalFiles, progressMetrices.totalSize, progressStreamDebounceTime, true)

	for destinationFileAbsPath, file := range zipFilePathListMap {
		select {
		case <-session.isDone():
			session.endProgress(ProgressStatusCancelled)
			return session.ctxError()
		default:
		}

		progressMetrices.updateArchiveFilesProgressCount(file.zipFileInfo.FileInfo().IsDir())
		if err := session.fileProgress(destinationFileAbsPath, progressMetrices.filesProgressCount, file.zipFileInfo.FileInfo().IsDir(), func() error {
			return makeAddFileFromZipToDisk(session, file.zipFileInfo, destinationFileAbsPath, pctx)
		}); err != nil {
			return err
		}
	}

	if !Exists(destinationPath) {
		if err := os.Mkdir(destinationPath, 0755); err != nil {
			return err
		}
	}

	session.endProgress(ProgressStatusCompleted)

	return err
}

func makeAddFileFromZipToDisk(session *Session, zippedFileToExtractInfo *zip.File, destinationFileAbsPath string, pctx *PasswordContext) error {
	isEncrypted := zippedFileToExtractInfo.IsEncrypted()

	if !isEncrypted {
		return addFileFromZipToDisk(session, zippedFileToExtractInfo, destinationFileAbsPath, "", isEncrypted)
	}

	if !pctx.hasPasswords() {
		return fmt.Errorf(string(ErrorInvalidPassword))
	}

	for _, password := range pctx.passwords {
		err := addFileFromZipToDisk(session, zippedFileToExtractInfo, destinationFileAbsPath, password, isEncrypted)

		if err == nil {
			return nil
		}

		if strings.Contains(err.Error(), string(ErrorInvalidPassword)) {
			continue
		}

		return err
	}

	return fmt.Errorf(string(ErrorInvalidPassword))
}

func addFileFromZipToDisk(session *Session, zippedFileToExtractInfo *zip.File, destinationFileAbsPath string, password string, isEncrypted bool) error {
	if isEncrypted {
		if len(password) < 1 {
			return fmt.Errorf(string(ErrorInvalidPassword))
		}

		zippedFileToExtractInfo.SetPassword(password)
	}

	fileToExtract, err := zippedFileToExtractInfo.Open()
	if err != nil {
		return err
	}
	defer func() {
		err := fileToExtract.Close()

		if err != nil {
			zipPasswordErr := zipIncorrectPasswordErrorHandling(err, isEncrypted)
			if zipPasswordErr != nil {
				return
			}

			fmt.Printf("%v\n", err)
		}

	}()

	if zippedFileToExtractInfo.FileInfo().IsDir() {
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

	if IsSymlink(zippedFileToExtractInfo.FileInfo()) {
		originalTargetPathBytes, err := io.ReadAll(fileToExtract)
		if err != nil {
			zipPasswordErr := zipIncorrectPasswordErrorHandling(err, isEncrypted)
			if zipPasswordErr != nil {
				return zipPasswordErr
			}
			return err
		}

		originalTargetPath := string(originalTargetPathBytes)
		targetPathToWrite := filepath.ToSlash(originalTargetPath)

		err = os.Symlink(targetPathToWrite, destinationFileAbsPath)
		zipPasswordErr := zipIncorrectPasswordErrorHandling(err, isEncrypted)
		if zipPasswordErr != nil {
			return zipPasswordErr
		}
		if err != nil {
			return err
		}

		session.linkSizeProgress(originalTargetPath, targetPathToWrite)

		// todo add a check if continue of error then dont return
		return nil
	}

	writer, err := os.Create(destinationFileAbsPath)
	if err != nil {
		return err
	}

	// todo add a check if continue of error then dont return
	numBytesWritten, err := SessionAwareCopy(session, writer, fileToExtract, zippedFileToExtractInfo.FileInfo().IsDir(), zippedFileToExtractInfo.FileInfo().Size())

	zipPasswordErr := zipIncorrectPasswordErrorHandling(err, isEncrypted)
	if zipPasswordErr != nil {
		session.revertSizeProgress(numBytesWritten)

		return zipPasswordErr
	}

	return err
}

func zipIncorrectPasswordErrorHandling(zipError error, isEncrypted bool) error {
	if zipError != nil && isEncrypted {
		if errors.Is(zipError, zip.ErrChecksum) {
			return fmt.Errorf(string(ErrorInvalidPassword))
		}

		if strings.Contains(zipError.Error(), "corrupt input") {
			return fmt.Errorf(string(ErrorInvalidPassword))
		}

		if strings.Contains(zipError.Error(), "unexpected EOF") {
			return fmt.Errorf(string(ErrorInvalidPassword))
		}

	}

	return nil
}
