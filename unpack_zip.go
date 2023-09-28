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

		progressMetrices.updateArchiveProgressMetrices(zipFilePathListMap, destinationFileAbsPath, fileInfo.Size(), fileInfo.IsDir())
		zipFilePathListMap[destinationFileAbsPath] = extractZipFileInfo{
			absFilepath: destinationFileAbsPath,
			name:        fileName,
			fileInfo:    &fileInfo,
			zipFileInfo: file,
		}
	}

	session.initializeProgress(progressMetrices.totalFiles, progressMetrices.totalSize, true)

	for destinationFileAbsPath, file := range zipFilePathListMap {
		select {
		case <-session.isDone():
			session.endProgress(ProgressStatusCancelled)
			return session.ctxError()
		default:
		}

		progressMetrices.updateArchiveFilesProgressCount(file.zipFileInfo.FileInfo().IsDir())
		session.enableCtxCancel()
		session.fileProgress(destinationFileAbsPath, progressMetrices.filesProgressCount)

		if err := makeAddFileFromZipToDisk(session, file.zipFileInfo, destinationFileAbsPath, pctx); err != nil {
			return err
		}
	}

	if !exists(destinationPath) {
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
		return addFileFromZipToDisk(session, zippedFileToExtractInfo, destinationFileAbsPath, "")
	}

	for _, password := range pctx.passwords {
		err := addFileFromZipToDisk(session, zippedFileToExtractInfo, destinationFileAbsPath, password)

		if err == nil {
			return nil
		}

		if strings.Contains(err.Error(), string(ErrorInvalidPassword)) {
			continue
		}

		return err
	}

	return nil
}

func addFileFromZipToDisk(session *Session, zippedFileToExtractInfo *zip.File, destinationFileAbsPath string, password string) error {
	if len(password) > 0 {
		zippedFileToExtractInfo.SetPassword(password)
	}

	fileToExtract, err := zippedFileToExtractInfo.Open()
	if err != nil {
		return err
	}
	defer func() {
		if err := fileToExtract.Close(); err != nil {
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

	if isSymlink(zippedFileToExtractInfo.FileInfo()) {
		originalTargetPathBytes, err := io.ReadAll(fileToExtract)
		if err != nil {
			return err
		}

		originalTargetPath := string(originalTargetPathBytes)
		targetPathToWrite := filepath.ToSlash(originalTargetPath)

		err = os.Symlink(targetPathToWrite, destinationFileAbsPath)
		if errors.Is(err, zip.ErrChecksum) {
			return fmt.Errorf(string(ErrorInvalidPassword))
		}
		if err != nil {
			return err
		}

		session.symlinkSizeProgress(originalTargetPath, targetPathToWrite)

		// todo add a check if continue of error then dont return
		return nil
	}

	writer, err := os.Create(destinationFileAbsPath)
	if err != nil {
		return err
	}

	// todo add a check if continue of error then dont return
	numBytesWritten, err := SessionAwareCopy(session, writer, fileToExtract, zippedFileToExtractInfo.FileInfo().IsDir(), zippedFileToExtractInfo.FileInfo().Size())

	if errors.Is(err, zip.ErrChecksum) {
		session.revertSizeProgress(numBytesWritten)

		return fmt.Errorf(string(ErrorInvalidPassword))
	}

	return err
}
