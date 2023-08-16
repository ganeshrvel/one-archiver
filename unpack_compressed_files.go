package onearchiver

import (
	"fmt"
	"github.com/ganeshrvel/archiver"
	ignore "github.com/sabhiram/go-gitignore"
	"os"
	"path/filepath"
)

func startUnpackingCompressedFiles(arc compressedFile, arcFileDecompressor interface{ archiver.Decompressor }, ph *ProgressHandler) error {
	fileList := arc.unpack.FileList
	gitIgnorePattern := arc.meta.GitIgnorePattern
	destinationPath := arc.unpack.Destination
	sourceFilepath := arc.meta.Filename

	var ignoreList []string
	ignoreList = append(ignoreList, GlobalPatternDenylist...)
	ignoreList = append(ignoreList, gitIgnorePattern...)
	ignoreMatches := ignore.CompileIgnoreLines(ignoreList...)

	allowFileFiltering := len(fileList) > 0

	fiList, err := GetArchiveFileList(&arc.meta, &arc.read)
	if err != nil {
		return err
	}

	compressedFilePathListMap := make(map[string]extractCommonArchiveFileInfo)

	arcFileStat, err := os.Lstat(sourceFilepath)
	if err != nil {
		return err
	}

	for _, file := range fiList {
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

		compressedFilePathListMap[destinationFileAbsPath] = extractCommonArchiveFileInfo{
			absFilepath: destinationFileAbsPath,
			name:        fileInfo.Name,
			fileInfo:    &fileInfo,
		}
	}

	totalFiles := len(compressedFilePathListMap)
	pInfo, ch := initProgress(totalFiles, ph)

	count := 0
	for destinationFileAbsPath, file := range compressedFilePathListMap {
		count += 1
		pInfo.progress(ch, totalFiles, destinationFileAbsPath, count)

		if err := addFileFromCompressedFileToDisk(&arcFileDecompressor, file.fileInfo, destinationFileAbsPath, sourceFilepath); err != nil {
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

func addFileFromCompressedFileToDisk(arcFileDecompressor *interface{ archiver.Decompressor }, fileInfo *ArchiveFileInfo, destinationFileAbsPath, sourceFilepath string) error {
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

	err = (*arcFileDecompressor).Decompress(reader, writer)
	// todo add a check if continue of error then dont return
	if err != nil {
		return err
	}

	return nil
}
