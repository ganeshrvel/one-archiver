package one_archiver

import (
	"fmt"
	"github.com/yeka/zip"
	"time"
)

func ListArchive() {
	filename := getDesktopFiles("test.zip")

	if exist := FileExists(filename); !exist {
		fmt.Printf("file does not exist: %s\n", filename)

		return
	}

	_metaObj := &ArchiveMeta{
		Filename:         filename,
		Password:         "",
		GitIgnorePattern: []string{},
	}

	_listObj := &ArchiveRead{
		ListDirectoryPath: "test-directory/",
		Recursive:         true,
		OrderBy:           OrderByName,
		OrderDir:          OrderDirAsc,
	}

	result, err := GetArchiveFileList(_metaObj, _listObj)

	if err != nil {
		fmt.Printf("Error occured: %+v\n", err)

		return
	}

	fmt.Printf("Result: %+v\n", result)

}

func IsArchiveEncrypted() {
	filename := getDesktopFiles("test.enc.zip")
	//filename := getDesktopFiles("test.enc.rar")

	if exist := FileExists(filename); !exist {
		fmt.Printf("file does not exist %s\n", filename)

		return
	}

	_metaObj := &ArchiveMeta{
		Filename: filename,
		Password: "1234567",
	}

	result, err := isArchiveEncrypted(_metaObj)

	if err != nil {
		fmt.Printf("Error occured: %+v\n", err)

		return
	}

	fmt.Printf("Result; isEncrypted: %v, isValidPassword: %v\n", result.isEncrypted, result.isValidPassword)
}

func Pack() {
	filename := getDesktopFiles("12345.pack.zip")
	path1 := getDesktopFiles("test")
	path2 := getDesktopFiles("openmtp")

	_metaObj := &ArchiveMeta{
		Filename:         filename,
		GitIgnorePattern: []string{},
		Password:         "",
		EncryptionMethod: zip.StandardEncryption,
	}

	_packObj := &ArchivePack{
		FileList: []string{path1, path2},
	}

	ph := ProgressHandler{
		onReceived: func(pInfo *ProgressInfo) {
			fmt.Printf("received: %v\n", pInfo)
		},
		onError: func(err error, pInfo *ProgressInfo) {
			fmt.Printf("error: %e\n", err)
		},
		onCompleted: func(pInfo *ProgressInfo) {
			elapsed := time.Since(pInfo.startTime)

			fmt.Println("observable is closed")
			fmt.Printf("Time taken to create the archive: %s", elapsed)
		},
	}

	err := startPacking(_metaObj, _packObj, &ph)
	if err != nil {
		fmt.Printf("Error occured: %+v\n", err)

		return
	}

	fmt.Printf("Result: %+v\n", "Success")
}

func Unpack() {
	filename := getTestMocksAsset("mock_test_file1.zip")
	tempDir := newTempMocksDir("arc_test_pack/", false)

	_metaObj := &ArchiveMeta{
		Filename:         filename,
		Password:         "",
		GitIgnorePattern: []string{},
	}

	_unpackObj := &ArchiveUnpack{
		FileList:    []string{},
		Destination: tempDir,
	}

	ph := ProgressHandler{
		onReceived: func(pInfo *ProgressInfo) {
			fmt.Printf("received: %v\n", pInfo)
		},
		onError: func(err error, pInfo *ProgressInfo) {
			fmt.Printf("error: %e\n", err)
		},
		onCompleted: func(pInfo *ProgressInfo) {
			elapsed := time.Since(pInfo.startTime)

			fmt.Println("observable is closed")
			fmt.Printf("Time taken to unpack the archive: %s", elapsed)
		},
	}

	err := startUnpacking(_metaObj, _unpackObj, &ph)
	if err != nil {
		fmt.Printf("Error occured: %+v\n", err)

		return
	}

	fmt.Printf("Result: %+v\n", "Success")
}

func main() {
	//ListArchive()
	//IsArchiveEncrypted()
	//Pack()
	//Unpack()
}
