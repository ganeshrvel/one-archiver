package onearchiver

import (
	"fmt"
	"github.com/ganeshrvel/archiver"
	"io"
	"os"
)

func packCompressedFile(session *Session, arc *commonArchive, arcFileCompressor interface{ archiver.Compressor }, fileList *[]string) error {
	destinationFilename := arc.meta.Filename
	gitIgnorePattern := arc.meta.GitIgnorePattern
	sourceFilepath := ""

	if len(*fileList) < 1 {
		return fmt.Errorf(string(ErrorCompressedFileNoFileFound))
	}

	if len(*fileList) > 1 {
		return fmt.Errorf(string(ErrorCompressedFileInvalidSize))
	}

	if len(*fileList) == 1 {
		sourceFilepath = (*fileList)[0]

		if isDirectory(sourceFilepath) {
			return fmt.Errorf(string(ErrorCompressedFileOnlyFileAllowed))
		}
	}

	destinationFileWriter, err := os.Create(destinationFilename)
	if err != nil {
		return err
	}
	defer func() {
		if err := destinationFileWriter.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	zipFilePathListMap := make(map[string]createArchiveFileInfo)
	progressMetrices, err := processFilesForPackingCompressedFile(&zipFilePathListMap, sourceFilepath, &gitIgnorePattern)
	if err != nil {
		return err
	}

	if progressMetrices.totalFiles < 1 {
		return fmt.Errorf(string(ErrorCompressedFileNoFileFound))
	}

	session.initializeProgress(progressMetrices.totalFiles, progressMetrices.totalSize)

	for absolutePath, item := range zipFilePathListMap {
		select {
		case <-session.isDone():
			return session.ctxError()
		default:
		}

		progressMetrices.updateArchiveFilesProgressCount(item.isDir)
		session.enableCtxCancel()
		session.fileProgress(absolutePath, progressMetrices.filesProgressCount)

		if err := addFileToCompressedFile(&arcFileCompressor, destinationFileWriter, item.absFilepath); err != nil {
			return err
		}
	}

	session.endProgress()

	return err
}

// todo add progress intruption ctxcopy
func addFileToCompressedFile(arcFileCompressor *interface{ archiver.Compressor }, destinationFileWriter io.Writer, sourceFilepath string) error {
	_arcFileCompressor := *arcFileCompressor

	fileToArchive, err := os.Open(sourceFilepath)
	if err != nil {
		return err
	}

	defer func() {
		if err := fileToArchive.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	// todo add a check if continue of error then dont return
	err = _arcFileCompressor.Compress(fileToArchive, destinationFileWriter)

	return err
}
