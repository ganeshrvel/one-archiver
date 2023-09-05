package onearchiver

import (
	zip "github.com/ganeshrvel/yeka_zip"
	. "github.com/smartystreets/goconvey/convey"
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"
)

// TODO symlink and hardlink
// todo update tests for hardlinks to read the contents of the file
func _testListingUnpackedArchive(metaObj *ArchiveMeta, unpackObj *ArchiveUnpack, archiveFilesAssertionArr []string, directoryFilesAssertionArr []string, contentsAssertionArr []map[string][]byte, passwords []string) {
	destination := unpackObj.Destination

	Convey("recursive=true | Asc - it should not throw an error", func() {
		_listObj := &ArchiveRead{
			ListDirectoryPath: "",
			Recursive:         true,
			OrderBy:           OrderByFullPath,
			OrderDir:          OrderDirAsc,
			Passwords:         passwords,
		}

		result, err := GetArchiveFileList(metaObj, _listObj)

		So(err, ShouldBeNil)

		var itemsArr []string

		for _, item := range result {
			itemsArr = append(itemsArr, item.FullPath)
		}

		So(itemsArr, ShouldResemble, archiveFilesAssertionArr)
	})

	Convey("Read the extracted directory  - it should not throw an error", func() {
		filesArr := listUnpackedDirectory(destination)

		So(filesArr, ShouldResemble, directoryFilesAssertionArr)

		if len(contentsAssertionArr) > 0 {

			contentsArr := make([]map[string][]byte, len(contentsAssertionArr))

			for idx, m := range contentsAssertionArr {
				for key := range m {
					contentsArr[idx] = map[string][]byte{key: nil}
				}
			}

			getContentsUnpackedDirectory(destination, &contentsArr)

			Convey("Read contents and match them  - it should not throw an error", func() {
				So(contentsArr, ShouldResemble, contentsAssertionArr)
			})
		}
	})
}

func _testArchiveUnpackingInvalidPassword(_metaObj *ArchiveMeta, ph *ProgressHandler) {
	Convey("Incorrect Password - it should throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{},
			Destination: _destination,
			Passwords:   []string{"wrongpassword"},
		}

		err := StartUnpacking(_metaObj, unpackObj, ph)

		So(err, ShouldBeError)
		So(err.Error(), ShouldContainSubstring, "invalid password")
	})

	Convey("Empty Password - it should throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{},
			Destination: _destination,
		}

		err := StartUnpacking(_metaObj, unpackObj, ph)

		So(err, ShouldBeError)
		So(err.Error(), ShouldContainSubstring, "invalid password")
	})

	Convey("Correct Password - it should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{},
			Destination: _destination,
			Passwords:   []string{"1234567"},
		}

		err := StartUnpacking(_metaObj, unpackObj, ph)

		So(err, ShouldBeNil)
	})
}

func _testArchiveUnpackingInvalidPasswordCommonArchivesAndZip(_metaObj *ArchiveMeta, ph *ProgressHandler) {
	Convey("Incorrect Password | common archives - it should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{},
			Destination: _destination,
			Passwords:   []string{"wrong"},
		}

		err := StartUnpacking(_metaObj, unpackObj, ph)

		So(err, ShouldBeNil)
	})
}

func _testUnpackingCommonArchives(metaObj *ArchiveMeta, ph *ProgressHandler, passwords []string) {
	Convey("Warm up test | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{},
			Destination: _destination,
			Passwords:   passwords,
		}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			assertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			contentsAssertionArr := []map[string][]byte{
				{"mock_dir1/": nil},
				{"mock_dir1/a.txt": []byte("abc d efg")},
				{"mock_dir1/1/": nil},
				{"mock_dir1/1/a.txt": []byte("abcdefg\n")},
				{"mock_dir1/2/": nil},
				{"mock_dir1/2/b.txt": []byte("123456")},
				{"mock_dir1/3/": nil},
				{"mock_dir1/3/b.txt": []byte("123456")},
				{"mock_dir1/3/2/": nil},
				{"mock_dir1/3/2/b.txt": []byte("123456")},
			}

			_testListingUnpackedArchive(metaObj, unpackObj, assertionArr, assertionArr, contentsAssertionArr, passwords)
		})
	})

	Convey("invalid 'FileList' item | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"dummy/path"},
			Destination: _destination,
			Passwords:   passwords,
		}

		metaObj.GitIgnorePattern = []string{"a.txt"}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			var assertionArr []string

			filesArr := listUnpackedDirectory(_destination)

			So(filesArr, ShouldResemble, assertionArr)
		})
	})

	Convey("gitIgnore | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{},
			Destination: _destination,
			Passwords:   passwords,
		}

		metaObj.GitIgnorePattern = []string{"a.txt"}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			assertionArr := []string{"mock_dir1/", "mock_dir1/1/", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, assertionArr, assertionArr, []map[string][]byte{}, passwords)
		})
	})

	Convey("fileList | 1 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"mock_dir1/"},
			Destination: _destination,
			Passwords:   passwords,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}
			directoryFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, directoryFilesAssertionArr, []map[string][]byte{}, passwords)
		})
	})

	Convey("fileList | 2 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"mock_dir1/2/"},
			Destination: _destination,
			Passwords:   passwords,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}
			directoryFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/2/", "mock_dir1/2/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, directoryFilesAssertionArr, []map[string][]byte{}, passwords)
		})
	})

	Convey("fileList | 3 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"mock_dir1/2/b.txt"},
			Destination: _destination,
			Passwords:   passwords,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}
			directoryFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/2/", "mock_dir1/2/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, directoryFilesAssertionArr, []map[string][]byte{}, passwords)
		})
	})

	Convey("fileList | 4 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"mock_dir1/a.txt"},
			Destination: _destination,
			Passwords:   passwords,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}
			directoryFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, directoryFilesAssertionArr, []map[string][]byte{}, passwords)
		})
	})

	Convey("fileList | 5 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"mock_dir1", "mock_dir1/a.txt"},
			Destination: _destination,
			Passwords:   passwords,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, archiveFilesAssertionArr, []map[string][]byte{}, passwords)
		})
	})

	Convey("fileList | 6 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"mock_dir1/2/b.txt", "mock_dir1/a.txt"},
			Destination: _destination,
			Passwords:   passwords,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			directoryFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, directoryFilesAssertionArr, []map[string][]byte{}, passwords)
		})
	})

	Convey("fileList | 7 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"mock_dir1/2/b.txt", "mock_dir1/3/"},
			Destination: _destination,
			Passwords:   passwords,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			directoryFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, directoryFilesAssertionArr, []map[string][]byte{}, passwords)
		})
	})

	Convey("fileList | 8 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"mock_dir1/2", "mock_dir1/3"},
			Destination: _destination,
			Passwords:   passwords,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			directoryFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, directoryFilesAssertionArr, []map[string][]byte{}, passwords)
		})
	})

	Convey("fileList | 9 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"mock_dir1/2/b.txt", "mock_dir1/3/b.txt"},
			Destination: _destination,
			Passwords:   passwords,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			directoryFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, directoryFilesAssertionArr, []map[string][]byte{}, passwords)
		})
	})

	Convey("gitIgnorePattern | fileList | 1 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"mock_dir1/2/b.txt", "mock_dir1/3/b.txt"},
			Destination: _destination,
			Passwords:   passwords,
		}

		metaObj.GitIgnorePattern = []string{"mock_dir1/3"}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt"}

			directoryFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/2/", "mock_dir1/2/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, directoryFilesAssertionArr, []map[string][]byte{}, passwords)
		})
	})
}

func _testListingUnpackedCompressedFiles(metaObj *ArchiveMeta, unpackObj *ArchiveUnpack, archiveFilesAssertionArr []string, directoryFilesAssertionArr []string, contentsAssertionArr []map[string][]byte, passwords []string) {
	destination := unpackObj.Destination

	Convey("Asc - it should not throw an error", func() {
		_listObj := &ArchiveRead{
			ListDirectoryPath: "",
			Recursive:         false,
			OrderBy:           OrderByFullPath,
			OrderDir:          OrderDirAsc,
		}

		result, err := GetArchiveFileList(metaObj, _listObj)

		So(err, ShouldBeNil)

		var itemsArr []string

		for _, item := range result {
			itemsArr = append(itemsArr, item.FullPath)
		}

		So(itemsArr, ShouldResemble, archiveFilesAssertionArr)
	})

	Convey("Read the extracted directory  - it should not throw an error", func() {
		filesArr := listUnpackedDirectory(destination)

		So(filesArr, ShouldResemble, directoryFilesAssertionArr)

		if len(contentsAssertionArr) > 0 {
			contentsArr := make([]map[string][]byte, len(contentsAssertionArr))

			for idx, m := range contentsAssertionArr {
				for key := range m {
					contentsArr[idx] = map[string][]byte{key: nil}
				}
			}

			getContentsUnpackedDirectory(destination, &contentsArr)

			Convey("Read contents and match them  - it should not throw an error", func() {
				So(contentsArr, ShouldResemble, contentsAssertionArr)
			})
		}
	})
}

func _testUnpackingCompressedFiles(metaObj *ArchiveMeta, ph *ProgressHandler, destinationFilename string, passwords []string) {

	Convey("Warm up test | It should not throw an error", func() {
		_destination := newTempMocksDir(filepath.Join("mock_test_file1", filepath.Base(metaObj.Filename)), true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{},
			Destination: _destination,
			Passwords:   passwords,
		}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			assertionArr := []string{destinationFilename}

			contentsAssertionArr := []map[string][]byte{
				{destinationFilename: []byte("abc d efg")},
			}

			_testListingUnpackedCompressedFiles(metaObj, unpackObj, assertionArr, assertionArr, contentsAssertionArr, passwords)

		})

	})

	Convey("invalid 'FileList' item | It should not throw an error", func() {
		_destination := newTempMocksDir(filepath.Join("mock_test_file1", filepath.Base(metaObj.Filename)), true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"dummy/path"},
			Destination: _destination,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			var assertionArr []string

			filesArr := listUnpackedDirectory(destinationFilename)

			So(filesArr, ShouldResemble, assertionArr)
		})
	})

	Convey("gitIgnore | It should not throw an error", func() {
		_destination := newTempMocksDir(filepath.Join("mock_test_file1", filepath.Base(metaObj.Filename)), true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{},
			Destination: _destination,
		}

		metaObj.GitIgnorePattern = []string{destinationFilename}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			var assertionArr []string

			_testListingUnpackedCompressedFiles(metaObj, unpackObj, assertionArr, assertionArr, []map[string][]byte{}, passwords)
		})
	})

	Convey("gitIgnore 2 | It should not throw an error", func() {
		_destination := newTempMocksDir(filepath.Join("mock_test_file1", filepath.Base(metaObj.Filename)), true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{destinationFilename},
			Destination: _destination,
		}

		metaObj.GitIgnorePattern = []string{destinationFilename}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			var assertionArr []string

			_testListingUnpackedCompressedFiles(metaObj, unpackObj, assertionArr, assertionArr, []map[string][]byte{}, passwords)
		})
	})

	Convey("fileList | 3 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"b.txt"},
			Destination: _destination,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{destinationFilename}
			var directoryFilesAssertionArr []string

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, directoryFilesAssertionArr, []map[string][]byte{}, passwords)
		})
	})

	Convey("fileList | 4 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{destinationFilename},
			Destination: _destination,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{destinationFilename}
			directoryFilesAssertionArr := []string{destinationFilename}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, directoryFilesAssertionArr, []map[string][]byte{}, passwords)
		})
	})
}

func TestUnpacking(t *testing.T) {
	//if testing.Short() {
	//	t.Skip("skipping 'TestUnpacking' testing in short mode")
	//}

	ph := ProgressHandler{
		OnReceived: func(pInfo *ProgressInfo) {
			//fmt.Printf("received: %v\n", pInfo)
		},
		OnError: func(err error, pInfo *ProgressInfo) {
			//fmt.Printf("error: %e\n", err)
		},
		OnCompleted: func(pInfo *ProgressInfo) {
			//elapsed := time.Since(pInfo.StartTime)
			//
			//fmt.Println("observable is closed")
			//fmt.Printf("Time taken to unpack the archive: %s", elapsed)
		},
	}

	Convey("Unpacking | No encryption - ZIP", t, func() {
		filename := getTestMocksAsset("mock_test_file1.zip")

		metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testUnpackingCommonArchives(metaObj, &ph, passwords)
	})

	Convey("Unpacking | Encryption - ZIP", t, func() {
		filename := getTestMocksAsset("mock_enc_test_file1.zip")

		metaObj := &ArchiveMeta{Filename: filename}

		passwords := []string{"1234567"}
		_testUnpackingCommonArchives(metaObj, &ph, passwords)
	})

	Convey("Unpacking | windows Encryption legacy - ZIP", t, func() {
		filename := getTestMocksAsset("windows_mocks/mock_dir1_enc_legacy.zip")

		metaObj := &ArchiveMeta{Filename: filename}

		passwords := []string{"1234567"}
		_testUnpackingCommonArchives(metaObj, &ph, passwords)
	})

	Convey("Unpacking | Multiple password Encryption - ZIP", t, func() {
		filename := getTestMocksAsset("mock_enc_multiple_password_test_file1.zip")

		metaObj := &ArchiveMeta{Filename: filename}

		passwords := []string{"1234567", "12345678", "123456789", "1234567890"}
		_testUnpackingCommonArchives(metaObj, &ph, passwords)
	})

	Convey("Unpacking | rar", t, func() {
		filename := getTestMocksAsset("mock_test_file1.rar")

		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testUnpackingCommonArchives(_metaObj, &ph, passwords)
	})

	Convey("Unpacking | windows encrypted rar", t, func() {
		filename := getTestMocksAsset("windows_mocks/mock_dir1_best_enc.rar")

		_metaObj := &ArchiveMeta{Filename: filename}

		passwords := []string{"1234567"}
		_testUnpackingCommonArchives(_metaObj, &ph, passwords)
	})

	Convey("Unpacking | windows file names encrypted rar", t, func() {
		filename := getTestMocksAsset("windows_mocks/mock_dir1_enc_file_names_encrypted.rar")

		_metaObj := &ArchiveMeta{Filename: filename}

		passwords := []string{"1234567"}
		_testUnpackingCommonArchives(_metaObj, &ph, passwords)
	})

	Convey("Unpacking | windows solid rar", t, func() {
		filename := getTestMocksAsset("windows_mocks/mock_dir1_solid_archive_lock.rar")

		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testUnpackingCommonArchives(_metaObj, &ph, passwords)
	})

	Convey("Unpacking | windows dict best rar", t, func() {
		filename := getTestMocksAsset("windows_mocks/mock_dir1_1024mb_dict_best.rar")

		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testUnpackingCommonArchives(_metaObj, &ph, passwords)
	})

	Convey("Unpacking | Tar", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar")

		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testUnpackingCommonArchives(_metaObj, &ph, passwords)
	})

	Convey("Unpacking | Tar.gz", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.gz")

		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testUnpackingCommonArchives(_metaObj, &ph, passwords)
	})

	Convey("Unpacking | Tar.bz2", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.bz2")

		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testUnpackingCommonArchives(_metaObj, &ph, passwords)
	})

	Convey("Unpacking | Tar.br (brotli)", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.br")

		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testUnpackingCommonArchives(_metaObj, &ph, passwords)
	})

	Convey("Unpacking | Tar.lz4", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.lz4")

		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testUnpackingCommonArchives(_metaObj, &ph, passwords)
	})

	Convey("Unpacking | Tar.sz", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.sz")

		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testUnpackingCommonArchives(_metaObj, &ph, passwords)
	})

	Convey("Unpacking | Tar.xz", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.xz")

		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testUnpackingCommonArchives(_metaObj, &ph, passwords)
	})

	Convey("Unpacking | Tar.zst (zstd)", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.zst")

		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testUnpackingCommonArchives(_metaObj, &ph, passwords)
	})
	Convey("Unpacking compressed file | GZ", t, func() {
		filename := getTestMocksAsset("mock_test_file1.zst")

		metaObj := &ArchiveMeta{Filename: filename}

		_testUnpackingCompressedFiles(metaObj, &ph, "mock_test_file1", []string{})
	})
	Convey("Unpacking compressed file | GZ", t, func() {
		filename := getTestMocksAsset("mock_test_file1.a.txt.gz")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testUnpackingCompressedFiles(_metaObj, &ph, "mock_test_file1.a.txt", []string{})
	})

	Convey("Unpacking compressed file | Zstd", t, func() {
		filename := getTestMocksAsset("mock_test_file1.zst")
		_metaObj := &ArchiveMeta{Filename: filename}
		_testUnpackingCompressedFiles(_metaObj, &ph, "mock_test_file1", []string{})
	})

	Convey("Unpacking compressed file | Xz", t, func() {
		filename := getTestMocksAsset("mock_test_file1.xz")
		_metaObj := &ArchiveMeta{Filename: filename}
		_testUnpackingCompressedFiles(_metaObj, &ph, "mock_test_file1", []string{})
	})

	Convey("Unpacking compressed file | sz (Snappy)", t, func() {
		filename := getTestMocksAsset("mock_test_file1.sz")
		_metaObj := &ArchiveMeta{Filename: filename}
		_testUnpackingCompressedFiles(_metaObj, &ph, "mock_test_file1", []string{})
	})

	Convey("Unpacking compressed file | Lz4", t, func() {
		filename := getTestMocksAsset("mock_test_file1.lz4")
		_metaObj := &ArchiveMeta{Filename: filename}
		_testUnpackingCompressedFiles(_metaObj, &ph, "mock_test_file1", []string{})
	})

	Convey("Unpacking compressed file | Bz2", t, func() {
		filename := getTestMocksAsset("mock_test_file1.bz2")
		_metaObj := &ArchiveMeta{Filename: filename}
		_testUnpackingCompressedFiles(_metaObj, &ph, "mock_test_file1", []string{})
	})

	Convey("Unpacking compressed file | BR (Brotli)", t, func() {
		filename := getTestMocksAsset("mock_test_file1.br")
		_metaObj := &ArchiveMeta{Filename: filename}
		_testUnpackingCompressedFiles(_metaObj, &ph, "mock_test_file1", []string{})
	})
}

func TestArchiveUnpackingPassword(t *testing.T) {
	ph := ProgressHandler{
		OnReceived: func(pInfo *ProgressInfo) {
			//fmt.Printf("received: %v\n", pInfo)
		},
		OnError: func(err error, pInfo *ProgressInfo) {
			//fmt.Printf("error: %e\n", err)
		},
		OnCompleted: func(pInfo *ProgressInfo) {
			//elapsed := time.Since(pInfo.StartTime)
			//
			//fmt.Println("observable is closed")
			//fmt.Printf("Time taken to unpack the archive: %s", elapsed)
		},
	}

	Convey("Wrong password | Archive Unpacking - ZIP", t, func() {
		filename := getTestMocksAsset("mock_enc_test_file1.zip")
		_metaObj := &ArchiveMeta{Filename: filename}

		_testArchiveUnpackingInvalidPasswordCommonArchivesAndZip(_metaObj, &ph)
	})

	Convey("Wrong password | Archive Unpacking - RAR", t, func() {
		filename := getTestMocksAsset("mock_enc_test_file1.rar")
		_metaObj := &ArchiveMeta{Filename: filename}

		_testArchiveUnpackingInvalidPassword(_metaObj, &ph)
	})

	Convey("Wrong password | Archive Unpacking - Common Archives", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar")
		_metaObj := &ArchiveMeta{Filename: filename}

		_testArchiveUnpackingInvalidPasswordCommonArchivesAndZip(_metaObj, &ph)
	})
}

func _testUnpackingSymlinkCommonArchives(metaObj *ArchiveMeta, ph *ProgressHandler, passwords []string) {
	Convey("Warm up test CommonArchives | symlink | It should not throw an error", func() {
		_destination := newTempMocksDir("symlink_mock_test_file5", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{},
			Destination: _destination,
			Passwords:   passwords,
		}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			assertionArr := []string{"a.txt", "b.txt", "cc.txt", "1/", "1/a.txt", "2/", "2/b.txt", "3/", "3/b.txt", "3/2/", "3/2/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, assertionArr, assertionArr, []map[string][]byte{}, passwords)
		})

		Convey("read the unpacked directory and confirm the symlink property", func() {
			filesArr := listUnpackedDirectory(_destination)
			for _, v := range filesArr {
				if v != "cc.txt" && v != "b.txt" {
					continue
				}

				symlinkPath := path.Join(_destination, v)
				lstat, err := os.Lstat(symlinkPath)
				if err != nil {
					log.Panicf("%v\n", err)
				}
				So(isSymlink(lstat), ShouldBeTrue)

				target, err := os.Readlink(symlinkPath)

				if err != nil {
					log.Panicf("%v\n", err)
				}

				if v == "cc.txt" {
					So(target, ShouldResemble, "notfound.txt")
				} else if v == "b.txt" {
					So(target, ShouldResemble, "a.txt")
				}
			}

		})
	})

}

func _testUnpackingSymlinkCompressedFile(metaObj *ArchiveMeta, ph *ProgressHandler, destinationFilename string, passwords []string) {
	Convey("Warm up test compressed file | symlink | It should not throw an error", func() {
		_destination := newTempMocksDir("symlink_mock_test_file5_compressed", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{},
			Destination: _destination,
		}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			assertionArr := []string{destinationFilename}

			_testListingUnpackedArchive(metaObj, unpackObj, assertionArr, assertionArr, []map[string][]byte{}, passwords)
		})

		Convey("read the unpacked directory and read the hardlinked symlink file", func() {
			filesArr := listUnpackedDirectory(_destination)
			for _, v := range filesArr {
				if v != destinationFilename {
					continue
				}

				text, err := os.ReadFile(path.Join(_destination, v))

				if err != nil {
					log.Panicf("%v\n", err)
				}

				So(string(text), ShouldResemble, "abc d efg")
			}

		})

	})

}

func TestSymlinkUnpacking(t *testing.T) {
	//if testing.Short() {
	//	t.Skip("skipping 'TestPacking' testing in short mode")
	//}

	ph := ProgressHandler{
		OnReceived: func(pInfo *ProgressInfo) {
			//fmt.Printf("received: %v\n", pInfo)
		},
		OnError: func(err error, pInfo *ProgressInfo) {
			//fmt.Printf("error: %e\n", err)
		},
		OnCompleted: func(pInfo *ProgressInfo) {
			//elapsed := time.Since(pInfo.StartTime)
			//
			//fmt.Println("observable is closed")
			//fmt.Printf("Time taken to create the archive: %s", elapsed)
		},
	}

	Convey("Unpacking | No encryption - ZIP", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.zip")

		_metaObj := &ArchiveMeta{
			Filename: filename,
		}

		_testUnpackingSymlinkCommonArchives(_metaObj, &ph, []string{})
	})

	Convey("Unpacking | Encrypted - ZIP (StandardEncryption)", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_stdenc_pack.zip")

		_metaObj := &ArchiveMeta{
			Filename:         filename,
			EncryptionMethod: zip.StandardEncryption,
		}

		passwords := []string{"1234567"}
		_testUnpackingSymlinkCommonArchives(_metaObj, &ph, passwords)
	})

	Convey("Unpacking | Encrypted - ZIP (AES128Encryption)", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_aes128enc_pack.zip")

		_metaObj := &ArchiveMeta{
			Filename:         filename,
			EncryptionMethod: zip.AES128Encryption,
		}

		passwords := []string{"1234567"}
		_testUnpackingSymlinkCommonArchives(_metaObj, &ph, passwords)
	})

	Convey("Unpacking | Encrypted - ZIP (AES256Encryption)", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_aes256enc_pack.zip")

		_metaObj := &ArchiveMeta{
			Filename:         filename,
			EncryptionMethod: zip.AES256Encryption,
		}

		passwords := []string{"1234567"}
		_testUnpackingSymlinkCommonArchives(_metaObj, &ph, passwords)
	})

	Convey("Unpacking | Encrypted - ZIP (AES192Encryption)", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_aes192enc_pack.zip")

		_metaObj := &ArchiveMeta{
			Filename:         filename,
			EncryptionMethod: zip.AES192Encryption,
		}

		passwords := []string{"1234567"}
		_testUnpackingSymlinkCommonArchives(_metaObj, &ph, passwords)
	})

	Convey("Unpacking | Tar", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.tar")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testUnpackingSymlinkCommonArchives(_metaObj, &ph, []string{})
	})

	Convey("Unpacking | Tar.gz", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.tar.gz")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testUnpackingSymlinkCommonArchives(_metaObj, &ph, []string{})
	})

	Convey("Unpacking | Tar.bz2", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.tar.bz2")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testUnpackingSymlinkCommonArchives(_metaObj, &ph, []string{})
	})

	Convey("Unpacking | Tar.br (brotli)", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.tar.br")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testUnpackingSymlinkCommonArchives(_metaObj, &ph, []string{})
	})

	Convey("Unpacking | Tar.lz4", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.tar.lz4")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testUnpackingSymlinkCommonArchives(_metaObj, &ph, []string{})
	})

	Convey("Unpacking | Tar.sz", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.tar.sz")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testUnpackingSymlinkCommonArchives(_metaObj, &ph, []string{})
	})

	Convey("Unpacking | Tar.xz", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.tar.xz")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testUnpackingSymlinkCommonArchives(_metaObj, &ph, []string{})
	})

	Convey("Unpacking | Tar.zst (zstd)", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.tar.zst")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testUnpackingSymlinkCommonArchives(_metaObj, &ph, []string{})
	})

	Convey("Unpacking compressed file | GZ", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.gz")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testUnpackingSymlinkCompressedFile(_metaObj, &ph, "arc_test_pack", []string{})
	})

	Convey("Unpacking compressed file | Zstd", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.zst")
		_metaObj := &ArchiveMeta{Filename: filename}
		_testUnpackingSymlinkCompressedFile(_metaObj, &ph, "arc_test_pack", []string{})
	})

	Convey("Unpacking compressed file | Xz", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.xz")
		_metaObj := &ArchiveMeta{Filename: filename}
		_testUnpackingSymlinkCompressedFile(_metaObj, &ph, "arc_test_pack", []string{})
	})

	Convey("Unpacking compressed file | sz (Snappy)", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.sz")
		_metaObj := &ArchiveMeta{Filename: filename}
		_testUnpackingSymlinkCompressedFile(_metaObj, &ph, "arc_test_pack", []string{})
	})

	Convey("Unpacking compressed file | Lz4", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.lz4")
		_metaObj := &ArchiveMeta{Filename: filename}
		_testUnpackingSymlinkCompressedFile(_metaObj, &ph, "arc_test_pack", []string{})
	})

	Convey("Unpacking compressed file | Bz2", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.bz2")
		_metaObj := &ArchiveMeta{Filename: filename}
		_testUnpackingSymlinkCompressedFile(_metaObj, &ph, "arc_test_pack", []string{})
	})

	Convey("Unpacking compressed file | BR (Brotli)", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.br")
		_metaObj := &ArchiveMeta{Filename: filename}
		_testUnpackingSymlinkCompressedFile(_metaObj, &ph, "arc_test_pack", []string{})
	})
}
