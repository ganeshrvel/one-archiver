package onearchiver

import (
	"archive/tar"
	"fmt"
	"github.com/ganeshrvel/archiver"
	"github.com/nwaples/rardecode"
	ignore "github.com/sabhiram/go-gitignore"
	"io"
	"os"
	"path/filepath"
)

func startUnpackingCommonArchives(session *Session, arc commonArchive, arcWalker interface{ archiver.Walker }) error {
	sourceFilename := arc.meta.Filename
	gitIgnorePattern := arc.meta.GitIgnorePattern
	fileList := arc.unpack.FileList
	destinationPath := arc.unpack.Destination

	arcFileStat, err := os.Lstat(sourceFilename)
	if err != nil {
		return err
	}

	allowFileFiltering := len(fileList) > 0

	var ignoreList []string
	ignoreList = append(ignoreList, GlobalPatternDenylist...)
	ignoreList = append(ignoreList, gitIgnorePattern...)

	ignoreMatches := ignore.CompileIgnoreLines(ignoreList...)
	commonArchiveFilePathListMap := make(map[string]extractCommonArchiveFileInfo)

	progressMetrices := newArchiveProgressMetrices[extractCommonArchiveFileInfo]()

	err = arcWalker.Walk(sourceFilename, func(file archiver.File) error {
		var fileInfo ArchiveFileInfo

		switch fileHeader := file.Header.(type) {
		case *tar.Header:
			fullPath := filepath.ToSlash(fileHeader.Name)
			isDir := file.IsDir()

			fileInfo = ArchiveFileInfo{
				Mode:       file.Mode(),
				Size:       file.Size(),
				IsDir:      isDir,
				ModTime:    sanitizeTime(file.ModTime(), arcFileStat.ModTime()),
				Name:       file.Name(),
				FullPath:   fullPath,
				ParentPath: GetParentDirectory(fullPath),
			}

		case *rardecode.FileHeader:
			isDir := file.IsDir()
			fullPath := fixDirSlash(isDir, filepath.ToSlash(file.Name()))

			fileInfo = ArchiveFileInfo{
				Mode:       file.Mode(),
				Size:       file.Size(),
				IsDir:      isDir,
				ModTime:    sanitizeTime(file.ModTime(), arcFileStat.ModTime()),
				Name:       filepath.Base(fullPath),
				FullPath:   fullPath,
				ParentPath: GetParentDirectory(fullPath),
			}
		}

		if allowFileFiltering {
			matched := StringFilter(fileList, func(s string) bool {
				_fName := fixDirSlash(fileInfo.IsDir, fileInfo.FullPath)

				return subpathExists(s, _fName)
			})

			if len(matched) < 1 {
				return nil
			}
		}

		if ignoreMatches.MatchesPath(fileInfo.FullPath) {
			return nil
		}

		destinationFileAbsPath := filepath.Join(destinationPath, fileInfo.FullPath)

		progressMetrices.updateArchiveProgressMetrices(commonArchiveFilePathListMap, destinationFileAbsPath, fileInfo.Size, fileInfo.IsDir)
		commonArchiveFilePathListMap[destinationFileAbsPath] = extractCommonArchiveFileInfo{
			absFilepath:           destinationFileAbsPath,
			name:                  fileInfo.Name,
			fileInfo:              &fileInfo,
			sourceArchiveFileInfo: &file.FileInfo,
		}

		return nil
	})

	session.initializeProgress(progressMetrices.totalFiles, progressMetrices.totalSize)

	err = arcWalker.Walk(sourceFilename, func(file archiver.File) error {
		select {
		case <-session.isDone():
			return session.ctxError()
		default:
		}

		var fullPath string
		switch fileHeader := file.Header.(type) {
		case *tar.Header:
			fullPath = filepath.ToSlash(fileHeader.Name)

		case *rardecode.FileHeader:
			isDir := file.IsDir()
			fullPath = fixDirSlash(isDir, filepath.ToSlash(file.Name()))
		}

		destinationFileAbsPath := filepath.Join(destinationPath, fullPath)
		arcFileObj, exists := commonArchiveFilePathListMap[destinationFileAbsPath]
		if !exists {
			return nil
		}

		progressMetrices.updateArchiveFilesProgressCount(file.FileInfo.IsDir())
		session.enableCtxCancel()
		session.fileProgress(destinationFileAbsPath, progressMetrices.filesProgressCount)

		if err := addFileFromCommonArchiveToDisk(session, &arcFileObj, &file, destinationFileAbsPath); err != nil {
			return err
		}

		return nil
	})

	session.endProgress()

	if !exists(destinationPath) {
		if err := os.Mkdir(destinationPath, 0755); err != nil {
			return err
		}
	}

	return err
}

func addFileFromCommonArchiveToDisk(session *Session, arcFileObj *extractCommonArchiveFileInfo, file *archiver.File, destinationFileAbsPath string) error {
	_arcFileObj := *arcFileObj
	if _arcFileObj.fileInfo.IsDir {
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

	if isSymlink(*_arcFileObj.sourceArchiveFileInfo) {
		originalTargetPath := ""
		switch fileHeader := file.Header.(type) {
		case *tar.Header:
			originalTargetPath = fileHeader.Linkname
		}

		if originalTargetPath == "" {
			r, err := io.ReadAll(file.ReadCloser)
			if err != nil {
				return err
			}
			defer func() {
				if err := file.ReadCloser.Close(); err != nil {
					fmt.Printf("%v\n", err)
				}
			}()

			originalTargetPath = string(r)
		}

		targetPathToWrite := filepath.ToSlash(originalTargetPath)

		// todo add a check if continue of error then dont return
		err := os.Symlink(targetPathToWrite, _arcFileObj.absFilepath)
		if err != nil {
			return err
		}

		session.symlinkSizeProgress(originalTargetPath, targetPathToWrite)

		// todo add a check if continue of error then dont return
		return nil
	}

	writer, err := os.OpenFile(destinationFileAbsPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, _arcFileObj.fileInfo.Mode)
	if err != nil {
		return err
	}
	defer func() {
		if err := writer.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	defer func() {
		if err := file.ReadCloser.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	// todo add a check if continue of error then dont return
	numBytesWritten, err := CtxCopy(session.contextHandler.ctx, writer, file.ReadCloser, _arcFileObj.fileInfo.IsDir, func(soFarTransferredSize, lastTransferredSize int64) {
		session.sizeProgress(_arcFileObj.fileInfo.Size, soFarTransferredSize, lastTransferredSize)
	})
	if err != nil && !(numBytesWritten == _arcFileObj.fileInfo.Size && err == io.EOF) {
		return err
	}

	return nil
}
