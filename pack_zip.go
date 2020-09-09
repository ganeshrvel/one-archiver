package one_archiver

import (
	"fmt"
	"github.com/yeka/zip"
	"io"
	"os"
)

func createZipFile(arc *zipArchive, fileList []string, commonParentPath string, ph *ProgressHandler) error {
	_filename := arc.meta.Filename
	_password := arc.meta.Password
	_gitIgnorePattern := arc.meta.GitIgnorePattern
	_encryptionMethod := arc.meta.EncryptionMethod

	newZipFile, err := os.Create(_filename)
	if err != nil {
		return err
	}

	defer func() {
		if err := newZipFile.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	zipWriter := zip.NewWriter(newZipFile)

	zipFilePathListMap := make(map[string]createArchiveFileInfo)

	err = processFilesForPacking(&zipFilePathListMap, &fileList, commonParentPath, &_gitIgnorePattern)
	if err != nil {
		return err
	}

	totalFiles := len(zipFilePathListMap)
	pInfo, ch := initProgress(totalFiles, ph)

	count := 0
	for absolutePath, item := range zipFilePathListMap {
		count += 1
		pInfo.progress(ch, totalFiles, absolutePath, count)

		if _password == "" {
			if err := addFileToRegularZip(zipWriter, *item.fileInfo, item.absFilepath, item.relativeFilePath); err != nil {
				return err
			}
		} else if err := addFileToEncryptedZip(zipWriter, item.absFilepath, item.relativeFilePath, _password, _encryptionMethod); err != nil {
			return err
		}
	}

	pInfo.endProgress(ch, totalFiles)

	defer func() {
		if err := zipWriter.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	defer func() {
		if err := zipWriter.Flush(); err != nil {
			fmt.Println(err)
		}
	}()

	return err
}

func addFileToRegularZip(zipWriter *zip.Writer, fileInfo os.FileInfo, filename string, relativeFilename string) error {
	fileToZip, err := os.Open(filename)

	if err != nil {
		return err
	}

	defer func() {
		if err := fileToZip.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	header, err := zip.FileInfoHeader(fileInfo)

	if err != nil {
		return err
	}

	header.Name = relativeFilename

	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, _ = io.Copy(writer, fileToZip)

	return err
}

func addFileToEncryptedZip(zipWriter *zip.Writer, filename string, relativeFilename string, password string,
	encryptionMethod zip.EncryptionMethod) error {
	fileToZip, err := os.Open(filename)

	if err != nil {
		return err
	}

	defer func() {
		if err := fileToZip.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	writer, err := zipWriter.Encrypt(relativeFilename, password, encryptionMethod)

	if err != nil {
		return err
	}

	_, _ = io.Copy(writer, fileToZip)

	return err
}
