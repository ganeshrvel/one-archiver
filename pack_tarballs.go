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
	progressStreamDebounceTime := arc.pack.ProgressStreamDebounceTime

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

	tarFilePathListMap := make(map[string]createArchiveFileInfo)

	progressMetrices, err := processFilesForPackingArchives(&tarFilePathListMap, fileList, commonParentPath, &gitIgnorePattern)
	if err != nil {
		return err
	}

	session.initializeProgress(progressMetrices.totalFiles, progressMetrices.totalSize, progressStreamDebounceTime, true)

	for absolutePath, item := range tarFilePathListMap {
		select {
		case <-session.isDone():
			session.endProgress(ProgressStatusCancelled)
			return session.ctxError()
		default:
		}

		progressMetrices.updateArchiveFilesProgressCount(item.isDir)
		if err := session.fileProgress(absolutePath, progressMetrices.filesProgressCount, item.isDir, func() error {
			return addFileToTarBall(session, &arcFileObj, *item.fileInfo, item.absFilepath, item.relativeFilePath, item.isDir)
		}); err != nil {
			return err
		}
	}

	session.endProgress(ProgressStatusCompleted)

	return err
}

func addFileToTarBall(session *Session, arcFileObj *interface{ archiver.Writer }, fileInfo os.FileInfo, sourceFilename string, relativeSourceFilename string, isDir bool) error {
	_arcFileObj := *arcFileObj

	_relativeFilename := relativeSourceFilename

	if isDir {
		_relativeFilename = strings.TrimRight(_relativeFilename, PathSep)
	}

	var fileToArchive io.ReadCloser
	if IsSymlink(fileInfo) {
		originalTargetPath, err := os.Readlink(sourceFilename)
		if err != nil {
			return err
		}
		targetPathToWrite := filepath.ToSlash(originalTargetPath)
		targetReader := bytes.NewReader([]byte(targetPathToWrite))
		fileToArchive = io.NopCloser(targetReader)

		session.linkSizeProgress(originalTargetPath, targetPathToWrite)
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

	af := archiver.File{
		FileInfo: archiver.FileInfo{
			FileInfo:   fileInfo,
			CustomName: _relativeFilename,
			SourcePath: sourceFilename,
		},
		ReadCloser: fileToArchive,
	}

	// todo add a check if continue of error then dont return
	err := _arcFileObj.WriteBare(af, func(w io.Writer, f archiver.File) (written int64, err error) {
		return SessionAwareCopy(session, w, f, fileInfo.IsDir(), fileInfo.Size())
	})

	return err
}
