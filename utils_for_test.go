package onearchiver_test

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	. "github.com/ganeshrvel/one-archiver"
	"github.com/samber/lo"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func getTestMocksAsset(_filePath string) string {
	currentDir, err := os.Getwd()

	if err != nil {
		log.Panicf("\nunable to fetch the current directory: %s\n", currentDir)
	}

	resultPath := fmt.Sprintf("%s/tests/mocks/", currentDir)

	resultPath = fmt.Sprintf("%s%s", resultPath, _filePath)

	if exist := Exists(resultPath); !exist {
		fi, err := os.Lstat(resultPath)
		if err != nil {
			log.Panicf("\ninvalid lstat: %s\n", err)
		}

		if IsSymlink(fi) {
			return resultPath
		}

		log.Panicf("\nthe 'mocks' asset not found: %s\n", resultPath)
	}

	return resultPath
}

func newTestMocksAsset(_filePath string) string {
	currentDir, err := os.Getwd()

	if err != nil {
		log.Panicf("\nunable to fetch the current directory: %s\n", currentDir)
	}

	resultPath := fmt.Sprintf("%s/tests/mocks/", currentDir)

	resultPath = fmt.Sprintf("%s%s", resultPath, _filePath)

	return resultPath
}

func newTempMocksAsset(_filePath string) string {
	currentDir, err := os.Getwd()

	if err != nil {
		log.Panicf("\nunable to fetch the current directory: %s\nerror: %+v\n", currentDir, err)
	}

	resultPath := fmt.Sprintf("%s/tests/mocks-build/", currentDir)

	if exist := IsDirectory(resultPath); !exist {
		_, err := os.Create(resultPath)

		if err != nil {
			log.Panicf("\n'mocks-build' directory not found: %s\nerror: %+v\n", resultPath, err)
		}
	}

	resultPath = fmt.Sprintf("%s%s", resultPath, _filePath)

	return resultPath
}

func newTempMocksDir(_dirPath string, resetDir bool) string {
	currentDir, err := os.Getwd()

	if err != nil {
		log.Panicf("\nunable to fetch the current directory: %s\nerror: %+v\n", currentDir, err)
	}

	resultPath := filepath.Join(currentDir, "tests/mocks-build", _dirPath)

	if resetDir {
		err := os.RemoveAll(resultPath)

		if err != nil {
			log.Panic(err)
		}

		if exist := IsDirectory(resultPath); !exist {
			err = os.MkdirAll(resultPath, os.ModePerm)

			if err != nil {
				log.Panicf("\ntemp mocks directory not found: %s\nerror: %+v\n", resultPath, err)
			}
		}
	}

	if exist := IsDirectory(resultPath); !exist {
		err := os.MkdirAll(resultPath, os.ModePerm)

		if err != nil {
			log.Panicf("\ntemp mocks directory not found: %s\nerror: %+v\n", resultPath, err)
		}
	}

	return resultPath
}

func testAssetDir(_dirPath string, resetDir bool) string {
	currentDir, err := os.Getwd()

	if err != nil {
		log.Panicf("\nunable to fetch the current directory: %s\nerror: %+v\n", currentDir, err)
	}

	if resetDir {
		err := os.RemoveAll(_dirPath)

		if err != nil {
			log.Panic(err)
		}

		if exist := IsDirectory(_dirPath); !exist {
			err = os.MkdirAll(_dirPath, os.ModePerm)

			if err != nil {
				log.Panicf("\ntest asset Directory not found: %s\nerror: %+v\n", _dirPath, err)
			}
		}
	}

	if exist := IsDirectory(_dirPath); !exist {
		err := os.MkdirAll(_dirPath, os.ModePerm)

		if err != nil {
			log.Panicf("\ntest asset Directory not found: %s\nerror: %+v\n", _dirPath, err)
		}
	}

	return _dirPath
}

func listUnpackedDirectory(destination string) (path []string) {
	var filePathList []FilePathListSortInfo

	err := filepath.Walk(destination, func(path string, info os.FileInfo, err error) error {
		if destination == path {
			return nil
		}

		var pathSplitted [2]string

		if !info.IsDir() {
			pathSplitted = [2]string{filepath.Dir(path), filepath.Base(path)}

		} else {
			path = FixDirSlash(true, path)
			_dir := filepath.Dir(path)

			pathSplitted = [2]string{_dir, ""}
		}

		filePathList = append(filePathList, FilePathListSortInfo{
			IsDir:         info.IsDir(),
			FullPath:      path,
			SplittedPaths: pathSplitted,
		})

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	SortBySplittedPath(&filePathList, OrderDirAsc)

	var itemsArr []string

	for _, x := range filePathList {
		_path := strings.Replace(x.FullPath, destination, "", -1)
		_path = strings.TrimLeft(_path, "/")

		itemsArr = append(itemsArr, _path)
	}

	return itemsArr
}

func getContentsUnpackedDirectory(destination string, fileContents *[]map[string][]byte) {
	err := filepath.Walk(destination, func(path string, info os.FileInfo, err error) error {
		if destination == path {
			return nil
		}

		_path := strings.Replace(path, destination, "", -1)
		_path = strings.TrimLeft(_path, "/")

		if IsSymlink(info) {
			return nil
		}

		if !info.IsDir() {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}

			for idx, m := range *fileContents {
				if _, exists := m[_path]; exists {
					(*fileContents)[idx] = map[string][]byte{_path: data}
				}
			}
		} else {
			path = FixDirSlash(true, path)

			for idx, m := range *fileContents {
				if _, exists := m[_path]; exists {
					(*fileContents)[idx] = map[string][]byte{_path: nil}
				}
			}

		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

}

func MatchRegex(input, pattern string) bool {
	match, _ := regexp.MatchString(pattern, input)
	return match
}

func getFilesInDirectory(fullpath string, filter []string) *map[string]os.FileInfo {

	fileInfo := make(map[string]os.FileInfo)

	err := filepath.Walk(fullpath, func(path string, info os.FileInfo, err error) error {

		// dont returned parent directory
		if fullpath == path {
			return nil
		}

		// if filter is present then only the ones in the filter will be returned
		if len(filter) > 0 {
			if lo.Contains(filter, path) {
				fileInfo[path] = info
			}
			return nil
		}

		// else return all fileinfo inside the directory
		fileInfo[path] = info

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return &fileInfo
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
