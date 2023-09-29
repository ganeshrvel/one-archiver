package onearchiver

import (
	"fmt"
	"github.com/ganeshrvel/yeka_zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func createZipFile(session *Session, arc *zipArchive, fileList []string, commonParentPath string) error {
	filename := arc.meta.Filename
	pctx := arc.read.passwordContext()

	gitIgnorePattern := arc.meta.GitIgnorePattern
	encryptionMethod := arc.meta.EncryptionMethod
	password := pctx.getSinglePassword()

	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if err := newZipFile.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	zipWriter := zip.NewWriter(newZipFile)
	defer func() {
		if err := zipWriter.Close(); err != nil {
			fmt.Println(err)
		}

		if err := zipWriter.Flush(); err != nil {
			fmt.Println(err)
		}
	}()

	zipFilePathListMap := make(map[string]createArchiveFileInfo)

	progressMetrices, err := processFilesForPackingArchives(&zipFilePathListMap, &fileList, commonParentPath, &gitIgnorePattern)
	if err != nil {
		return err
	}

	session.initializeProgress(progressMetrices.totalFiles, progressMetrices.totalSize, true)

	for destinationFileAbsPath, item := range zipFilePathListMap {
		select {
		case <-session.isDone():
			session.endProgress(ProgressStatusCancelled)
			return session.ctxError()
		default:
		}

		progressMetrices.updateArchiveFilesProgressCount(item.isDir)
		session.enableCtxCancel()
		session.fileProgress(destinationFileAbsPath, progressMetrices.filesProgressCount)

		if err := addFileToZip(session, zipWriter, *item.fileInfo, item.absFilepath, item.relativeFilePath, password, encryptionMethod); err != nil {
			return err
		}
	}

	session.endProgress(ProgressStatusCompleted)

	return err
}

func addFileToZip(
	session *Session,
	zipWriter *zip.Writer,
	fileInfo os.FileInfo,
	filename string,
	relativeFilename string,
	password string,
	encryptionMethod zip.EncryptionMethod,

) error {

	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}

	// see http://golang.org/pkg/archive/zip/#pkg-constants

	header.Name = relativeFilename
	header.Method = zip.Deflate
	header.SetModTime(fileInfo.ModTime())

	if password != "" {
		header.SetPassword(password)
		header.SetEncryptionMethod(encryptionMethod)
	}

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	if isSymlink(fileInfo) {
		originalTargetPath, err := os.Readlink(filename)
		if err != nil {
			return err
		}

		// todo add a check if continue of error then dont return
		// Write symlink's target to writer - file's body for symlinks is the symlink target.
		targetPathToWrite := filepath.ToSlash(originalTargetPath)
		_, _ = writer.Write([]byte(targetPathToWrite))
		if err != nil {
			return err
		}

		session.symlinkSizeProgress(originalTargetPath, targetPathToWrite)

		return nil
	}

	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		if err := fileToZip.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	if fileInfo.IsDir() {
		_, err = io.Copy(writer, fileToZip)

		// todo add a check if continue of error then dont return
		// Check for the specific error where the source is a directory.
		// If the error indicates that the source "is a directory", we ignore it and return nil.
		if strings.Contains(err.Error(), "is a directory") {
			return nil
		}

		return err
	}

	// todo add a check if continue of error then dont return
	_, err = SessionAwareCopy(session, writer, fileToZip, fileInfo.IsDir(), fileInfo.Size())

	return err
}
