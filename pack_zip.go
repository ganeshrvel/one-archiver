package onearchiver

import (
	"fmt"
	"github.com/ganeshrvel/yeka_zip"
	"io"
	"os"
	"path/filepath"
)

func createZipFile(arc *zipArchive, fileList []string, commonParentPath string, ph *ProgressHandler) error {
	filename := arc.meta.Filename
	password := arc.meta.Password
	gitIgnorePattern := arc.meta.GitIgnorePattern
	encryptionMethod := arc.meta.EncryptionMethod

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

	err = processFilesForPackingArchives(&zipFilePathListMap, &fileList, commonParentPath, &gitIgnorePattern)
	if err != nil {
		return err
	}

	totalFiles := len(zipFilePathListMap)
	pInfo, ch := initProgress(totalFiles, ph)

	count := 0
	for absolutePath, item := range zipFilePathListMap {
		count += 1
		pInfo.progress(ch, totalFiles, absolutePath, count)

		if err := addFileToZip(zipWriter, *item.fileInfo, item.absFilepath, item.relativeFilePath, password, encryptionMethod); err != nil {
			return err
		}
	}

	pInfo.endProgress(ch, totalFiles)

	return err
}

func addFileToZip(
	zipWriter *zip.Writer,
	fileInfo os.FileInfo,
	filename string,
	relativeFilename string,
	password string,
	encryptionMethod zip.EncryptionMethod,

) error {
	var reader io.Reader

	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}

	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate
	header.Name = relativeFilename

	if password != "" {
		header.SetModTime(fileInfo.ModTime())
		header.SetPassword(password)
		header.SetEncryptionMethod(encryptionMethod)
	}

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	if isSymlink(fileInfo) {
		target, err := os.Readlink(filename)
		if err != nil {
			return err
		}

		// Write symlink's target to writer - file's body for symlinks is the symlink target.
		// todo add a check if continue of error then dont return
		_, _ = writer.Write([]byte(filepath.ToSlash(target)))
		if err != nil {
			return err
		}

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
	reader = fileToZip

	// todo add a check if continue of error then dont return
	_, _ = io.Copy(writer, reader)

	return err
}
