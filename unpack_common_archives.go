package onearchiver

import (
	"archive/tar"
	"fmt"
	"github.com/ganeshrvel/archiver"
	"github.com/ganeshrvel/rardecode"
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
	progressStreamDebounceTime := arc.unpack.ProgressStreamDebounceTime

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

			size := file.Size()
			if IsSymlink(file) {

				_, symlink, err := getCommonArchivesTargetSymlinkPath(&file)

				if err != nil {
					return err
				}

				size = int64(len(symlink))
			}

			isDir := file.IsDir()

			fileInfo = ArchiveFileInfo{
				Mode:       file.Mode(),
				Size:       size,
				IsDir:      isDir,
				ModTime:    sanitizeTime(file.ModTime(), arcFileStat.ModTime()),
				Name:       file.Name(),
				FullPath:   fullPath,
				ParentPath: GetParentDirectory(fullPath),
			}

		case *rardecode.FileHeader:
			isDir := file.IsDir()
			fullPath := FixDirSlash(isDir, filepath.ToSlash(file.Name()))
			size := file.Size()

			if IsSymlink(file) {
				_, symlink, err := getCommonArchivesTargetSymlinkPath(&file)
				if err != nil {
					return err
				}
				size = int64(len(symlink))
			}

			fileInfo = ArchiveFileInfo{
				Mode:       file.Mode(),
				Size:       size,
				IsDir:      isDir,
				ModTime:    sanitizeTime(file.ModTime(), arcFileStat.ModTime()),
				Name:       filepath.Base(fullPath),
				FullPath:   fullPath,
				ParentPath: GetParentDirectory(fullPath),
			}

		}

		if allowFileFiltering {
			matched := StringFilter(fileList, func(s string) bool {
				_fName := FixDirSlash(fileInfo.IsDir, fileInfo.FullPath)

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

	session.initializeProgress(progressMetrices.totalFiles, progressMetrices.totalSize, progressStreamDebounceTime, true)

	err = arcWalker.Walk(sourceFilename, func(file archiver.File) error {
		isArcFileSymlink := false
		select {
		case <-session.isDone():
			session.endProgress(ProgressStatusCancelled)
			return session.ctxError()
		default:
		}

		var fullPath string
		switch fileHeader := file.Header.(type) {
		case *tar.Header:
			if IsSymlink(file) {
				isArcFileSymlink = true
			}

			fullPath = filepath.ToSlash(fileHeader.Name)

		case *rardecode.FileHeader:
			if IsSymlink(file) {
				isArcFileSymlink = true
			}

			isDir := file.IsDir()
			fullPath = FixDirSlash(isDir, filepath.ToSlash(file.Name()))
		}

		destinationFileAbsPath := filepath.Join(destinationPath, fullPath)
		arcFileObj, exists := commonArchiveFilePathListMap[destinationFileAbsPath]
		if !exists {
			return nil
		}

		progressMetrices.updateArchiveFilesProgressCount(file.FileInfo.IsDir())
		if err := session.fileProgress(destinationFileAbsPath, progressMetrices.filesProgressCount, file.FileInfo.IsDir(), func() error {
			return addFileFromCommonArchiveToDisk(session, &arcFileObj, &file, destinationFileAbsPath, isArcFileSymlink)
		}); err != nil {
			return err
		}

		return nil
	})

	if !Exists(destinationPath) {
		if err := os.Mkdir(destinationPath, 0755); err != nil {
			return err
		}
	}

	session.endProgress(ProgressStatusCompleted)

	return err
}

func getCommonArchivesTargetSymlinkPath(file *archiver.File) (originalTargetPath, targetPathToWrite string, err error) {
	switch fileHeader := file.Header.(type) {
	case *tar.Header:
		originalTargetPath = fileHeader.Linkname
	case *rardecode.FileHeader:
		originalTargetPath = fileHeader.LinkName
	}

	if originalTargetPath == "" {
		r, err := io.ReadAll(file.ReadCloser)
		if err != nil {
			return "", "", err
		}
		defer func() {
			if err := file.ReadCloser.Close(); err != nil {
				fmt.Printf("%v\n", err)
			}
		}()
		originalTargetPath = string(r)
	}

	targetPathToWrite = filepath.ToSlash(originalTargetPath)

	return originalTargetPath, targetPathToWrite, nil
}

func addFileFromCommonArchiveToDisk(session *Session, arcFileObj *extractCommonArchiveFileInfo, file *archiver.File, destinationFileAbsPath string, isArcFileSymlink bool) error {
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

	if isArcFileSymlink {
		originalTargetPath, targetPathToWrite, err := getCommonArchivesTargetSymlinkPath(file)
		if err != nil {
			return err
		}

		// todo add a check if continue of error then dont return
		err = os.Symlink(targetPathToWrite, _arcFileObj.absFilepath)
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
	_, err = SessionAwareCopy(session, writer, file.ReadCloser, _arcFileObj.fileInfo.IsDir, _arcFileObj.fileInfo.Size)

	return err
}
