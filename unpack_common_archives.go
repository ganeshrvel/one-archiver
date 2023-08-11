package onearchiver

import (
	"archive/tar"
	"github.com/ganeshrvel/archiver"
	"github.com/nwaples/rardecode"
	ignore "github.com/sabhiram/go-gitignore"
	"io"
	"os"
	"path/filepath"
)

func startUnpackingCommonArchives(arc commonArchive, arcWalker interface{ archiver.Walker }, ph *ProgressHandler) error {
	sourceFilename := arc.meta.Filename
	gitIgnorePattern := arc.meta.GitIgnorePattern
	fileList := arc.unpack.FileList
	destinationPath := arc.unpack.Destination

	allowFileFiltering := len(fileList) > 0

	var ignoreList []string
	ignoreList = append(ignoreList, GlobalPatternDenylist...)
	ignoreList = append(ignoreList, gitIgnorePattern...)

	ignoreMatches := ignore.CompileIgnoreLines(ignoreList...)

	commonArchiveFilePathListMap := make(map[string]extractCommonArchiveFileInfo)

	err := arcWalker.Walk(sourceFilename, func(file archiver.File) error {
		var fileInfo ArchiveFileInfo

		switch fileHeader := file.Header.(type) {
		case *tar.Header:
			fullPath := filepath.ToSlash(fileHeader.Name)
			isDir := file.IsDir()

			fileInfo = ArchiveFileInfo{
				Mode:       file.Mode(),
				Size:       file.Size(),
				IsDir:      isDir,
				ModTime:    file.ModTime(),
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
				ModTime:    file.ModTime(),
				Name:       filepath.Base(fullPath),
				FullPath:   fullPath,
				ParentPath: GetParentDirectory(fullPath),
			}

		// not currently being used
		default:
			fullPath := filepath.ToSlash(file.FileInfo.Name())
			isDir := file.IsDir()

			fileInfo = ArchiveFileInfo{
				Mode:       file.Mode(),
				Size:       file.Size(),
				IsDir:      isDir,
				ModTime:    file.ModTime(),
				Name:       file.Name(),
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

		fileData := make([]byte, file.Size())
		numBytesRead, err := file.Read(fileData)
		if err != nil && !(numBytesRead == int(file.Size()) && err == io.EOF) {
			return err
		}

		commonArchiveFilePathListMap[destinationFileAbsPath] = extractCommonArchiveFileInfo{
			absFilepath: destinationFileAbsPath,
			name:        fileInfo.Name,
			fileInfo:    &fileInfo,
			fileBytes:   &fileData,
		}

		return nil
	})

	totalFiles := len(commonArchiveFilePathListMap)
	pInfo, ch := initProgress(totalFiles, ph)

	count := 0
	for absolutePath, file := range commonArchiveFilePathListMap {
		count += 1
		pInfo.progress(ch, totalFiles, absolutePath, count)

		if err := addFileFromCommonArchiveToDisk(&file, absolutePath); err != nil {
			return err
		}
	}

	pInfo.endProgress(ch, totalFiles)

	if !exists(destinationPath) {
		if err := os.Mkdir(destinationPath, 0755); err != nil {
			return err
		}
	}

	return err
}

func addFileFromCommonArchiveToDisk(file *extractCommonArchiveFileInfo, destinationFileAbsPath string) error {
	if file.fileInfo.IsDir {
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

	return os.WriteFile(file.absFilepath, *file.fileBytes, file.fileInfo.Mode)
}
