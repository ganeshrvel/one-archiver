package onearchiver

import (
	"bytes"
	"fmt"
	"github.com/ganeshrvel/archiver"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func packTarballs(session *Session, arc *commonArchive, arcFileObj interface{ archiver.Writer }, fileList *[]string, commonParentPath string) error {
	destinationFilename := arc.meta.Filename
	gitIgnorePattern := arc.meta.GitIgnorePattern

	out, err := os.Create(destinationFilename)
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

	progressMetrices, err := processFilesForPackingArchives(&zipFilePathListMap, fileList, commonParentPath, &gitIgnorePattern)
	if err != nil {
		return err
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

		if err := addFileToTarBall(&arcFileObj, *item.fileInfo, item.absFilepath, item.relativeFilePath, item.isDir); err != nil {
			return err
		}
	}

	session.endProgress()

	return err
}

// todo add progress intruption ctxcopy
func addFileToTarBall(arcFileObj *interface{ archiver.Writer }, fileInfo os.FileInfo, sourceFilename string, relativeSourceFilename string, isDir bool) error {
	_arcFileObj := *arcFileObj

	_relativeFilename := relativeSourceFilename

	if isDir {
		_relativeFilename = strings.TrimRight(_relativeFilename, PathSep)
	}

	var fileToArchive io.ReadCloser
	if isSymlink(fileInfo) {
		target, err := os.Readlink(sourceFilename)
		if err != nil {
			return err
		}

		targetReader := bytes.NewReader([]byte(filepath.ToSlash(target)))
		fileToArchive = io.NopCloser(targetReader)
	} else {
		f, err := os.Open(sourceFilename)
		if err != nil {
			return err
		}
		fileToArchive = f
	}

	defer func() {
		if err := fileToArchive.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	// todo add a check if continue of error then dont return
	err := _arcFileObj.Write(archiver.File{
		FileInfo: archiver.FileInfo{
			FileInfo:   fileInfo,
			CustomName: _relativeFilename,
			SourcePath: sourceFilename,
		},
		ReadCloser: fileToArchive,
	})

	return err
}
