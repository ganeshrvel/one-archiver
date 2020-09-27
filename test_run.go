package onearchiver

import (
	"fmt"
	"github.com/yeka/zip"
	"time"
)

func ListArchive() {
	filename := getTestMocksAsset("mock_test_file1.zip")

	if exist := FileExists(filename); !exist {
		fmt.Printf("file does not exist: %s\n", filename)

		return
	}

	am := &ArchiveMeta{
		Filename:         filename,
		Password:         "",
		GitIgnorePattern: []string{},
	}

	ar := &ArchiveRead{
		ListDirectoryPath: "test-directory/",
		Recursive:         true,
		OrderBy:           OrderByName,
		OrderDir:          OrderDirAsc,
	}

	result, err := GetArchiveFileList(am, ar)

	if err != nil {
		fmt.Printf("Error occured: %+v\n", err)

		return
	}

	fmt.Printf("Result: %+v\n", result)
}

func IsEncrypted() {
	filename := getDesktopFiles("test.enc.zip")
	//filename := getDesktopFiles("test.enc.rar")

	if exist := FileExists(filename); !exist {
		fmt.Printf("file does not exist %s\n", filename)

		return
	}

	am := &ArchiveMeta{
		Filename: filename,
		Password: "1234567",
	}

	result, err := IsArchiveEncrypted(am)

	if err != nil {
		fmt.Printf("Error occured: %+v\n", err)

		return
	}

	fmt.Printf("Result; IsEncrypted: %v, IsValidPassword: %v\n", result.IsEncrypted, result.IsValidPassword)
}

func Pack() {
	filename := getDesktopFiles("12345.pack.zip")
	path1 := getDesktopFiles("test")
	path2 := getDesktopFiles("openmtp")

	am := &ArchiveMeta{
		Filename:         filename,
		GitIgnorePattern: []string{},
		Password:         "",
		EncryptionMethod: zip.StandardEncryption,
	}

	ap := &ArchivePack{
		FileList: []string{path1, path2},
	}

	ph := &ProgressHandler{
		OnReceived: func(pInfo *ProgressInfo) {
			fmt.Printf("received: %v\n", pInfo)
		},
		OnError: func(err error, pInfo *ProgressInfo) {
			fmt.Printf("error: %e\n", err)
		},
		OnCompleted: func(pInfo *ProgressInfo) {
			elapsed := time.Since(pInfo.StartTime)

			fmt.Println("observable is closed")
			fmt.Printf("Time taken to create the archive: %s", elapsed)
		},
	}

	err := StartPacking(am, ap, ph)
	if err != nil {
		fmt.Printf("Error occured: %+v\n", err)

		return
	}

	fmt.Printf("Result: %+v\n", "Success")
}

func Unpack() {
	filename := getTestMocksAsset("mock_test_file1.zip")
	tempDir := newTempMocksDir("arc_test_pack/", false)

	am := &ArchiveMeta{
		Filename:         filename,
		Password:         "",
		GitIgnorePattern: []string{},
	}

	au := &ArchiveUnpack{
		FileList:    []string{},
		Destination: tempDir,
	}

	ph := &ProgressHandler{
		OnReceived: func(pInfo *ProgressInfo) {
			fmt.Printf("received: %v\n", pInfo)
		},
		OnError: func(err error, pInfo *ProgressInfo) {
			fmt.Printf("error: %e\n", err)
		},
		OnCompleted: func(pInfo *ProgressInfo) {
			elapsed := time.Since(pInfo.StartTime)

			fmt.Println("observable is closed")
			fmt.Printf("Time taken to unpack the archive: %s", elapsed)
		},
	}

	err := StartUnpacking(am, au, ph)
	if err != nil {
		fmt.Printf("Error occured: %+v\n", err)

		return
	}

	fmt.Printf("Result: %+v\n", "Success")
}

func main() {
	// ListArchive()
	//IsArchiveEncrypted()
	//Pack()
	//Unpack()

	// TODO Cancel a task
	// TODO symlink and hardlink
	// TODO add to archive
	// TODO delete from archive
	// TODO common compression and decompression
}
