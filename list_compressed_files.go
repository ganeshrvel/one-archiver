package onearchiver

import (
	"fmt"
	"github.com/ganeshrvel/archiver"
	ignore "github.com/sabhiram/go-gitignore"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// getCompressedFileSize computes the size of the uncompressed data from a compressed source.
// It accepts a reader that provides the compressed data and a decompressor that knows how to decompress it.
// It returns the size of the uncompressed data.
func getCompressedFileSize(reader io.Reader, compressedFileReader *interface{ archiver.DecompressorBare }) (uncompressedFileSize int64, err error) {
	// Create a buffer to read chunks of data from the decompressed source.
	buf := make([]byte, 2*1024*1024) // 2MB

	// Decompress the data by reading it through the given decompressor.
	err = (*compressedFileReader).DecompressBare(reader, func(r io.Reader) (_ int64, err error) {
		// Continuously read from the decompressed data until EOF or an error occurs.
		for {
			n, err := r.Read(buf)
			// Add the number of bytes read to the total uncompressed size.
			uncompressedFileSize += int64(n)
			if err == io.EOF {
				break
			}
			if err != nil {
				return 0, err
			}
		}

		return 0, err
	})

	if err != nil {
		return 0, err
	}

	return uncompressedFileSize, nil
}

func readCompressedFiles(sourceFilepath string, compressedFileReader *interface{ archiver.DecompressorBare }) (fileInfo ArchiveFileInfo, error error) {
	afi := ArchiveFileInfo{}
	compressFileExt := filepath.Ext(sourceFilepath)

	reader, err := os.Open(sourceFilepath)
	if err != nil {
		return afi, err
	}
	defer func() {
		if err := reader.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()
	stat, err := reader.Stat()
	if err != nil {
		return afi, err
	}

	uncompressedFileSize, err := getCompressedFileSize(reader, compressedFileReader)
	if err != nil {
		return afi, err
	}

	strippedFileName := strings.TrimRight(stat.Name(), compressFileExt)
	fileExt := extension(strippedFileName)
	fullPath := filepath.ToSlash(strippedFileName)

	afi.ModTime = stat.ModTime()
	afi.Mode = stat.Mode()
	afi.Size = uncompressedFileSize
	afi.IsDir = false
	afi.Name = strippedFileName
	afi.FullPath = fullPath
	afi.ParentPath = GetParentDirectory(fullPath)
	afi.Extension = fileExt

	return afi, nil
}

// List files in the compressed file
func (arc compressedFile) list() ([]ArchiveFileInfo, error) {
	filename := arc.meta.Filename
	pctx := arc.read.passwordContext()
	listDirectoryPath := arc.read.ListDirectoryPath
	recursive := arc.read.Recursive
	gitIgnorePattern := arc.meta.GitIgnorePattern

	arcFileObj, err := archiver.ByExtension(filename)
	if err != nil {
		return nil, err
	}

	err = archiveFormat(&arcFileObj, pctx, OverwriteExisting)
	var compressedFileReader interface{ archiver.DecompressorBare }

	switch arcFileReader := arcFileObj.(type) {
	case *archiver.Gz:
		compressedFileReader = arcFileReader
	case *archiver.Brotli:
		compressedFileReader = arcFileReader
	case *archiver.Bz2:
		compressedFileReader = arcFileReader
	case *archiver.Lz4:
		compressedFileReader = arcFileReader
	case *archiver.Snappy:
		compressedFileReader = arcFileReader
	case *archiver.Xz:
		compressedFileReader = arcFileReader
	case *archiver.Zstd:
		compressedFileReader = arcFileReader

	default:
		return nil, fmt.Errorf(string(ErrorFormatUnSupported))
	}

	fileInfo, err := readCompressedFiles(filename, &compressedFileReader)
	if err != nil {
		return nil, fmt.Errorf(string(ErrorArchiverList))
	}

	isListDirectoryPathExist := listDirectoryPath == ""
	var filteredPaths []ArchiveFileInfo

	var ignoreList []string
	ignoreList = append(ignoreList, GlobalPatternDenylist...)
	ignoreList = append(ignoreList, gitIgnorePattern...)
	compiledGitIgnoreLines := ignore.CompileIgnoreLines(ignoreList...)

	includeFile := getFilteredFiles(
		fileInfo, listDirectoryPath, recursive,
	)

	if includeFile {
		if !compiledGitIgnoreLines.MatchesPath(fileInfo.FullPath) {
			filteredPaths = append(filteredPaths, fileInfo)
		}
	}

	if !isListDirectoryPathExist && subpathExists(listDirectoryPath, fileInfo.FullPath) {
		isListDirectoryPathExist = true
	}

	if !isListDirectoryPathExist {
		return filteredPaths, fmt.Errorf("%s: %s", string(ErrorNoPathToFilter), listDirectoryPath)
	}

	return filteredPaths, err
}
