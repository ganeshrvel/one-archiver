package onearchiver

import (
	"fmt"
	"github.com/ganeshrvel/archiver"
	"io"
	"os"
)

func packCompressedFile(arc *commonArchive, arcFileCompressor interface{ archiver.Compressor }, fileList *[]string, ph *ProgressHandler) error {
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
	err = processFilesForPackingCompressedFile(&zipFilePathListMap, sourceFilepath, &gitIgnorePattern)
	if err != nil {
		return err
	}

	totalFiles := len(zipFilePathListMap)

	if totalFiles < 1 {
		return fmt.Errorf(string(ErrorCompressedFileNoFileFound))
	}

	pInfo, ch := initProgress(totalFiles, ph)

	count := 0
	for absolutePath, item := range zipFilePathListMap {
		count += 1
		pInfo.progress(ch, totalFiles, absolutePath, count)

		if err := addFileToCompressedFile(&arcFileCompressor, destinationFileWriter, item.absFilepath); err != nil {
			return err
		}
	}

	pInfo.endProgress(ch, totalFiles)

	return err
}

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
