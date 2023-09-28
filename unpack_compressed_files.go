package onearchiver

import (
	"fmt"
	"github.com/ganeshrvel/archiver"
	ignore "github.com/sabhiram/go-gitignore"
	"io"
	"os"
	"path/filepath"
)

func startUnpackingCompressedFiles(session *Session, arc compressedFile, arcFileDecompressor interface{ archiver.DecompressorBare }) error {
	fileList := arc.unpack.FileList
	gitIgnorePattern := arc.meta.GitIgnorePattern
	destinationPath := arc.unpack.Destination
	sourceFilepath := arc.meta.Filename

	var ignoreList []string
	ignoreList = append(ignoreList, GlobalPatternDenylist...)
	ignoreList = append(ignoreList, gitIgnorePattern...)
	ignoreMatches := ignore.CompileIgnoreLines(ignoreList...)

	allowFileFiltering := len(fileList) > 0

	compressedFileList, err := GetArchiveFileList(&arc.meta, &arc.read)
	if err != nil {
		return err
	}

	compressedFilePathListMap := make(map[string]extractCommonArchiveFileInfo)

	arcFileStat, err := os.Lstat(sourceFilepath)
	if err != nil {
		return err
	}

	progressMetrices := newArchiveProgressMetrices[extractCommonArchiveFileInfo]()
	for _, file := range compressedFileList {
		var fileInfo ArchiveFileInfo

		fullPath := filepath.ToSlash(file.Name)
		isDir := file.IsDir

		fileInfo = ArchiveFileInfo{
			Mode:       file.Mode,
			Size:       file.Size,
			IsDir:      isDir,
			ModTime:    sanitizeTime(file.ModTime, arcFileStat.ModTime()),
			Name:       file.Name,
			FullPath:   fullPath,
			ParentPath: GetParentDirectory(fullPath),
		}

		if allowFileFiltering {
			matched := StringFilter(fileList, func(s string) bool {
				_fName := fixDirSlash(fileInfo.IsDir, fileInfo.FullPath)

				return subpathExists(s, _fName)
			})

			if len(matched) < 1 {
				continue
			}
		}

		if ignoreMatches.MatchesPath(fileInfo.FullPath) {
			return nil
		}
		destinationFileAbsPath := filepath.Join(destinationPath, fileInfo.FullPath)

		progressMetrices.updateArchiveProgressMetrices(compressedFilePathListMap, destinationFileAbsPath, fileInfo.Size, fileInfo.IsDir)
		compressedFilePathListMap[destinationFileAbsPath] = extractCommonArchiveFileInfo{
			absFilepath: destinationFileAbsPath,
			name:        fileInfo.Name,
			fileInfo:    &fileInfo,
		}
	}

	session.initializeProgress(progressMetrices.totalFiles, progressMetrices.totalSize)

	for destinationFileAbsPath, file := range compressedFilePathListMap {
		select {
		case <-session.isDone():
			return session.ctxError()
		default:
		}

		progressMetrices.updateArchiveFilesProgressCount(file.fileInfo.IsDir)
		session.enableCtxCancel()
		session.fileProgress(destinationFileAbsPath, progressMetrices.filesProgressCount)

		if err := addFileFromCompressedFileToDisk(session, &arcFileDecompressor, file.fileInfo, destinationFileAbsPath, sourceFilepath); err != nil {
			return err
		}
	}

	session.endProgress()

	if !exists(destinationPath) {
		if err := os.Mkdir(destinationPath, 0755); err != nil {
			return err
		}
	}

	return err
}

func addFileFromCompressedFileToDisk(session *Session, arcFileDecompressor *interface{ archiver.DecompressorBare }, fileInfo *ArchiveFileInfo, destinationFileAbsPath, sourceFilepath string) error {
	if fileInfo.IsDir {
		if err := os.MkdirAll(destinationFileAbsPath, os.ModePerm); err != nil {
			return err
		}

		return nil
	} else {
		_parent := filepath.Dir(destinationFileAbsPath)

		if err := os.MkdirAll(_parent, os.ModePerm); err != nil {
			return err
		}
	}

	writer, err := os.Create(destinationFileAbsPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := writer.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	reader, err := os.Open(sourceFilepath)
	if err != nil {
		return err
	}
	defer func() {
		if err := reader.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	err = (*arcFileDecompressor).DecompressBare(reader, func(r io.Reader) (written int64, err error) {
		numBytesWritten, err := CtxCopy(session.contextHandler.ctx, writer, r, fileInfo.IsDir, func(soFarTransferredSize, lastTransferredSize int64) {
			session.sizeProgress(fileInfo.Size, soFarTransferredSize, lastTransferredSize)
		})
		if err != nil && !(numBytesWritten == fileInfo.Size && err == io.EOF) {
			return numBytesWritten, err
		}

		return numBytesWritten, nil
	})

	return err
}
