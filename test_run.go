package onearchiver

import (
	"fmt"
	zip "github.com/ganeshrvel/yeka_zip"
	"github.com/kr/pretty"
	"log"
)

func ListArchive(filename string) []ArchiveFileInfo {
	//filename := GetDesktopFile("squash-test-assets/huge_file.zip")

	if exist := FileExists(filename); !exist {
		fmt.Printf("file does not exist: %s\n", filename)

		return nil
	}

	am := &ArchiveMeta{
		Filename:         filename,
		GitIgnorePattern: []string{},
	}

	ar := &ArchiveRead{
		ListDirectoryPath: "",
		Recursive:         true,
		OrderBy:           OrderByName,
		OrderDir:          OrderDirAsc,
		Passwords:         []string{"1234567"},
	}

	result, err := GetArchiveFileList(am, ar)

	if err != nil {
		pretty.Println("Error: ", err)

		return nil
	}

	return result

	//pretty.Println(result)
}

func TestPrepareArchive() {
	filename := GetDesktopFile("test.enc.zip")
	//filename := GetDesktopFile("test.enc.rar")

	if exist := FileExists(filename); !exist {
		fmt.Printf("file does not exist %s\n", filename)

		return
	}

	am := &ArchiveMeta{
		Filename: filename,
	}

	result, err := PrepareArchive(am, []string{"1234567"})

	if err != nil {
		fmt.Printf("Error occured: %+v\n", err)

		return
	}

	fmt.Printf("Result; IsPasswordRequired: %v, IsValidPassword: %v, IsSinglePasswordMode: %v\n", result.IsPasswordRequired, result.IsValidPassword, result.IsSinglePasswordMode)
}

func TestPack(filename string, fileList []string, session *Session) {
	am := &ArchiveMeta{
		Filename:         filename,
		GitIgnorePattern: []string{},
	}

	ap := &ArchivePack{
		FileList:            fileList,
		Password:            "",
		ZipEncryptionMethod: zip.StandardEncryption,
	}

	err := StartPacking(am, ap, session)
	if err != nil {
		fmt.Printf("Error occured: %+v\n", err)

		return
	}

	fmt.Printf("Result: %+v\n", "Success")
}

func TestUnpack(filename, tempDir string, session *Session) {
	//filename := getTestMocksAsset("mock_test_file1.zip")
	//tempDir := newTempMocksDir("arc_test_pack/", false)

	am := &ArchiveMeta{
		Filename:         filename,
		GitIgnorePattern: []string{},
	}

	passwords := []string{"12345678", "123456789", "1234567890", "1234567"}

	au := &ArchiveUnpack{
		FileList:    []string{},
		Passwords:   passwords,
		Destination: tempDir,
	}

	err := StartUnpacking(am, au, session)

	if err != nil {
		log.Panic(err)

		return
	}

	fmt.Printf("Result: %+v\n", "Success")
}
