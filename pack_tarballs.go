package onearchiver

import (
	"fmt"
	"github.com/ganeshrvel/archiver"
	"os"
	"strings"
)

func packTarballs(arc *commonArchive, arcFileObj interface{ archiver.Writer }, fileList *[]string, commonParentPath string, ph *ProgressHandler) error {
	filename := arc.meta.Filename
	gitIgnorePattern := arc.meta.GitIgnorePattern

	out, err := os.Create(filename)
	if err != nil {
		return err
	}

	err = arcFileObj.Create(out)
	if err != nil {
		return err
	}
	defer func() {
		if err := arcFileObj.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	zipFilePathListMap := make(map[string]createArchiveFileInfo)

	err = processFilesForPackingArchives(&zipFilePathListMap, fileList, commonParentPath, &gitIgnorePattern)
	if err != nil {
		return err
	}

	totalFiles := len(zipFilePathListMap)
	pInfo, ch := initProgress(totalFiles, ph)

	count := 0
	for absolutePath, item := range zipFilePathListMap {
		count += 1
		pInfo.progress(ch, totalFiles, absolutePath, count)

		if err := addFileToTarBall(&arcFileObj, *item.fileInfo, item.absFilepath, item.relativeFilePath, item.isDir); err != nil {
			return err
		}
	}

	pInfo.endProgress(ch, totalFiles)

	return err
}

func addFileToTarBall(arcFileObj *interface{ archiver.Writer }, fileInfo os.FileInfo, filename string, relativeFilename string, isDir bool) error {
	_arcFileObj := *arcFileObj

	_relativeFilename := relativeFilename

	if isDir {
		_relativeFilename = strings.TrimRight(_relativeFilename, PathSep)
	}

	fileToArchive, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		if err := fileToArchive.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	err = _arcFileObj.Write(archiver.File{
		FileInfo: archiver.FileInfo{
			FileInfo:   fileInfo,
			CustomName: _relativeFilename,
		},
		ReadCloser: fileToArchive,
	})

	return err
}
